// Package transcription is an external service wrapper

package transcription

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/config"
)

type Transcriber struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

const timeoutSec = 10

func New(cfg *config.Config) *Transcriber {
	c := &http.Client{
		Timeout: time.Second * timeoutSec,
	}

	return &Transcriber{
		httpClient: c,
		baseURL:    cfg.TranscriberUrl,
		apiKey:     cfg.TranscriberApiKey,
	}
}

func (t *Transcriber) Transcribe(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, nil)
	if err != nil {
		return fmt.Errorf("new req: %w", err)
	}

	_, err = t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("response: %w", err)
	}

	return nil
}
