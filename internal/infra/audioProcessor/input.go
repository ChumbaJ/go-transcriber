package audioprocessor

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

func (a *AudioProcessor) prepareAudio(ctx context.Context, r io.Reader) (string, string, error) {
	started := time.Now()
	source, err := a.createTempFile(r)
	if err != nil {
		return "", "", err
	}
	sourcePath := source.Name()
	fmt.Printf("audio split: save input=%s\n", time.Since(started))
	if err := source.Close(); err != nil {
		os.Remove(sourcePath)
		return "", "", fmt.Errorf("close temp file: %w", err)
	}

	format, err := probeAudioFormat(ctx, sourcePath)
	if err != nil {
		os.Remove(sourcePath)
		return "", "", err
	}
	if format == "mp3" || format == "m4a" || format == "ogg" {
		fmt.Printf("audio split: input is %s, encoding skipped\n", format)
		return sourcePath, format, nil
	}

	mp3, err := os.CreateTemp("", "audio-*.mp3")
	if err != nil {
		os.Remove(sourcePath)
		return "", "", fmt.Errorf("create temp mp3: %w", err)
	}
	mp3Path := mp3.Name()
	started = time.Now()
	encodeErr := a.encodeFile(ctx, sourcePath, mp3)
	closeErr := mp3.Close()
	if encodeErr != nil || closeErr != nil {
		os.Remove(mp3Path)
		os.Remove(sourcePath)
		if encodeErr != nil {
			return "", "", fmt.Errorf("encode file: %w", encodeErr)
		}
		return "", "", fmt.Errorf("close temp mp3: %w", closeErr)
	}
	fmt.Printf("audio split: encode mp3=%s\n", time.Since(started))
	if err := os.Remove(sourcePath); err != nil {
		os.Remove(mp3Path)
		return "", "", fmt.Errorf("remove source temp file: %w", err)
	}
	return mp3Path, "mp3", nil
}

func (a *AudioProcessor) createTempFile(r io.Reader) (*os.File, error) {
	f, err := os.CreateTemp("", "audio-*")
	if err != nil {
		return nil, fmt.Errorf("os create temp: %w", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, fmt.Errorf("io copy: %w", err)
	}
	return f, nil
}
