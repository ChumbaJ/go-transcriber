package audioprocessor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func (a *AudioProcessor) encodeFile(ctx context.Context, path string, output *os.File) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-i", path,
		"-vn",
		"-c:a", "libmp3lame",
		"-b:a", "64k",
		"-f", "mp3",
		"-",
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr bytes.Buffer

	cmd.Stdout = output
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("encode mp3: %w: %s", err, stderr.String())
	}

	return nil
}

func (a *AudioProcessor) copySlice(ctx context.Context, path, format string, start, end float64) ([]byte, error) {
	switch format {
	case "mp3", "ogg":
		return a.copySliceToPipe(ctx, path, format, start, end)
	case "m4a":
		return a.copySliceM4A(ctx, path, start, end)
	default:
		return nil, fmt.Errorf("unsupported audio format: %q", format)
	}
}

func (a *AudioProcessor) copySliceToPipe(ctx context.Context, path, format string, start, end float64) ([]byte, error) {
	args := []string{
		"-hide_banner",
		"-nostdin",
	}
	args = append(args,
		"-ss", strconv.FormatFloat(start, 'f', 3, 64),
		"-to", strconv.FormatFloat(end, 'f', 3, 64),
		"-i", path,
		"-vn",
		"-c:a", "copy",
		"-f", format,
		"-",
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var buf bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &buf
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg copy %s slice: %w: %s", format, err, stderr.String())
	}

	return buf.Bytes(), nil
}

func (a *AudioProcessor) copySliceM4A(ctx context.Context, path string, start, end float64) ([]byte, error) {
	output, err := os.CreateTemp("", "chunk-*.m4a")
	if err != nil {
		return nil, fmt.Errorf("create temp m4a chunk: %w", err)
	}
	outputPath := output.Name()
	defer os.Remove(outputPath)
	if err := output.Close(); err != nil {
		return nil, fmt.Errorf("close temp m4a chunk: %w", err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner", "-loglevel", "error", "-nostdin", "-y",
		"-ss", strconv.FormatFloat(start, 'f', 3, 64),
		"-to", strconv.FormatFloat(end, 'f', 3, 64),
		"-i", path,
		"-map", "0:a:0", "-c:a", "copy", "-f", "ipod", outputPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg copy m4a slice: %w: %s", err, output)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read temp m4a chunk: %w", err)
	}
	return data, nil
}
