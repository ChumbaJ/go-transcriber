// Package openai
package openai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/ChumbaJ/go-transcriber/internal/config"
	"github.com/ChumbaJ/go-transcriber/internal/infra/transcription"
	"github.com/openai/openai-go/v3"
)

type OpenAiClient struct {
	client *openai.Client
}

func New(cfg *config.Config) *OpenAiClient {
	client := openai.NewClient()

	return &OpenAiClient{
		client: &client,
	}
}

func (c *OpenAiClient) Transcribe(ctx context.Context, b []byte, format string) (string, error) {
	var filename, contentType string
	switch format {
	case "mp3":
		filename, contentType = "chunk.mp3", "audio/mpeg"
	case "m4a":
		filename, contentType = "chunk.m4a", "audio/mp4"
	case "ogg":
		filename, contentType = "chunk.ogg", "audio/ogg"
	default:
		return "", fmt.Errorf("unsupported audio format: %q", format)
	}

	res, err := c.client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:  openai.File(bytes.NewReader(b), filename, contentType),
		Model: "gpt-transcribe",
		// Language:       param.Opt[string]{Value: "RU"},
		ResponseFormat: "json",
	})
	if err != nil {
		return "", fmt.Errorf("openai audio transcribe: %w", classifyError(ctx, err))
	}

	return res.Text, nil
}

func classifyError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}

	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode == http.StatusRequestTimeout ||
			apiErr.StatusCode == http.StatusConflict ||
			apiErr.StatusCode >= http.StatusInternalServerError ||
			(apiErr.StatusCode == http.StatusTooManyRequests && !isQuotaError(apiErr.Code)) {
			return fmt.Errorf("%w: %w", transcription.ErrServiceUnavaliable, err)
		}
		return err
	}

	var urlErr *url.Error
	var netErr net.Error
	if errors.As(err, &urlErr) || errors.As(err, &netErr) {
		return fmt.Errorf("%w: %w", transcription.ErrServiceUnavaliable, err)
	}

	return err
}

func isQuotaError(code string) bool {
	switch code {
	case "credit_balance_exhausted",
		"organization_spend_limit_exceeded",
		"project_spend_limit_exceeded",
		"organization_usage_limit_exceeded":
		return true
	default:
		return false
	}
}
