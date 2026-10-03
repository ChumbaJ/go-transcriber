// Package transcription coordinates audio transcription.
package transcription

import (
	"context"
	"errors"
	"fmt"
)

type Client interface {
	Transcribe(ctx context.Context, audio []byte, format string) (string, error)
}

type Transcriber struct {
	client Client
}

var ErrServiceUnavaliable = errors.New("external service unavaliable")

func New(client Client) *Transcriber {
	return &Transcriber{client: client}
}

func (t *Transcriber) Transcribe(ctx context.Context, audio []byte, format string) (string, error) {
	text, err := t.client.Transcribe(ctx, audio, format)
	if err != nil {
		return "", fmt.Errorf("create transcription: %w", err)
	}
	return text, nil
}
