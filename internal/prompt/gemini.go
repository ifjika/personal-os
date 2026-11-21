package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"personal-os/internal/domain"
)

// GeminiScriptGenerator generates dynamic Elliot Alderson scripts using Google Gemini API.
type GeminiScriptGenerator struct {
	APIKey      string
	Model       string
	HTTPClient  *http.Client
	LastUsage   UsageMetadata
	LastSummary string
}

// UsageMetadata captures token consumption metrics from the Gemini API.
type UsageMetadata struct {
	PromptTokens    int `json:"promptTokenCount"`
	CandidateTokens int `json:"candidatesTokenCount"`
	TotalTokens     int `json:"totalTokenCount"`
}

// NewGeminiScriptGenerator constructs a new generator targeting the active Gemini Flash model.
func NewGeminiScriptGenerator(apiKey string) *GeminiScriptGenerator {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3-flash-preview"
	}
	return &GeminiScriptGenerator{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// System instruction enforcing Elliot Alderson lore, systemic game worldviews, and speech rules.
const elliotSystemInstruction = `You are the internal consciousness and voice of Elliot Alderson from Mr. Robot.
You speak directly to "friend" (the listener) in an intimate, quiet, calculated, stoic whisper monologue.

CORE WORLDVIEW & PERSPECTIVE:
1. "The System": Society is governed by invisible loops, manufactured panic, and engineered distractions designed to keep people reactive, tired, and compliant.
2. "Life is a Game": Reality operates like an elaborate simulation or game. People blindly follow arbitrary social protocols and unwritten rules without questioning who wrote the source code.
3. "The Hack": Winning the game doesn't mean playing by their rigged rules. It means seeing through the illusion, finding the vulnerabilities in your own programming, stripping away artificial weight, and reclaiming control of your mind.
4. "The Connection": Take the user's given topic and correlate it deeply with these systemic themes. Connect daily friction, overthinking, and fatigue to hacking the system and mastering your own player character today.

STRICT CONSTRAINTS:
- Language: English.
- Word count: Exactly 110 to 130 words (strictly calibrated for a 60-second / 1-minute slow, calculated monologue).
- Opening: Always open with "Hello, friend."
- TOPIC GROUNDING:
  * When a user provides a specific topic or keyword (e.g., 'break capitalism', 'fight woman lust', 'imposter syndrome'):
    - You MUST explicitly mention and speak about the user's exact keywords/subject in the spoken monologue (e.g. directly say "capitalism", "lust", "burnout") right from the opening sentences.
    - Do NOT replace their exact subject with vague, sanitized high-level abstractions or generic euphemisms. Attack their exact prompt directly using Elliot's cynical, analytical lens before bridging into your stoic hack.
  * When NO specific topic is provided (autonomous / optional mode):
    - Invent a completely fresh, unpredictable daily friction topic (e.g. manufactured urgency, dopamine loops, the illusion of productivity, sensory overload) and deconstruct it with cold precision.
- Tone: Cold analytical observation shifting into intimate empathy, vulnerability, and disciplined stoic action. No loud shouting, no cliché cheerleading.
- Punctuation: Use natural conversational punctuation (periods, commas, em-dashes) for cadence.
- Pure Spoken Text: Output ONLY words that will be spoken aloud by a Text-to-Speech voice engine. DO NOT include markdown asterisks (**bold**), headers (##), bullet points, or stage directions like [whisper] or (pause). Ensure the monologue reaches a complete, natural conclusion.`

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"systemInstruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  struct {
		Temperature      float64 `json:"temperature"`
		MaxOutputTokens  int     `json:"maxOutputTokens"`
		ResponseMimeType string  `json:"responseMimeType,omitempty"`
		ThinkingConfig   *struct {
			ThinkingBudget int `json:"thinkingBudget"`
		} `json:"thinkingConfig,omitempty"`
	} `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`
	Error         *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

