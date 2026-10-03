package audioprocessor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type Silence struct {
	Start float64
	Mid   float64
	End   float64
}

func (a *AudioProcessor) findPauses(ctx context.Context, path string) ([]Silence, float64, error) {
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-i", path,
		"-af", "silencedetect=noise=-30dB:d=0.3",
		"-f", "null",
		"-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, 0, fmt.Errorf("detect silence: %w: %s", err, output)
	}

	s := bufio.NewScanner(bytes.NewReader(output))
	var silences []Silence
	var pending *float64
	var duration float64

	for s.Scan() {
		line := s.Text()

		if _, after, found := strings.Cut(line, "silence_start:"); found {
			v, err := strconv.ParseFloat(strings.TrimSpace(after), 64)
			if err == nil {
				v = math.Round(v*100) / 100
				pending = &v
			}
		}

		if _, after, found := strings.Cut(line, "silence_end:"); found {
			// у end есть хвост "| silence_duration: ...", режем по пробелу
			num, _, _ := strings.Cut(strings.TrimSpace(after), " ")
			v, err := strconv.ParseFloat(num, 64)
			if err == nil && pending != nil {
				v = math.Round(v*100) / 100
				silences = append(silences, Silence{
					Start: *pending,
					Mid:   math.Round((*pending+v)/2*100) / 100,
					End:   v,
				})
				pending = nil
			}
		}

		if _, after, found := strings.Cut(line, "Duration:"); found {
			// after = " 00:39:12.34, start: 0.000000, bitrate: ..."
			ts, _, _ := strings.Cut(strings.TrimSpace(after), ",")
			// ts = "00:39:12.34"
			if d, err := parseDuration(ts); err == nil {
				duration = d
			}
		}
	}

	if err := s.Err(); err != nil {
		return nil, 0, fmt.Errorf("scan silence output: %w", err)
	}

	return silences, duration, nil
}

func parseDuration(s string) (float64, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("bad duration: %q", s)
	}

	h, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}
	m, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}
	sec, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}

	return h*3600 + m*60 + sec, nil
}
