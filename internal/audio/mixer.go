package audio

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// AudioProcessor defines the audio post-processing contract for concatenation and mastering.
type AudioProcessor interface {
	StitchAndMaster(ctx context.Context, chunkPaths []string, ambientPath string, outputPath string) error
}

// FFmpegMixer executes audio graph pipelines using the local ffmpeg binary.
type FFmpegMixer struct{}

func NewFFmpegMixer() *FFmpegMixer {
	return &FFmpegMixer{}
}

// StitchAndMaster crossfades vocal chunks, layers ambient audio, and enforces EBU R128 mastering.
func (m *FFmpegMixer) StitchAndMaster(ctx context.Context, chunkPaths []string, ambientPath string, outputPath string) error {
	if len(chunkPaths) == 0 {
		return errors.New("vocal chunk list cannot be empty")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 1. Build input arguments for all vocal chunks
	var args []string
	for _, p := range chunkPaths {
		args = append(args, "-i", p)
	}

	// 2. Build dynamic filter_complex graph with 250ms equal-power quarter-sine crossfade
	var filterComplex string
	lastLabel := "0:a"

	if len(chunkPaths) > 1 {
		for i := 1; i < len(chunkPaths); i++ {
			outLabel := fmt.Sprintf("v%d", i)
			// Equal-power quarter-sine (qsin) curve maintains constant acoustic energy across transitions without 3dB volume dips
			filterComplex += fmt.Sprintf("[%s][%d:a]acrossfade=d=0.25:c1=qsin:c2=qsin[%s];", lastLabel, i, outLabel)
			lastLabel = outLabel
		}
	}

	// 3. Layer ambient background track if provided and file exists
	if ambientPath != "" {
		if _, err := os.Stat(ambientPath); err == nil {
			args = append(args, "-stream_loop", "-1", "-i", ambientPath)
			ambientIdx := len(chunkPaths)
			filterComplex += fmt.Sprintf("[%d:a]volume=0.12[amb];[%s][amb]amix=inputs=2:duration=first:dropout_transition=2[mixed];", ambientIdx, lastLabel)
			lastLabel = "mixed"
		}
	}

	// 4. Enforce vocal gain boost and broadcast loudness normalization (EBU R128 -14 LUFS for mobile)
	filterComplex += fmt.Sprintf("[%s]volume=1.8,loudnorm=I=-14:LRA=7:TP=-1.0[out]", lastLabel)

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[out]",
		"-c:a", "libmp3lame",
		"-b:a", "320k",
		"-ar", "48000",
		"-y", outputPath,
	)

	// 5. Execute FFmpeg subprocess
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg execution failed: %w\nOutput log:\n%s", err, string(outputBytes))
	}

	return nil
}
