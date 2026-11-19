package chunker_test

import (
	"testing"

	"personal-os/internal/chunker"
)

func TestBoundaryChunker(t *testing.T) {
	c := chunker.NewBoundaryChunker(100, 250)

	script := `Hello, friend. It is happening again. The world is too loud, and everyone is pretending to be okay. But you and I know the truth. Focus on what is right in front of you. Control only what is within your hands.`

	chunks, err := c.Chunk(script)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected chunks, got empty slice")
	}

	for i, chunk := range chunks {
		t.Logf("Chunk [%d] (%d chars): \"%s\"", i, len(chunk.Text), chunk.Text)
		if len(chunk.Text) > 250 {
			t.Errorf("Chunk [%d] exceeds 250 char limit: %d", i, len(chunk.Text))
		}
	}
}
