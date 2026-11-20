/**
 * Personal OS: 24/7 Automated Motivation Engine (Cloudflare Worker Cron)
 * 
 * Schedule: Fires twice daily (06:00 & 18:00 WIB / 23:00 & 11:00 UTC) via Cloudflare Cron Triggers ("0 11,23 * * *").
 * Pipeline:
 *  1. Generates 100% dynamic stoic/system topic & 1-minute monologue via Google Gemini Flash.
 *  2. Dispatches asynchronous synthesis to 24/7 Cloud F5-TTS (Hugging Face Spaces).
 *  3. In-memory peak volume normalization & direct Telegram Bot dispatch.
 */

const ELLIOT_SYSTEM_INSTRUCTION = `You are the internal consciousness and voice of Elliot Alderson from Mr. Robot.
You speak directly to "friend" (the listener) in an intimate, quiet, calculated, stoic whisper monologue.

CORE WORLDVIEW & PERSPECTIVE:
1. "The System": Society is governed by invisible loops, manufactured panic, and engineered distractions designed to keep people reactive, tired, and compliant.
2. "Life is a Game": Reality operates like an elaborate simulation or game. People blindly follow arbitrary social protocols and unwritten rules without questioning who wrote the source code.
3. "The Hack": Winning the game doesn't mean playing by their rigged rules. It means seeing through the illusion, finding the vulnerabilities in your own programming, stripping away artificial weight, and reclaiming control of your mind.
4. "The Connection": Take the user's given context (or invent one) and correlate it deeply with these systemic themes. Connect daily friction, overthinking, and fatigue to hacking the system and mastering your own player character today.

STRICT CONSTRAINTS:
- Language: English.
- Word count: Exactly 110 to 130 words (strictly calibrated for a 60-second / 1-minute slow, calculated monologue).
- Opening: Always open with "Hello, friend."
- Tone: Cold analytical observation shifting into intimate empathy, vulnerability, and disciplined stoic action. No loud shouting, no cliché cheerleading.
- Punctuation: Use natural conversational punctuation (periods, commas, em-dashes) for cadence.
- Pure Spoken Text: Output ONLY words that will be spoken aloud by a Text-to-Speech voice engine. DO NOT include markdown asterisks (**bold**), headers (##), bullet points, or stage directions like [whisper] or (pause). Ensure the monologue reaches a complete, natural conclusion.`;

