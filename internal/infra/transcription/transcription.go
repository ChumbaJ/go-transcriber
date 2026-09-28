// Package transcription is an external service wrapper

package transcription

import (
	"context"
	"errors"
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

func (t *Transcriber) Transcribe(ctx context.Context, b []byte) (result string, error error) {
	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, nil)
	if err != nil {
		return "", fmt.Errorf("new req: %w", err)
	}

	_, err = t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("response: %w", err)
	}

	return t.fakeTranscribe(ctx)
}

var fakeFailCount = 2

func (*Transcriber) fakeTranscribe(ctx context.Context) (result string, error error) {
	if fakeFailCount > 0 {
		time.Sleep(5)
		fakeFailCount -= 1
		fakeErr := errors.New("service is not avaliable now")
		return "", fmt.Errorf("request: %w", fakeErr)
	}
	return "hello world", nil
}
