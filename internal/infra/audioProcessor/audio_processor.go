// Package audioprocessor
package audioprocessor

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ChumbaJ/go-transcriber/internal/job"
)

type AudioProcessor struct{}

func New() *AudioProcessor {
	return &AudioProcessor{}
}

func (a *AudioProcessor) Split(ctx context.Context, r io.Reader) ([][]byte, string, error) {
	totalStarted := time.Now()
	path, format, err := a.prepareAudio(ctx, r)
	if err != nil {
		return nil, "", err
	}
	defer os.Remove(path)

	info, err := os.Stat(path)
	if err != nil {
		return nil, "", fmt.Errorf("stat temp audio: %w", err)
	}

	if info.Size() <= job.MaxChunkSize {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, "", fmt.Errorf("read temp audio: %w", err)
		}
		fmt.Printf("audio split: total=%s chunks=1\n", time.Since(totalStarted))
		return [][]byte{data}, format, nil
	}

	stageStarted := time.Now()
	pauses, duration, err := a.findPauses(ctx, path)
	if err != nil {
		return nil, "", err
	}
	fmt.Printf("audio split: find pauses=%s pauses=%d\n", time.Since(stageStarted), len(pauses))
	if duration <= 0 {
		return nil, "", fmt.Errorf("invalid audio duration: %.3f", duration)
	}
	bitrate := float64(info.Size()) * 8 / duration

	stageStarted = time.Now()
	chunks, err := a.cutChunks(ctx, duration, bitrate, pauses, path, format)
	if err != nil {
		return nil, "", err
	}
	fmt.Printf("audio split: cut chunks=%s chunks=%d total=%s\n", time.Since(stageStarted), len(chunks), time.Since(totalStarted))

	return chunks, format, nil
}