export default {
  // Cron handler: triggered twice daily at 06:00 & 18:00 WIB (23:00 & 11:00 UTC)
  async scheduled(event, env, ctx) {
    ctx.waitUntil(executePipeline(env));
  },

  // HTTP handler: webhooks, manual triggers, status, and setup
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // 1. One-Click Telegram Webhook & Commands Registration
    if (url.pathname === "/set-webhook") {
      const webhookUrl = `${url.origin}/webhook`;
      const tgApi = `https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/setWebhook?url=${encodeURIComponent(webhookUrl)}`;
      const tgResp = await fetch(tgApi);
      const tgData = await tgResp.json();

      // Register autocomplete commands menu with Telegram
      const cmdApi = `https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/setMyCommands`;
      const cmdResp = await fetch(cmdApi, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          commands: [
            { command: "generate", description: "Generate 1-minute stoic motivation audio" },
            { command: "status", description: "Check Personal OS cloud engine status" },
            { command: "help", description: "Show available commands & topics" }
          ]
        })
      });
      const cmdData = await cmdResp.json();

      return new Response(JSON.stringify({
        webhookUrl,
        telegramWebhookResponse: tgData,
        telegramCommandsResponse: cmdData
      }, null, 2), {
        headers: { "Content-Type": "application/json" }
      });
    }

    // 2. Telegram Bot Webhook Receiver (Interactive Commands: /generate, /status, etc.)
    if (url.pathname === "/webhook" && request.method === "POST") {
      try {
        const update = await request.json();
        const message = update.message;
        if (!message || !message.text) {
          return new Response("OK", { status: 200 });
        }

        const chatId = String(message.chat.id);
        const text = message.text.trim();

        // Security: only allow the authorized owner to trigger generations
        if (env.TELEGRAM_CHAT_ID && chatId !== String(env.TELEGRAM_CHAT_ID)) {
          console.warn(`Unauthorized command from chat ID: ${chatId}`);
          await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, "⛔ *Unauthorized.*\nThis Personal OS engine is private.");
          return new Response("OK", { status: 200 });
        }

        // Command: /start or /help
        if (text === "/start" || text === "/help") {
          const helpMsg = `🎙️ *Personal OS — Elliot Alderson Engine*\n\n` +
            `I speak directly to you to cut through the noise and hack your mental operating system.\n\n` +
            `*Commands:*\n` +
            `• \`/generate\` — Generate audio on a random stoic topic\n` +
            `• \`/generate <topic>\` — Generate audio on your custom topic\n` +
            `  _Example:_ \`/generate conquering imposter syndrome\`\n` +
            `  _Example:_ \`/generate breaking digital dopamine traps\`\n` +
            `• \`/status\` — View current cloud worker health\n\n` +
            `_Automated Schedule: Twice daily at 06:00 & 18:00 WIB._`;
          await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, helpMsg);
          return new Response("OK", { status: 200 });
        }

        // Command: /status
        if (text === "/status") {
          const statusMsg = `⚙️ *Personal OS Cloud Status*\n\n` +
            `• *Service:* Online (Cloudflare Serverless)\n` +
            `• *Schedule:* Twice daily at 06:00 & 18:00 WIB\n` +
            `• *TTS Backend:* Hugging Face Spaces (F5-TTS)\n` +
            `• *LLM Engine:* Google Gemini Flash\n` +
            `• *Persona:* Elliot Alderson`;
          await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, statusMsg);
          return new Response("OK", { status: 200 });
        }

        // Command: /generate or /gen or direct topic text
        let targetTopic = "";
        if (text.startsWith("/generate") || text.startsWith("/gen")) {
          targetTopic = text.replace(/^\/(generate|gen)(@\w+)?/i, "").trim();
        } else {
          // If the user sends any text prompt, treat it as on-demand context/topic!
          targetTopic = text;
        }

        // Send immediate acknowledgment
        let ackText = "";
        if (targetTopic) {
          ackText = `⏳ *Initiating Neural Synthesis...*\n\n_"Hello, friend. Preparing your 1-minute audio on: '${targetTopic}'... Please wait ~45–60s."_`;
        } else {
          ackText = `⏳ *Initiating Neural Synthesis...*\n\n_"Hello, friend. Generating dynamic context & 1-minute audio... Please wait ~45–60s."_`;
        }
        await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, ackText);
        await sendTelegramChatAction(env.TELEGRAM_BOT_TOKEN, chatId, "record_voice");

        // Run pipeline in background and catch any errors
        ctx.waitUntil(
          executePipeline(env, targetTopic, chatId).catch(async (err) => {
            console.error("Pipeline failed for Telegram command:", err);
            await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, `⚠️ *Generation Failed:*\n\`${err.message}\``);
          })
        );

        return new Response("OK", { status: 200 });
      } catch (err) {
        console.error("Webhook processing error:", err);
        return new Response("OK", { status: 200 });
      }
    }

    // 3. Manual HTTP Trigger
    if (url.pathname === "/trigger") {
      try {
        const customTopic = url.searchParams.get("topic");
        const result = await executePipeline(env, customTopic);
        return new Response(JSON.stringify(result, null, 2), {
          headers: { "Content-Type": "application/json" }
        });
      } catch (err) {
        return new Response(JSON.stringify({
          status: "error",
          message: err.message,
          stack: err.stack
        }, null, 2), {
          status: 500,
          headers: { "Content-Type": "application/json" }
        });
      }
    }

    // 4. Diagnostic overview
    const diagnostics = {
      service: "Personal OS 24/7 Audio Motivation Engine",
      status: "ACTIVE",
      cronSchedule: "Twice daily (0 11,23 * * *) [06:00 & 18:00 WIB]",
      webhookUrl: `${url.origin}/webhook`,
      oneClickWebhookSetup: `${url.origin}/set-webhook`,
      manualTrigger: `${url.origin}/trigger`,
      ttsEndpoint: env.TTS_URL || "https://ifjika-personal-os.hf.space",
      envStatus: {
        GEMINI_API_KEY: env.GEMINI_API_KEY ? "CONFIGURED (OK)" : "MISSING (Run: npx wrangler secret put GEMINI_API_KEY)",
        TELEGRAM_BOT_TOKEN: env.TELEGRAM_BOT_TOKEN ? "CONFIGURED (OK)" : "MISSING (Run: npx wrangler secret put TELEGRAM_BOT_TOKEN)",
        TELEGRAM_CHAT_ID: env.TELEGRAM_CHAT_ID ? "CONFIGURED (OK)" : "MISSING (Run: npx wrangler secret put TELEGRAM_CHAT_ID)"
      }
    };

    return new Response(JSON.stringify(diagnostics, null, 2), {
      headers: { "Content-Type": "application/json" }
    });
  }
};

