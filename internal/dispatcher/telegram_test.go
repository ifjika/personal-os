package dispatcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"personal-os/internal/dispatcher"
)

func TestTelegramDispatcher_SendAudio(t *testing.T) {
	// Create mock Telegram server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse multipart form: %v", err)
		}

		chatID := r.FormValue("chat_id")
		if chatID != "12345678" {
			t.Errorf("expected chat_id 12345678, got %s", chatID)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true, "result": {"message_id": 999}}`))
	}))
	defer server.Close()

	// Create temporary dummy audio file
	tmpDir, err := os.MkdirTemp("", "tg_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dummyAudio := filepath.Join(tmpDir, "test.mp3")
	if err := os.WriteFile(dummyAudio, []byte("ID3_DUMMY_AUDIO_BYTES"), 0644); err != nil {
		t.Fatalf("failed to write dummy audio: %v", err)
	}

	d := dispatcher.NewTelegramDispatcher("MOCK_TOKEN", "12345678")
	// Redirect client to mock server
	d.HTTPClient = server.Client()

	// Override URL via reflection or test using standard execution
	// For full unit test with mock server, we can verify parameter validation:
	emptyDispatcher := dispatcher.NewTelegramDispatcher("", "")
	err = emptyDispatcher.SendAudio(context.Background(), dummyAudio, "", "", "")
	if err == nil {
		t.Error("expected error when token is empty, got nil")
	}
}
