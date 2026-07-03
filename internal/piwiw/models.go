package piwiw

import "encoding/json"

func mergeForcedParams(body []byte, forced json.RawMessage) ([]byte, error) {
	if len(forced) == 0 {
		return body, nil
	}
	var merged map[string]json.RawMessage
	if err := json.Unmarshal(body, &merged); err != nil {
		return nil, err
	}
	var overrides map[string]json.RawMessage
	if err := json.Unmarshal(forced, &overrides); err != nil {
		return nil, err
	}
	for k, v := range overrides {
		merged[k] = v
	}
	return json.Marshal(merged)
}

type OpenAIAPIChatRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIAPIChatRequest struct {
	Model    *string                       `json:"model,omitempty"`
	Messages []OpenAIAPIChatRequestMessage `json:"messages"`
}

type OllamaAPIChatRequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaAPIChatRequest struct {
	Model    *string                       `json:"model,omitempty"`
	Messages []OllamaAPIChatRequestMessage `json:"messages"`
	Stream   *bool                         `json:"stream,omitempty"`
}

func (ollamaReq *OllamaAPIChatRequest) ToOpenAIAPIChatRequest() *OpenAIAPIChatRequest {
	openAIMessages := make([]OpenAIAPIChatRequestMessage, len(ollamaReq.Messages))
	for i, oMessage := range ollamaReq.Messages {
		openAIMessages[i] = OpenAIAPIChatRequestMessage{
			Role:    oMessage.Role,
			Content: oMessage.Content,
		}
	}

	return &OpenAIAPIChatRequest{
		Model:    ollamaReq.Model,
		Messages: openAIMessages,
	}
}

type OllamaAPIChatResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaAPIChatResponse struct {
	Message OllamaAPIChatResponseMessage `json:"message"`
	Model   string                       `json:"model"`
}

type OpenAIAPIChatResponseMessage struct {
	Role    string  `json:"role"`
	Content *string `json:"content"`
}

type OpenAIAPIChatResponseChoice struct {
	Message OpenAIAPIChatResponseMessage `json:"message"`
	Index   int                          `json:"index"`
}

type OpenAIAPIChatResponse struct {
	Choices []OpenAIAPIChatResponseChoice `json:"choices"`
	Model   string                        `json:"model"`
}

func (openAIResp *OpenAIAPIChatResponse) ToOllamaAPIChatResponse() *OllamaAPIChatResponse {
	if openAIResp == nil || len(openAIResp.Choices) == 0 {
		return &OllamaAPIChatResponse{
			Message: OllamaAPIChatResponseMessage{
				Role:    "assistant",
				Content: cfg.EmptyContentText,
			},
		}
	}

	content := cfg.EmptyContentText
	if openAIResp.Choices[0].Message.Content != nil && *openAIResp.Choices[0].Message.Content != "" {
		content = *openAIResp.Choices[0].Message.Content
	}

	return &OllamaAPIChatResponse{
		Message: OllamaAPIChatResponseMessage{
			Role:    openAIResp.Choices[0].Message.Role,
			Content: content,
		},
		Model: openAIResp.Model,
	}
}
