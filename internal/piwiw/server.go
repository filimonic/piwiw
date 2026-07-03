package piwiw

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat", handleChatRequest)

	maxOllamaRequestTimeout := time.Duration((cfg.MaxRetries+1)*cfg.RequestTimeout)*time.Second + time.Duration(cfg.MaxRetries)*time.Duration(cfg.RetryDelay)*time.Second

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ProxyPort),
		Handler:      mux,
		ReadTimeout:  time.Minute * 10,
		WriteTimeout: maxOllamaRequestTimeout,
		IdleTimeout:  maxOllamaRequestTimeout,
	}

	stopTraceCleanup := make(chan struct{})
	go startTraceCleanup(stopTraceCleanup)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Printf("Server starting on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-sigCh
	log.Printf("Received signal: %v\n", sig)
	close(stopTraceCleanup)

	ctx, cancel := context.WithTimeout(context.Background(), maxOllamaRequestTimeout)
	defer cancel()
	log.Println("Shutting down server gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}
	log.Println("Server stopped")
}