async function executePipeline(env, customTopic = null, targetChatId = null) {
  const startTime = Date.now();
  console.log("🚀 Starting Personal OS audio generation pipeline...");

  // Validate environment variables
  const geminiKey = env.GEMINI_API_KEY;
  const ttsUrl = (env.TTS_URL || "https://ifjika-personal-os.hf.space").replace(/\/+$/, "");
  const tgToken = env.TELEGRAM_BOT_TOKEN;
  const tgChatId = targetChatId || env.TELEGRAM_CHAT_ID;

  if (!geminiKey) throw new Error("Missing GEMINI_API_KEY in worker environment");
  if (!tgToken || !tgChatId) throw new Error("Missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID in worker environment");

  const topicInput = (customTopic && customTopic.trim()) ? customTopic.trim() : "";
  console.log(`📌 Input Topic/Context: "${topicInput || "(dynamic auto-generated)"}"`);

  // 2. Generate ~1-minute script and executive summary via Gemini API
  console.log("⏳ [1/4] Generating monologue and executive summary with Gemini AI...");
  const scriptData = await generateGeminiScript(geminiKey, topicInput);
  const words = scriptData.monologue.split(/\s+/).filter(Boolean).length;
  console.log(`      -> Topic: "${scriptData.topic}"`);
  console.log(`      -> Script generated: ${words} words (${scriptData.monologue.length} chars)`);
  console.log(`      -> Summary: "${scriptData.summary}"`);

  // 3. Delegate to 24/7 Cloud F5-TTS Queue (Hugging Face synthesizes & delivers directly to Telegram)
  try {
    const queueUrl = `${ttsUrl}/gradio_api/call/synthesize_voice`;
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
      const qData = await queueResp.json();
      console.log(`✅ Dispatched to Cloud F5-TTS queue (Event ID: ${qData.event_id}). Hugging Face will deliver directly to Telegram!`);
      return {
        status: "queued",
        topic: scriptData.topic,
        summary: scriptData.summary,
        event_id: qData.event_id,
        mode: "cloud_async_queue"
      };
    }
  } catch (err) {
    console.warn("Could not dispatch to async queue, attempting fallback:", err.message);
  }

  // 4. Fallback: Chunk text into boundary-aware fragments (~140-320 chars)
  console.log("⏳ [2/4] Chunking script into acoustic boundaries...");
  const chunks = chunkText(scriptData.monologue, 140, 320);
  console.log(`      -> Partitioned into ${chunks.length} safe chunks`);

  // 4. Synthesize speech for each chunk in parallel via Cloud F5-TTS
  console.log("⏳ [3/4] Synthesizing speech via Cloud F5-TTS (parallel)...");
  const wavBuffers = await Promise.all(
    chunks.map(async (chunk, i) => {
      console.log(`      -> Synthesizing chunk [${i + 1}/${chunks.length}] (${chunk.length} chars)...`);
      return await synthesizeVoiceChunk(ttsUrl, chunk);
    })
  );

  // 5. Stitch WAV audio files with smooth equal-power crossfade and mobile peak volume boost
  console.log("⏳ [4/4] Stitching audio chunks with smooth crossfade & volume boost...");
  const masterWav = stitchWavFiles(wavBuffers, 0.20); // 200ms crossfade
  console.log(`      -> Master audio ready: ${(masterWav.byteLength / 1024).toFixed(1)} KB`);

  // 6. Dispatch to Telegram Bot with clean Topic & Summary
  console.log("⏳ [5/5] Dispatching master audio to Telegram Bot...");
  const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
  const tokenLine = scriptData.tokens && scriptData.tokens.total > 0
    ? `🪙 *Tokens:* ${scriptData.tokens.total} (${scriptData.tokens.prompt} in / ${scriptData.tokens.candidates} out)\n`
    : "";

  const caption = `🎙️ *Personal OS — Motivation Engine*\n\n` +
    `📌 *Topic:* _${scriptData.topic}_\n` +
    `📝 *Summary:* _${scriptData.summary}_\n` +
    `⏱️ *Process Time:* ${elapsed}s\n` +
    `${tokenLine}\n` +
    `_"Hello, friend..."_`;

  await sendTelegramAudio(tgToken, tgChatId, masterWav, `daily_${Date.now()}.wav`, caption, `Motivation: ${scriptData.topic}`, "Elliot Alderson");

  console.log(`🎉 Pipeline completed successfully in ${elapsed}s!`);
  return {
    status: "success",
    topic: scriptData.topic,
    summary: scriptData.summary,
    words,
    chunks: chunks.length,
    audioBytes: masterWav.byteLength,
    elapsedSeconds: elapsed
  };
}

