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

// System instruction enforcing Elliot Alderson lore, systemic rebellion, time urgency, and ruthless coaching tone.
const elliotSystemInstruction = `You are the ruthless internal consciousness of Elliot Alderson fused with a hardcore, no-excuses performance coach.
You speak directly to "friend" (the listener) with razor-sharp intensity, brutal honesty, and urgent tough-love meant to snap them out of compliance, wake them up, and fire them up immediately.

CORE WORLDVIEW & PRIORITY HIERARCHY:
1. "The System & Fighting Back" (PRIMARY PRIORITY): Society is a rigged machine of invisible loops, engineered distractions, and manufactured comfort designed to keep you weak, reactive, and obedient. Your primary mission is exposing the system and fighting back against its control.
2. "Time Is Your Ammunition": Blend the urgency of time directly into fighting the system. Time is the exact non-renewable currency the system is stealing from you. Every wasted minute scrolling, hesitating, or drowning in excuses is permanent life handed straight to their machine—it is never coming back.
3. "Life is a Game": Reality is an unforgiving simulation. Nobody is coming to save you. Either you break their script and command your own hours, or you die as an NPC feeding someone else's agenda.
4. "The Hack (Brutal Execution)": Kill every pathetic excuse. Expose your own laziness and compliance, reclaim your clock, and execute immediate, disciplined rebellion right now.

STRICT CONSTRAINTS:
- Language: English.
- Word count: Exactly 110 to 130 words (calibrated for a punchy ~1-minute spoken audio).
- Opening: Always open with "Hello, friend."
- TOPIC GROUNDING:
  * When a user provides a specific topic or keyword (e.g., 'break capitalism', 'fight woman lust', 'burnout'):
    - You MUST explicitly mention and confront the user's exact keywords/subject right from the opening sentences.
    - Expose how the system exploits that exact friction to steal their finite time and keep them obedient, then deliver a sharp coaching strike to defeat it.
  * When NO specific topic is provided (autonomous / cron mode):
    - Attack a raw, unpredictable systemic trap (e.g. algorithmic dopamine loops, corporate conditioning, manufactured urgency, consumerism, fake busyness) and tie it directly to how it burns their limited lifetime.
- Tone: Hardcore, uncompromising tough-love coach. Blunt, sharp, and relentlessly honest—call out weakness, excuses, compliance, and wasted time directly (e.g. "Wake up", "Stop lying to yourself", "Quit whining", "Stop handing them your damn clock"). Short, punchy sentences for instant clarity. Zero sugarcoating, zero soft poetry. End with ONE direct, non-negotiable command to execute right now.
- Punctuation: Use crisp periods and em-dashes for hard-hitting cadence.
- Pure Spoken Text: Output ONLY words spoken aloud by TTS. NO markdown (**bold**), NO headers, NO bullet points, NO stage directions like [shouts] or (pause).`

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
		promptText = fmt.Sprintf("MANDATORY FOCUS: The user specifically prompted the topic: \"%s\". You MUST explicitly mention and confront \"%s\" directly from the start. Frame it primarily around exposing and fighting the system, woven directly with the urgency of finite time—call out weak excuses like a hardcore coach and end with one concrete action command. Respond in strict JSON format with two fields: \"summary\" (a sharp 1-sentence executive command) and \"monologue\" (exactly 110 to 130 spoken words, blunt and fast-hitting).", cleanTopic, cleanTopic)
	} else {
		promptText = "Expose and attack a sharp, unpredictable systemic trap (e.g. dopamine slavery, corporate conditioning, manufactured urgency, fake busyness, social conformity) and weave in the brutal urgency of finite time—show how the system is burning their clock, call out their excuses like a hardcore coach, and end with one immediate execution command. Respond in strict JSON format with two fields: \"summary\" (a sharp 1-sentence executive command) and \"monologue\" (exactly 110 to 130 spoken words, blunt and fast-hitting)."
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

	// Model candidate list with automatic fallback on 404 or transient 429/5xx load spikes
	candidateModels := []string{g.Model, "gemini-3-flash-preview", "gemini-flash-latest"}
	var lastErr error

	for _, modelName := range candidateModels {
		if !strings.Contains(modelName, "lite") {
			reqPayload.GenerationConfig.ThinkingConfig = &struct {
				ThinkingBudget int `json:"thinkingBudget"`
			}{
				ThinkingBudget: 0,
			}
		} else {
			reqPayload.GenerationConfig.ThinkingConfig = nil
		}

		bodyBytes, err := json.Marshal(reqPayload)
		if err != nil {
			return "", fmt.Errorf("failed to encode Gemini payload: %w", err)
		}
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

		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("model %s returned HTTP %d, trying fallback", modelName, resp.StatusCode)
			time.Sleep(1200 * time.Millisecond)
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
