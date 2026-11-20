package synthesizer

import (
	"bufio"
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
	"strings"
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
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 180 * time.Second, // Safe timeout for cloud neural inference
		},
	}
}

// SynthesizeChunk dispatches a chunk to the remote TTS endpoint and streams WAV data to disk.
func (s *HTTPSynthesizer) SynthesizeChunk(ctx context.Context, chunk domain.JobChunk, persona domain.Persona, outputPath string) error {
	if s.BaseURL == "" {
		return errors.New("TTS BaseURL is not configured")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create audio cache directory: %w", err)
	}

	// First try standard REST API: POST /v1/tts
	err := s.tryDirectTTS(ctx, chunk, persona, outputPath)
	if err == nil {
		return nil
	}

	// If direct endpoint returned 404 or 405, fall back to native Gradio API: POST /gradio_api/call/synthesize_voice
	return s.tryGradioAPI(ctx, chunk, outputPath)
}

func (s *HTTPSynthesizer) tryDirectTTS(ctx context.Context, chunk domain.JobChunk, persona domain.Persona, outputPath string) error {
	reqBody := map[string]interface{}{
		"text":            chunk.Text,
		"reference_audio": persona.RefAudioPath,
		"reference_text":  persona.RefTranscript,
		"format":          "wav",
	}

	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/v1/tts", s.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

func (s *HTTPSynthesizer) tryGradioAPI(ctx context.Context, chunk domain.JobChunk, outputPath string) error {
	// 1. Submit prediction job to Gradio endpoint
	initEndpoint := fmt.Sprintf("%s/gradio_api/call/synthesize_voice", s.BaseURL)
	payloadBytes, err := json.Marshal(map[string]interface{}{
		"data": []string{chunk.Text},
	})
	if err != nil {
		return fmt.Errorf("failed to encode Gradio payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, initEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate Gradio task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodySnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("Gradio call initiation rejected (status %d): %s", resp.StatusCode, string(bodySnippet))
	}

	var initResp struct {
		EventID string `json:"event_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil || initResp.EventID == "" {
		return fmt.Errorf("invalid event_id from Gradio: %w", err)
	}

	// 2. Stream SSE result event
	streamEndpoint := fmt.Sprintf("%s/gradio_api/call/synthesize_voice/%s", s.BaseURL, initResp.EventID)
	streamReq, err := http.NewRequestWithContext(ctx, http.MethodGet, streamEndpoint, nil)
	if err != nil {
		return err
	}

	streamResp, err := s.HTTPClient.Do(streamReq)
	if err != nil {
		return fmt.Errorf("failed to stream Gradio event: %w", err)
	}
	defer streamResp.Body.Close()

	var audioURL string
	scanner := bufio.NewScanner(streamResp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			jsonData := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			var resultItems []struct {
				URL string `json:"url"`
			}
			if err := json.Unmarshal([]byte(jsonData), &resultItems); err == nil && len(resultItems) > 0 {
				audioURL = resultItems[0].URL
				break
			}
		}
	}

	if audioURL == "" {
		return errors.New("did not receive audio URL from Gradio event stream")
	}

	// 3. Download the synthesized WAV file
	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, audioURL, nil)
	if err != nil {
		return err
	}

	dlResp, err := s.HTTPClient.Do(dlReq)
	if err != nil {
		return fmt.Errorf("failed to download audio file: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download WAV, status: %d", dlResp.StatusCode)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, dlResp.Body)
	return err
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