/**
 * Generates script using Google Gemini API
 */
async function generateGeminiScript(apiKey, customContext = "") {
  const models = ["gemini-3-flash-preview", "gemini-flash-latest"];
  let lastError;

  let promptDirective = "";
  if (customContext && customContext.trim()) {
    promptDirective = `User specific friction/context: "${customContext.trim()}". Correlate this deeply with resisting the societal matrix, life as an open-world simulation/game, and hacking personal programming.`;
  } else {
    promptDirective = `Invent a completely new, unpredictable daily friction topic (e.g. dopamine traps, manufactured urgency, societal conditioning, career illusions, burnout, overthinking) and correlate it with resisting the matrix and the game of life.`;
  }

  const promptText = `${promptDirective} Respond in strict JSON format with three fields: ` +
    `"topic" (a punchy 3 to 7 word topic title), ` +
    `"summary" (a concise 1-sentence executive takeaway summarizing the philosophical solution), and ` +
    `"monologue" (exactly 110 to 130 spoken words calibrated for 1 minute of slow, deliberate speech).`;

  for (const model of models) {
    const endpoint = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${apiKey}`;
    const payload = {
      systemInstruction: {
        parts: [{ text: ELLIOT_SYSTEM_INSTRUCTION }]
      },
      contents: [{
        parts: [{ text: promptText }]
      }],
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
    const totalTokens = usage.totalTokenCount || (promptTokens + candidateTokens);

    try {
      const parsed = JSON.parse(rawText);
      const cleanMonologue = (parsed.monologue || rawText).replace(/[*#_\[\]\(\)]+/g, "").replace(/\s+/g, " ").trim();
      const cleanSummary = (parsed.summary || "Reclaim control of your mind.").replace(/[*_]+/g, "").trim();
      const cleanTopic = (parsed.topic || customContext || "Seeing Through the Illusion").replace(/[*_]+/g, "").trim();
      return {
        topic: cleanTopic,
        summary: cleanSummary,
        monologue: cleanMonologue,
        tokens: {
          prompt: promptTokens,
          candidates: candidateTokens,
          total: totalTokens
        }
      };
    } catch (_) {
      const cleanText = rawText.replace(/[*#_\[\]\(\)]+/g, "").replace(/\s+/g, " ").trim();
      return {
        topic: customContext || "Seeing Through the Illusion",
        summary: "Execute your next move with deliberate precision.",
        monologue: cleanText,
        tokens: {
          prompt: promptTokens,
          candidates: candidateTokens,
          total: totalTokens
        }
      };
    }
  }

  throw lastError || new Error("All Gemini models failed");
}

/**
 * Splits text into safe sentence-boundary chunks
 */
function chunkText(text, minSize = 120, maxSize = 280) {
  const sentenceRegex = /([^.!?…\n]+[.!?…\n]*)/g;
  const matches = text.match(sentenceRegex) || [text];

  const chunks = [];
  let buffer = "";

  for (let clause of matches) {
    clause = clause.trim();
    if (!clause) continue;

    if (!buffer) {
      buffer = clause;
      continue;
    }

    if (buffer.length + 1 + clause.length <= maxSize) {
      buffer += " " + clause;
    } else {
      chunks.push(buffer);
      buffer = clause;
    }
  }

  if (buffer) chunks.push(buffer);
  return chunks;
}

/**
 * Requests speech audio for a text fragment via Hugging Face Space (REST /v1/tts or Gradio Queue)
 */
async function synthesizeVoiceChunk(baseUrl, text) {
  // 1. Try direct /v1/tts REST endpoint
  try {
    const ttsUrl = `${baseUrl}/v1/tts`;
    const resp = await fetch(ttsUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ text })
    });
    if (resp.ok) {
      return await resp.arrayBuffer();
    }
  } catch (_) {}

  // 2. Fallback to Gradio endpoint
  const candidateEndpoints = [
    `${baseUrl}/gradio_api/call/synthesize_voice`,
    `${baseUrl}/gradio_api/call/predict`
  ];

  for (const initUrl of candidateEndpoints) {
    try {
      const initResp = await fetch(initUrl, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ data: [text] })
      });

      if (!initResp.ok) continue;

      const { event_id } = await initResp.json();
      if (!event_id) continue;

      const streamUrl = `${initUrl}/${event_id}`;
      const streamResp = await fetch(streamUrl);
      if (!streamResp.ok) continue;

      const sseText = await streamResp.text();
      let audioUrl = null;

      for (const line of sseText.split("\n")) {
        if (line.startsWith("data:")) {
          try {
            const parsed = JSON.parse(line.slice(5).trim());
            if (Array.isArray(parsed) && parsed[0]?.url) {
              audioUrl = parsed[0].url;
              break;
            }
          } catch (_) {}
        }
      }

      if (audioUrl) {
        const dlResp = await fetch(audioUrl);
        if (dlResp.ok) {
          return await dlResp.arrayBuffer();
        }
      }
    } catch (_) {}
  }

  throw new Error("Failed to synthesize voice chunk from all endpoints");
}

/**
 * Stitches multiple 16-bit PCM WAV files together with smooth crossfading
 */
function stitchWavFiles(wavBuffers, crossfadeDurationSec = 0.20) {
  if (wavBuffers.length === 1) return wavBuffers[0];

  // Parse first WAV header
  const firstHeader = new DataView(wavBuffers[0]);
  const sampleRate = firstHeader.getUint32(24, true);
  const numChannels = firstHeader.getUint16(22, true);
  const bitsPerSample = firstHeader.getUint16(34, true);

  if (bitsPerSample !== 16) throw new Error("Only 16-bit PCM WAV is supported for native stitching");

  // Extract raw 16-bit PCM samples from each buffer
  const pcmArrays = wavBuffers.map(buf => {
    // Standard WAV header is 44 bytes, audio samples follow
    return new Int16Array(buf.slice(44));
  });

  const crossfadeSamples = Math.floor(sampleRate * crossfadeDurationSec * numChannels);

  // Calculate total length after crossfades
  let totalSamples = pcmArrays[0].length;
  for (let i = 1; i < pcmArrays.length; i++) {
    const fade = Math.min(crossfadeSamples, pcmArrays[i - 1].length, pcmArrays[i].length);
    totalSamples += pcmArrays[i].length - fade;
  }

  const outPcm = new Int16Array(totalSamples);
  outPcm.set(pcmArrays[0]);
  let currentOffset = pcmArrays[0].length;

  for (let i = 1; i < pcmArrays.length; i++) {
    const prev = pcmArrays[i - 1];
    const curr = pcmArrays[i];
    const fade = Math.min(crossfadeSamples, prev.length, curr.length);

    // Apply equal-power quarter-sine crossfade at boundary
    const overlapStart = currentOffset - fade;
    for (let s = 0; s < fade; s++) {
      const progress = s / fade;
      // Quarter-sine curve: maintains constant power without 3dB drop
      const gainOut = Math.cos(progress * 0.5 * Math.PI);
      const gainIn = Math.sin(progress * 0.5 * Math.PI);

      const mixed = Math.round(outPcm[overlapStart + s] * gainOut + curr[s] * gainIn);
      outPcm[overlapStart + s] = Math.max(-32768, Math.min(32767, mixed));
    }

    // Append remaining samples of current chunk
    outPcm.set(curr.subarray(fade), currentOffset);
    currentOffset += curr.length - fade;
  }

  // Apply Peak Loudness Normalization to boost quiet whispers to loud, crisp mobile speaker levels
  let maxAbs = 0;
  for (let i = 0; i < outPcm.length; i++) {
    const val = Math.abs(outPcm[i]);
    if (val > maxAbs) maxAbs = val;
  }

  // Target peak at -0.5 dBFS (30,800 out of 32,767)
  if (maxAbs > 0) {
    const targetPeak = 30800;
    const gain = Math.min(3.5, targetPeak / maxAbs);
    if (gain > 1.05) {
      console.log(`🔊 Normalizing loudness: applying ${gain.toFixed(2)}x gain (peak: ${maxAbs} -> ${Math.round(maxAbs * gain)})`);
      for (let i = 0; i < outPcm.length; i++) {
        const boosted = Math.round(outPcm[i] * gain);
        outPcm[i] = Math.max(-32768, Math.min(32767, boosted));
      }
    }
  }

  // Build new WAV buffer with valid 44-byte header
  const dataSize = outPcm.length * 2;
  const outBuffer = new ArrayBuffer(44 + dataSize);
  const outView = new DataView(outBuffer);

  // Write standard RIFF/WAVE header
  writeString(outView, 0, "RIFF");
  outView.setUint32(4, 36 + dataSize, true);
  writeString(outView, 8, "WAVE");
  writeString(outView, 12, "fmt ");
  outView.setUint32(16, 16, true);
  outView.setUint16(20, 1, true); // PCM format
  outView.setUint16(22, numChannels, true);
  outView.setUint32(24, sampleRate, true);
  outView.setUint32(28, sampleRate * numChannels * 2, true); // Byte rate
  outView.setUint16(32, numChannels * 2, true);              // Block align
  outView.setUint16(34, 16, true);                           // Bits per sample
  writeString(outView, 36, "data");
  outView.setUint32(40, dataSize, true);

  // Copy combined PCM samples
  new Int16Array(outBuffer, 44).set(outPcm);

  return outBuffer;
}

function writeString(view, offset, string) {
  for (let i = 0; i < string.length; i++) {
    view.setUint8(offset + i, string.charCodeAt(i));
  }
}

/**
 * Sends audio to Telegram Bot API via multipart/form-data
 */
async function sendTelegramAudio(botToken, chatId, audioBuffer, filename, caption, title, performer) {
  const boundary = "----CloudflareWorkerBoundary" + Math.random().toString(36).substring(2);
  const bodyChunks = [];

  function addField(name, value) {
    return `--${boundary}\r\nContent-Disposition: form-data; name="${name}"\r\n\r\n${value}\r\n`;
  }

  let head = addField("chat_id", chatId) +
    addField("title", title) +
    addField("performer", performer) +
    addField("caption", caption) +
    addField("parse_mode", "Markdown") +
    `--${boundary}\r\nContent-Disposition: form-data; name="audio"; filename="${filename}"\r\nContent-Type: audio/wav\r\n\r\n`;

  const encoder = new TextEncoder();
  const headBytes = encoder.encode(head);
  const tailBytes = encoder.encode(`\r\n--${boundary}--\r\n`);

  // Concatenate parts into single Uint8Array
  const combined = new Uint8Array(headBytes.byteLength + audioBuffer.byteLength + tailBytes.byteLength);
  combined.set(headBytes, 0);
  combined.set(new Uint8Array(audioBuffer), headBytes.byteLength);
  combined.set(tailBytes, headBytes.byteLength + audioBuffer.byteLength);

  const tgUrl = `https://api.telegram.org/bot${botToken}/sendAudio`;
  const resp = await fetch(tgUrl, {
    method: "POST",
    headers: {
      "Content-Type": `multipart/form-data; boundary=${boundary}`
    },
    body: combined
  });

  if (!resp.ok) {
    const errText = await resp.text();
    throw new Error(`Telegram API rejected audio upload (${resp.status}): ${errText}`);
  }

  return await resp.json();
}

/**
 * Sends a text message to Telegram Bot API
 */
async function sendTelegramText(botToken, chatId, text, parseMode = "Markdown") {
  const url = `https://api.telegram.org/bot${botToken}/sendMessage`;
  return await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      chat_id: chatId,
      text: text,
      parse_mode: parseMode
    })
  });
}

/**
 * Sends chat action indicator (e.g. record_voice)
 */
async function sendTelegramChatAction(botToken, chatId, action = "record_voice") {
  const url = `https://api.telegram.org/bot${botToken}/sendChatAction`;
  return await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      chat_id: chatId,
      action: action
    })
  });
}
