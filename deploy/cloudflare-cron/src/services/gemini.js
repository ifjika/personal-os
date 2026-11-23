import { ELLIOT_SYSTEM_INSTRUCTION, buildGeminiPrompt } from "../prompts/elliot.js";

const GEMINI_MODELS = ["gemini-3-flash-preview", "gemini-flash-latest"];

function sanitizeSpeechText(text) {
  return (text || "").replace(/[*#_\[\]\(\)]+/g, "").replace(/\s+/g, " ").trim();
}

function parseGeminiPayload(rawText, customContext, tokens) {
  try {
    const parsed = JSON.parse(rawText);
    return {
      topic: (parsed.topic || customContext || "Seeing Through the Illusion").replace(/[*_]+/g, "").trim(),
      summary: (parsed.summary || "Reclaim control of your mind.").replace(/[*_]+/g, "").trim(),
      monologue: sanitizeSpeechText(parsed.monologue || rawText),
      tokens
    };
  } catch (_) {
    return {
      topic: customContext || "Seeing Through the Illusion",
      summary: "Execute your next move with deliberate precision.",
      monologue: sanitizeSpeechText(rawText),
      tokens
    };
  }
}

/**
 * Generates a 1-minute Elliot Alderson monologue, topic, and summary via Google Gemini API.
 */
export async function generateGeminiScript(apiKey, customContext = "") {
  const promptText = buildGeminiPrompt(customContext);
  let lastError;

  for (const model of GEMINI_MODELS) {
    const endpoint = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${apiKey}`;
    const payload = {
      systemInstruction: { parts: [{ text: ELLIOT_SYSTEM_INSTRUCTION }] },
      contents: [{ parts: [{ text: promptText }] }],
      generationConfig: {
        temperature: 0.90,
        maxOutputTokens: 950,
        responseMimeType: "application/json",
        thinkingConfig: { thinkingBudget: 0 }
      }
    };

    const resp = await fetch(endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    });

    if (resp.status === 404) {
      lastError = new Error(`Model ${model} returned 404`);
      continue;
    }

    if (!resp.ok) {
      const errText = await resp.text();
      throw new Error(`Gemini API error (${resp.status}): ${errText}`);
    }

    const data = await resp.json();
    const rawText = data.candidates?.[0]?.content?.parts?.[0]?.text;
    if (!rawText) throw new Error("Gemini returned empty candidate text");

    const usage = data.usageMetadata || {};
    const promptTokens = usage.promptTokenCount || 0;
    const candidateTokens = usage.candidatesTokenCount || 0;
    const tokens = {
      prompt: promptTokens,
      candidates: candidateTokens,
      total: usage.totalTokenCount || (promptTokens + candidateTokens)
    };

    return parseGeminiPayload(rawText, customContext, tokens);
  }

  throw lastError || new Error("All Gemini models failed");
}
