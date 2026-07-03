package piwiw

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func respondError(w http.ResponseWriter, requestID string, status int, msg string) {
	http.Error(w, msg, status)
	logW(requestID, "Incoming request responded and completed (status %d): %s", status, msg)
}

func logClientContextDone(requestID string, ctx context.Context) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		logW(requestID, "Incoming client timed out")
	} else {
		logW(requestID, "Incoming client disconnected")
	}
}

func handleChatRequest(w http.ResponseWriter, r *http.Request) {
	requestID := newRequestID()
	logI(requestID, "New incoming request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		respondError(w, requestID, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		respondError(w, requestID, http.StatusBadRequest, fmt.Errorf("Failed to read request body: %w", err).Error())
		return
	}

	trace := newTraceSession(requestID)
	trace.writeJSON(requestID, "1_ollama_request.incoming.json", bodyBytes)

	var ollamaReq OllamaAPIChatRequest
	if err := json.Unmarshal(bodyBytes, &ollamaReq); err != nil {
		respondError(w, requestID, http.StatusBadRequest, fmt.Errorf("Invalid JSON: %w", err).Error())
		return
	}

	if len(ollamaReq.Messages) == 0 {
		respondError(w, requestID, http.StatusBadRequest, "Messages are required")
		return
	}

	if ollamaReq.Stream != nil && *ollamaReq.Stream {
		respondError(w, requestID, http.StatusBadRequest, "Streamed requests are not supported")
		return
	}

	openAIReq := ollamaReq.ToOpenAIAPIChatRequest()

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipTLSVerify,
		},
	}

	client := &http.Client{
		Timeout:   time.Duration(cfg.RequestTimeout) * time.Second,
		Transport: transport,
	}

	var respBody []byte
	var lastErr error
	var lastStatus int
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		select {
		case <-r.Context().Done():
			logClientContextDone(requestID, r.Context())
			return
		default:
		}

		reqBody, err := json.Marshal(openAIReq)
		if err != nil {
			respondError(w, requestID, http.StatusInternalServerError, "Failed to marshal request")
			return
		}
		reqBody, err = mergeForcedParams(reqBody, cfg.GetOpenAIAPIChatForcedParams())
		if err != nil {
			respondError(w, requestID, http.StatusInternalServerError, "Failed to apply OPENAI_API_CHAT_FORCED_PARAMS")
			return
		}
		trace.writeJSON(requestID, "2_openai_request.outgoing.json", reqBody)

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, cfg.OpenAIAPIBaseUrl, bytes.NewReader(reqBody))
		if err != nil {
			respondError(w, requestID, http.StatusInternalServerError, "Failed to create request")
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.OpenAIAPIToken))

		logI(requestID, "OpenAI request started (attempt %d/%d)", attempt+1, cfg.MaxRetries+1)
		resp, doErr := client.Do(req)

		if doErr != nil {
			lastErr = doErr
			logW(requestID, "OpenAI request failed (by timeout): %v", doErr)
		} else {
			respBody, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = err
				logW(requestID, "OpenAI request failed (by timeout): %v", err)
			} else if resp.StatusCode >= 500 {
				lastErr = nil
				lastStatus = resp.StatusCode
				trace.writeJSON(requestID, "3_openai_response.incoming.json", respBody)
				logW(requestID, "OpenAI request failed (by response): status %d", resp.StatusCode)
			} else {
				trace.writeJSON(requestID, "3_openai_response.incoming.json", respBody)
				logI(requestID, "OpenAI request completed (status %d)", resp.StatusCode)
				lastErr = nil
				break // Success
			}
		}

		if attempt == cfg.MaxRetries {
			logE(requestID, "Out of retries making OpenAI request")
			if lastErr != nil {
				respondError(w, requestID, http.StatusBadGateway, fmt.Sprintf("Failed to reach OpenAI API: %v", lastErr))
			} else {
				respondError(w, requestID, http.StatusBadGateway, fmt.Sprintf("OpenAI API error: %d", lastStatus))
			}
			return
		}

		logI(requestID, "Delay before repeat OpenAI request (%ds)", cfg.RetryDelay)
		select {
		case <-r.Context().Done():
			logClientContextDone(requestID, r.Context())
			return
		case <-time.After(time.Duration(cfg.RetryDelay) * time.Second):
		}
	}

	var openAIResp OpenAIAPIChatResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		respondError(w, requestID, http.StatusInternalServerError, "Failed to parse OpenAI response")
		return
	}
	ollamaResp := openAIResp.ToOllamaAPIChatResponse()

	ollamaRespBytes, err := json.Marshal(ollamaResp)
	if err != nil {
		respondError(w, requestID, http.StatusInternalServerError, "Failed to marshal response")
		return
	}
	trace.writeJSON(requestID, "4_ollama_response.outgoing.json", ollamaRespBytes)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ollamaRespBytes)
	logI(requestID, "Incoming request responded and completed (status 200)")
}