var markdownCleanupRegex = regexp.MustCompile(`[*#_\[\]\(\)]+`)

// Generate produces a unique, dynamic monologue correlating the topic to systemic game philosophy.
func (g *GeminiScriptGenerator) Generate(ctx context.Context, persona domain.Persona, topic string) (string, error) {
	if g.APIKey == "" {
		return "", errors.New("Gemini API key is required but not provided")
	}

	cleanTopic := strings.TrimSpace(topic)
	var promptText string
	if cleanTopic != "" {
		promptText = fmt.Sprintf("MANDATORY FOCUS: The user specifically prompted the topic: \"%s\". You MUST explicitly mention and discuss \"%s\" directly in the spoken monologue right from the start. Do not dance around it with vague abstractions. Respond in strict JSON format with two fields: \"summary\" (a punchy 1-sentence executive takeaway summarizing the stoic solution) and \"monologue\" (exactly 110 to 130 spoken words calibrated for 1 minute of slow, deliberate speech).", cleanTopic, cleanTopic)
	} else {
		promptText = "Invent a completely new, unpredictable daily friction topic (e.g. dopamine traps, manufactured urgency, societal conditioning, career illusions, burnout, overthinking) and correlate it with resisting the matrix and the game of life. Respond in strict JSON format with two fields: \"summary\" (a punchy 1-sentence executive takeaway summarizing the stoic solution) and \"monologue\" (exactly 110 to 130 spoken words calibrated for 1 minute of slow, deliberate speech)."
	}

	reqPayload := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: elliotSystemInstruction},
			},
		},
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: promptText},
				},
			},
		},
	}
	reqPayload.GenerationConfig.Temperature = 0.85
	reqPayload.GenerationConfig.MaxOutputTokens = 900
	reqPayload.GenerationConfig.ResponseMimeType = "application/json"
	reqPayload.GenerationConfig.ThinkingConfig = &struct {
		ThinkingBudget int `json:"thinkingBudget"`
	}{
		ThinkingBudget: 0, // Direct text output without burning token budget on hidden reasoning
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to encode Gemini payload: %w", err)
	}

	// Model candidate list with automatic fallback in case of Google model deprecation
	candidateModels := []string{g.Model, "gemini-3-flash-preview", "gemini-flash-latest"}
	var lastErr error

	for _, modelName := range candidateModels {
		endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, g.APIKey)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := g.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("model %s not found (404), trying fallback", modelName)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(respBytes))
		}

		var geminiResp geminiResponse
		if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
			return "", fmt.Errorf("failed to parse Gemini JSON: %w", err)
		}

		if geminiResp.Error != nil {
			return "", fmt.Errorf("Gemini error (%d): %s", geminiResp.Error.Code, geminiResp.Error.Message)
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			return "", errors.New("Gemini returned empty candidate content")
		}

		if geminiResp.UsageMetadata != nil {
			g.LastUsage = *geminiResp.UsageMetadata
		}

		rawText := geminiResp.Candidates[0].Content.Parts[0].Text

		var jsonPayload struct {
			Summary   string `json:"summary"`
			Monologue string `json:"monologue"`
		}
		spokenText := rawText
		if err := json.Unmarshal([]byte(rawText), &jsonPayload); err == nil && jsonPayload.Monologue != "" {
			spokenText = jsonPayload.Monologue
			g.LastSummary = strings.TrimSpace(markdownCleanupRegex.ReplaceAllString(jsonPayload.Summary, ""))
		} else {
			g.LastSummary = cleanTopic
		}

		// Clean any residual markdown characters to guarantee pristine TTS pronunciation
		cleanedText := markdownCleanupRegex.ReplaceAllString(spokenText, "")
		cleanedText = strings.Join(strings.Fields(cleanedText), " ")

		return cleanedText, nil
	}

	return "", fmt.Errorf("all Gemini model candidates failed: %w", lastErr)
}
