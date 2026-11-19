package synthesizer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"personal-os/internal/domain"
)

// Synthesizer defines the neural voice synthesis client contract.
type Synthesizer interface {
	SynthesizeChunk(ctx context.Context, chunk domain.JobChunk, persona domain.Persona, outputPath string) error
}

// HTTPSynthesizer sends synthesis requests to a remote cloud or local TTS backend.
type HTTPSynthesizer struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewHTTPSynthesizer(baseURL string) *HTTPSynthesizer {
	return &HTTPSynthesizer{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 180 * time.Second, // Safe timeout for cloud neural inference
		},
	}
}

// ttsRequest represents the standard HTTP JSON payload for voice synthesis.
type ttsRequest struct {
	Text           string  `json:"text"`
	ReferenceAudio string  `json:"reference_audio,omitempty"`
	ReferenceText  string  `json:"reference_text,omitempty"`
	Temperature    float64 `json:"temperature,omitempty"`
	TopP           float64 `json:"top_p,omitempty"`
	Format         string  `json:"format,omitempty"`
}

// SynthesizeChunk dispatches a chunk to the remote TTS endpoint and streams WAV data to disk.
func (s *HTTPSynthesizer) SynthesizeChunk(ctx context.Context, chunk domain.JobChunk, persona domain.Persona, outputPath string) error {
	if s.BaseURL == "" {
		return errors.New("TTS BaseURL is not configured")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create audio cache directory: %w", err)
	}

	reqBody := ttsRequest{
		Text:           chunk.Text,
		ReferenceAudio: persona.RefAudioPath,
		ReferenceText:  persona.RefTranscript,
		Temperature:    0.7,
		TopP:           0.85,
		Format:         "wav",
	}

	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to serialize TTS payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/tts", s.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to TTS server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respSnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("TTS server responded with status %d: %s", resp.StatusCode, string(respSnippet))
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output audio file: %w", err)
	}
	defer outFile.Close()

	// Stream directly from network socket to disk (Zero-Heap Bloat)
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to stream audio bytes to disk: %w", err)
	}

	return nil
}

// EdgeTTSSynthesizer generates speech using Microsoft Edge neural voices (Zero-GPU fallback).
type EdgeTTSSynthesizer struct {
	Voice string
	Rate  string
}

func NewEdgeTTSSynthesizer(voice, rate string) *EdgeTTSSynthesizer {
	if voice == "" {
		voice = "en-US-ChristopherNeural"
	}
	if rate == "" {
		rate = "-8%"
	}
	return &EdgeTTSSynthesizer{
		Voice: voice,
		Rate:  rate,
	}
}

func (e *EdgeTTSSynthesizer) SynthesizeChunk(ctx context.Context, chunk domain.JobChunk, persona domain.Persona, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	voiceArg := fmt.Sprintf("--voice=%s", e.Voice)
	rateArg := fmt.Sprintf("--rate=%s", e.Rate)

	cmd := exec.CommandContext(ctx, "python", "-m", "edge_tts",
		voiceArg,
		rateArg,
		"--text", chunk.Text,
		"--write-media", outputPath,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("edge-tts synthesis failed: %w\nOutput: %s", err, string(out))
	}

	return nil
}

// MockSynthesizer produces synthetic tone audio for local development and offline testing.
type MockSynthesizer struct{}

func NewMockSynthesizer() *MockSynthesizer {
	return &MockSynthesizer{}
}

func (m *MockSynthesizer) SynthesizeChunk(ctx context.Context, chunk domain.JobChunk, persona domain.Persona, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-f", "lavfi", "-i", "sine=frequency=432:duration=3", "-ar", "48000", outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mock synthesizer failed to generate test audio: %w", err)
	}
	return nil
}
