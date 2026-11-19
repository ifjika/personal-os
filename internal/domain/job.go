package domain

import "time"

// JobStatus represents the state machine lifecycle stages of a render job.
type JobStatus int

const (
	StatusPending JobStatus = iota
	StatusGeneratingScript
	StatusChunking
	StatusSynthesizing
	StatusMixing
	StatusCompleted
	StatusFailed
)

func (s JobStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusGeneratingScript:
		return "GENERATING_SCRIPT"
	case StatusChunking:
		return "CHUNKING"
	case StatusSynthesizing:
		return "SYNTHESIZING"
	case StatusMixing:
		return "MIXING"
	case StatusCompleted:
		return "COMPLETED"
	case StatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// JobChunk is an atomic, boundary-aligned text fragment ready for TTS synthesis.
type JobChunk struct {
	Index       int     `json:"index"`
	Text        string  `json:"text"`
	AudioPath   string  `json:"audio_path"`
	DurationSec float64 `json:"duration_sec"`
}

// RenderJob tracks all state, metadata, and progress of a daily motivation session.
type RenderJob struct {
	ID              string     `json:"id"`
	Topic           string     `json:"topic"`
	Persona         Persona    `json:"persona"`
	Status          JobStatus  `json:"status"`
	Script          string     `json:"script"`
	Chunks          []JobChunk `json:"chunks"`
	MasterAudioPath string     `json:"master_audio_path"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}
