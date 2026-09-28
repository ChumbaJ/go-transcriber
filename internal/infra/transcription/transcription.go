// Package transcription is an external service wrapper

package transcription

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/config"
)

type Transcriber struct {
	mu            sync.Mutex
	fakeFailCount int
	httpClient    *http.Client
	baseURL       string
	apiKey        string
}

var ErrServiceUnavaliable = errors.New("external service unavaliable")

const timeoutSec = 10

func New(cfg *config.Config) *Transcriber {
	c := &http.Client{
		Timeout: time.Second * timeoutSec,
	}

	return &Transcriber{
		fakeFailCount: 2,
		httpClient:    c,
		baseURL:       cfg.TranscriberUrl,
		apiKey:        cfg.TranscriberApiKey,
	}
}

// Transcribe currently uses a fake response; real HTTP integration is pending.
func (t *Transcriber) Transcribe(ctx context.Context, b []byte) (string, error) {
	return t.fakeTranscribe(ctx)
}

func (t *Transcriber) fakeTranscribe(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	t.mu.Lock()
	fail := t.fakeFailCount > 0
	if fail {
		t.fakeFailCount--
	}
	t.mu.Unlock()
	if fail {
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timer.C:
		}
		return "", fmt.Errorf("request: %w", ErrServiceUnavaliable)
	}
	return "hello world", nil
}
