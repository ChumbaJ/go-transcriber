package audioprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func probeAudioFormat(ctx context.Context, path string) (string, error) {
	output, err := exec.CommandContext(ctx,
		"ffprobe", "-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "format=format_name:stream=codec_name",
		"-of", "json",
		path,
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("probe audio format: %w: %s", err, strings.TrimSpace(string(output)))
	}

	var probe struct {
		Streams []struct {
			CodecName string `json:"codec_name"`
		} `json:"streams"`
		Format struct {
			FormatName string `json:"format_name"`
		} `json:"format"`
	}
	if err := json.Unmarshal(output, &probe); err != nil {
		return "", fmt.Errorf("decode audio probe: %w", err)
	}
	if len(probe.Streams) == 0 {
		return "", fmt.Errorf("no audio stream found")
	}

	formats := strings.Split(probe.Format.FormatName, ",")
	for _, format := range formats {
		switch {
		case format == "mp3" && probe.Streams[0].CodecName == "mp3":
			return "mp3", nil
		case format == "m4a" && probe.Streams[0].CodecName == "aac":
			return "m4a", nil
		case format == "ogg" && (probe.Streams[0].CodecName == "vorbis" || probe.Streams[0].CodecName == "opus"):
			return "ogg", nil
		}
	}
	return "", nil // Other inputs use the MP3 encoding path.
}
