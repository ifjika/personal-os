package prompt

import (
	"context"
	"fmt"
	"strings"

	"personal-os/internal/domain"
)

// ScriptGenerator defines the abstraction contract for motivation monologue generation.
type ScriptGenerator interface {
	Generate(ctx context.Context, persona domain.Persona, topic string) (string, error)
}

// BuiltinScriptGenerator produces structured scripts adhering to Elliot Alderson's vocal rhythm.
type BuiltinScriptGenerator struct{}

func NewBuiltinScriptGenerator() *BuiltinScriptGenerator {
	return &BuiltinScriptGenerator{}
}

// Generate composes a structured introspective monologue script.
func (g *BuiltinScriptGenerator) Generate(ctx context.Context, persona domain.Persona, topic string) (string, error) {
	cleanTopic := strings.ToLower(strings.TrimSpace(topic))
	if cleanTopic == "" {
		cleanTopic = "the silence between your thoughts"
	}

	// Structured monologue: Opening -> Observation -> Vulnerability -> Stoic Action
	script := fmt.Sprintf(
		"Hello, friend. It's happening again, isn't it? The endless noise. The feeling that everything demands your attention all at once. Today, we need to talk about %s. "+
			"Look around you. Most of the panic you feel isn't even yours. It's manufactured by an environment designed to keep you reactive, exhausted, and scattered. "+
			"We spend so much time fighting variables we can never touch, while the things that actually matter sit quietly in front of us, waiting. "+
			"So here is the protocol for today. Strip away the extra weight. Stop checking the imaginary scoreboards. "+
			"Focus only on the next honest move right before your eyes. You don't have to fix the whole machine today. Just execute this single step with absolute precision. "+
			"Stay calm. Breathe. I will be right here with you.",
		cleanTopic,
	)

	return script, nil
}
