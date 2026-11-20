package dispatcher

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Dispatcher defines the contract for delivering finalized motivation audio.
type Dispatcher interface {
	SendAudio(ctx context.Context, filePath string, caption string, title string, performer string) error
}

// TelegramDispatcher delivers master audio files directly to a Telegram user or channel via Bot API.
type TelegramDispatcher struct {
	BotToken   string
	ChatID     string
	HTTPClient *http.Client
}

// NewTelegramDispatcher constructs a new Telegram dispatcher with safe timeouts.
func NewTelegramDispatcher(botToken, chatID string) *TelegramDispatcher {
	return &TelegramDispatcher{
		BotToken: botToken,
		ChatID:   chatID,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // Accommodate large file uploads on slower networks
		},
	}
}

// SendAudio uploads an audio file to Telegram using multipart/form-data.
func (t *TelegramDispatcher) SendAudio(ctx context.Context, filePath string, caption string, title string, performer string) error {
	if t.BotToken == "" {
		return errors.New("telegram bot token is required")
	}
	if t.ChatID == "" {
		return errors.New("telegram chat ID is required")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open audio file for telegram dispatch: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 1. Add chat_id field
	if err := writer.WriteField("chat_id", t.ChatID); err != nil {
		return err
	}

	// 2. Add title and performer metadata
	if title != "" {
		_ = writer.WriteField("title", title)
	}
	if performer != "" {
		_ = writer.WriteField("performer", performer)
	}

	// 3. Add caption with markdown formatting support
	if caption != "" {
		_ = writer.WriteField("caption", caption)
		_ = writer.WriteField("parse_mode", "Markdown")
	}

	// 4. Attach audio binary file
	part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create multipart form audio file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to pipe audio bytes into form: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// 5. Execute HTTP POST to Telegram Bot API
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendAudio", t.BotToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, body)
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram api network request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read telegram response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}
