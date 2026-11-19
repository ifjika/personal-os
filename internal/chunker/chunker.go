package chunker

import (
	"errors"
	"regexp"
	"strings"

	"personal-os/internal/domain"
)

// BoundaryChunker partitions long-form scripts into breath-safe acoustic units.
type BoundaryChunker struct {
	MinSize int
	MaxSize int
}

// NewBoundaryChunker constructs a chunker with safe character limits.
func NewBoundaryChunker(minSize, maxSize int) *BoundaryChunker {
	if minSize <= 0 {
		minSize = 100
	}
	if maxSize <= minSize {
		maxSize = 250
	}
	return &BoundaryChunker{
		MinSize: minSize,
		MaxSize: maxSize,
	}
}

// sentenceRegex captures complete sentence clauses delimited by punctuation marks or newlines.
var sentenceRegex = regexp.MustCompile(`([^.!?…\n]+[.!?…\n]*)`)

// Chunk partitions the input script into an array of valid domain.JobChunk fragments.
func (c *BoundaryChunker) Chunk(script string) ([]domain.JobChunk, error) {
	trimmed := strings.TrimSpace(script)
	if trimmed == "" {
		return nil, errors.New("script text cannot be empty")
	}

	rawClauses := sentenceRegex.FindAllString(trimmed, -1)
	if len(rawClauses) == 0 {
		rawClauses = []string{trimmed}
	}

	var chunks []domain.JobChunk
	var currentBuffer strings.Builder
	chunkIndex := 0

	for _, clause := range rawClauses {
		clause = strings.TrimSpace(clause)
		if clause == "" {
			continue
		}

		if currentBuffer.Len() == 0 {
			currentBuffer.WriteString(clause)
			continue
		}

		potentialLen := currentBuffer.Len() + 1 + len(clause)

		if potentialLen <= c.MaxSize {
			currentBuffer.WriteString(" ")
			currentBuffer.WriteString(clause)
		} else {
			chunks = append(chunks, domain.JobChunk{
				Index: chunkIndex,
				Text:  currentBuffer.String(),
			})
			chunkIndex++

			currentBuffer.Reset()
			currentBuffer.WriteString(clause)
		}
	}

	if currentBuffer.Len() > 0 {
		chunks = append(chunks, domain.JobChunk{
			Index: chunkIndex,
			Text:  currentBuffer.String(),
		})
	}

	return chunks, nil
}
