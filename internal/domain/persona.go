package domain

import "errors"

// Persona defines the vocal identity, acoustic attributes, and reference metadata.
type Persona struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	RefAudioPath   string `json:"ref_audio_path"`
	RefTranscript  string `json:"ref_transcript"`
	ToneDirective  string `json:"tone_directive"`
	TargetLanguage string `json:"target_language"`
}

// DefaultElliotPersona represents the baseline Elliot Alderson voice preset.
var DefaultElliotPersona = Persona{
	ID:             "elliot",
	Name:           "Elliot Alderson",
	RefAudioPath:   "storage/references/processed/elliot_ref_15s.wav",
	RefTranscript:  "Hello, friend. Hello, friend? That's lame. Maybe I should give you a name, but that's a slippery slope. You're only in my head. We have to remember that.",
	ToneDirective:  "Calm introspection, intimate whisper monologue, slow-to-moderate cadence, stoic, reflective.",
	TargetLanguage: "en",
}

// GetPersonaByID retrieves a persona configuration by unique identifier.
func GetPersonaByID(id string) (Persona, error) {
	switch id {
	case "elliot":
		return DefaultElliotPersona, nil
	default:
		return Persona{}, errors.New("persona not found in registry")
	}
}
