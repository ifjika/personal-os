import { generateGeminiScript } from "./gemini.js";

const DEFAULT_TTS_URL = "https://ifjika-personal-os.hf.space";
const MAX_QUEUE_ATTEMPTS = 3;
const RETRY_DELAY_MS = 4000;

/**
 * Orchestrates script generation via Gemini and async TTS dispatch to Hugging Face Gradio Queue.
 * Includes a 3-attempt retry loop to gracefully wake Hugging Face Spaces from cold sleep.
 */
export async function executePipeline(env, customTopic = null, targetChatId = null) {
  const startTime = Date.now();
  console.log("🚀 Starting Personal OS audio generation pipeline...");

  const geminiKey = env.GEMINI_API_KEY;
  const ttsUrl = (env.TTS_URL || DEFAULT_TTS_URL).replace(/\/+$/, "");
  const tgToken = env.TELEGRAM_BOT_TOKEN;
  const tgChatId = targetChatId || env.TELEGRAM_CHAT_ID;

  if (!geminiKey) throw new Error("Missing GEMINI_API_KEY in worker environment");
  if (!tgToken || !tgChatId) throw new Error("Missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID in worker environment");

  const topicInput = (customTopic && customTopic.trim()) ? customTopic.trim() : "";
  console.log(`📌 Input Topic/Context: "${topicInput || "(dynamic auto-generated)"}"`);

  console.log("⏳ [1/2] Generating monologue and executive summary with Gemini AI...");
  const scriptData = await generateGeminiScript(geminiKey, topicInput);
  const words = scriptData.monologue.split(/\s+/).filter(Boolean).length;
  console.log(`      -> Topic: "${scriptData.topic}" (${words} words)`);

  const queueUrl = `${ttsUrl}/gradio_api/call/synthesize_voice`;
  let queueData = null;
  let lastQueueError = null;

  for (let attempt = 1; attempt <= MAX_QUEUE_ATTEMPTS; attempt++) {
    try {
      console.log(`⏳ [2/2] Calling Cloud F5-TTS queue (attempt ${attempt}/${MAX_QUEUE_ATTEMPTS})...`);
      const queueResp = await fetch(queueUrl, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          data: [
            scriptData.monologue,
            scriptData.topic,
            scriptData.summary,
            String(tgChatId),
            tgToken,
            scriptData.tokens ? scriptData.tokens.total : 0,
            scriptData.tokens ? scriptData.tokens.prompt : 0,
            scriptData.tokens ? scriptData.tokens.candidates : 0
          ]
        })
      });

      if (queueResp.ok) {
        queueData = await queueResp.json();
        console.log(`✅ Dispatched to Cloud F5-TTS queue (Event ID: ${queueData.event_id}).`);
        break;
      }

      const errText = await queueResp.text().catch(() => "");
      lastQueueError = new Error(`Hugging Face Space returned HTTP ${queueResp.status}: ${errText}`);
      console.warn(`[Attempt ${attempt}/${MAX_QUEUE_ATTEMPTS}] HF Queue HTTP ${queueResp.status}`);
    } catch (err) {
      lastQueueError = err;
      console.warn(`[Attempt ${attempt}/${MAX_QUEUE_ATTEMPTS}] Network error: ${err.message}`);
    }

    if (attempt < MAX_QUEUE_ATTEMPTS) {
      await new Promise((r) => setTimeout(r, RETRY_DELAY_MS));
    }
  }

  if (queueData && queueData.event_id) {
    return {
      status: "queued",
      topic: scriptData.topic,
      summary: scriptData.summary,
      event_id: queueData.event_id,
      mode: "cloud_async_queue",
      elapsedSeconds: ((Date.now() - startTime) / 1000).toFixed(1)
    };
  }

  throw new Error(
    `Hugging Face Space (${ttsUrl}) is sleeping or unreachable. ${lastQueueError ? lastQueueError.message : "Request timed out"}. Please visit ${ttsUrl} to wake up the space.`
  );
}
