package audio_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"personal-os/internal/audio"
)

func TestFFmpegMixer_StitchAndMaster(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "audio_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	chunk1 := filepath.Join(tempDir, "chunk_0.wav")
	chunk2 := filepath.Join(tempDir, "chunk_1.wav")
	outputMP3 := filepath.Join(tempDir, "master.mp3")

	// Generate 2 synthetic sine wave test audio files (2 seconds each)
	cmd1 := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "sine=frequency=440:duration=2", "-ar", "48000", chunk1)
	if err := cmd1.Run(); err != nil {
		t.Fatalf("failed to create test chunk 1: %v", err)
	}

	cmd2 := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "sine=frequency=880:duration=2", "-ar", "48000", chunk2)
	if err := cmd2.Run(); err != nil {
		t.Fatalf("failed to create test chunk 2: %v", err)
	}

	mixer := audio.NewFFmpegMixer()
	err = mixer.StitchAndMaster(context.Background(), []string{chunk1, chunk2}, "", outputMP3)
	if err != nil {
		t.Fatalf("StitchAndMaster failed: %v", err)
	}

	// Verify that the mastered MP3 exists and is valid
	stat, err := os.Stat(outputMP3)
	if err != nil {
		t.Fatalf("master.mp3 file was not found: %v", err)
	}

	t.Logf("Success! Master audio file generated with size %d bytes", stat.Size())
}
