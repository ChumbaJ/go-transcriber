package audioprocessor

import (
	"context"
	"math"

	"github.com/ChumbaJ/go-transcriber/internal/job"
)

const (
	targetBytes = 18 * 1024 * 1024
	maxShift    = 20
)

func (a *AudioProcessor) findNearestPauseBefore(limit, min float64, pauses []Silence) float64 {
	cut := limit
	best := -1.0

	for _, p := range pauses {
		if p.Mid > limit || p.Mid <= min {
			continue
		}

		if p.Mid > best {
			best = p.Mid
			cut = p.Mid
		}
	}

	if best >= 0 && limit-cut > maxShift {
		return limit
	}

	return cut
}

func (a *AudioProcessor) cutChunks(ctx context.Context, audioDuration, bitrate float64, pauses []Silence, path, format string) ([][]byte, error) {
	start := float64(0)
	var result [][]byte

	seconds := float64(targetBytes*8) / bitrate

	for start < audioDuration {
		limit := math.Min(start+seconds, audioDuration)
		cutPoint := a.findNearestPauseBefore(limit, start, pauses)

		chunk, err := a.copySlice(ctx, path, format, start, cutPoint)
		if err != nil {
			return nil, err
		}

		for len(chunk) > int(job.MaxChunkSize) {
			ratio := float64(job.MaxChunkSize) / float64(len(chunk))
			rough := start + (cutPoint-start)*ratio
			newCut := a.findNearestPauseBefore(rough, start, pauses)

			cutPoint = newCut
			chunk, err = a.copySlice(ctx, path, format, start, cutPoint)
			if err != nil {
				return nil, err
			}
		}

		result = append(result, chunk)
		start = cutPoint
	}

	return result, nil
}
