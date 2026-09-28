package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/ChumbaJ/go-transcriber/internal/infra/transcription"
)

type retryTranscriber struct {
	calls int
	err   error
}

func (f *retryTranscriber) Transcribe(context.Context, []byte) (string, error) {
	f.calls++
	return "", f.err
}

func TestTranscribeRetryExhaustion(t *testing.T) {
	f := &retryTranscriber{err: transcription.ErrServiceUnavaliable}
	wp := &WorkerPool{transcriberClient: f}
	_, err := wp.transcribeWithRetry(context.Background(), nil)
	if f.calls != 3 || !errors.Is(err, transcription.ErrServiceUnavaliable) {
		t.Fatalf("calls=%d err=%v", f.calls, err)
	}
}

func TestTranscribeUnexpectedErrorIsNotRetried(t *testing.T) {
	expected := errors.New("invalid response")
	f := &retryTranscriber{err: expected}
	wp := &WorkerPool{transcriberClient: f}
	_, err := wp.transcribeWithRetry(context.Background(), nil)
	if f.calls != 1 || !errors.Is(err, expected) {
		t.Fatalf("calls=%d err=%v", f.calls, err)
	}
}

func TestTranscribeCancelledDoesNotCallClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := &retryTranscriber{}
	wp := &WorkerPool{transcriberClient: f}
	_, err := wp.transcribeWithRetry(ctx, nil)
	if f.calls != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("calls=%d err=%v", f.calls, err)
	}
}
