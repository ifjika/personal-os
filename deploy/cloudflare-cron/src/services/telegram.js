import { executePipeline } from "./pipeline.js";
import { runHealthDiagnostics, wakeHuggingFaceSpace } from "./diagnostics.js";

function escapeMarkdown(str) {
  return String(str || "").replace(/([_*`\[])/g, "\\$1");
}

export async function sendTelegramText(botToken, chatId, text, parseMode = "Markdown") {
  const url = `https://api.telegram.org/bot${botToken}/sendMessage`;
  const payload = parseMode ? { chat_id: chatId, text, parse_mode: parseMode } : { chat_id: chatId, text };
  const resp = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });

  // Fallback to plain text if Telegram rejects Markdown entities (e.g. unescaped underscores)
  if (!resp.ok && parseMode) {
    console.warn(`Telegram Markdown send failed (HTTP ${resp.status}), retrying as plain text...`);
    return await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ chat_id: chatId, text })
    });
  }

  return resp;
}

export async function sendTelegramChatAction(botToken, chatId, action = "record_voice") {
  const url = `https://api.telegram.org/bot${botToken}/sendChatAction`;
  return await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ chat_id: chatId, action })
  });
}

/**
 * Registers Telegram webhook URL and native command autocomplete menu.
 */
export async function registerTelegramWebhook(env, origin) {
  const webhookUrl = `${origin}/webhook`;
  const tgResp = await fetch(
    `https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/setWebhook?url=${encodeURIComponent(webhookUrl)}`
  );
  const telegramWebhookResponse = await tgResp.json();

  const cmdResp = await fetch(`https://api.telegram.org/bot${env.TELEGRAM_BOT_TOKEN}/setMyCommands`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      commands: [
        { command: "generate", description: "Generate 1-minute stoic motivation audio" },
        { command: "ping", description: "Live health check (Gemini, Hugging Face & Bot)" },
        { command: "wake", description: "Wake up sleeping Hugging Face TTS Space" },
        { command: "status", description: "Check Personal OS configuration & schedule" },
        { command: "help", description: "Show available commands & topics" }
      ]
    })
  });
  const telegramCommandsResponse = await cmdResp.json();

  return { webhookUrl, telegramWebhookResponse, telegramCommandsResponse };
}

/**
 * Handles incoming Telegram Bot webhook updates with strict owner chat_id authorization.
 */
export async function handleTelegramWebhook(request, env, ctx) {
  try {
    const update = await request.json();
    const message = update.message;
    if (!message || !message.text) return new Response("OK", { status: 200 });

    const chatId = String(message.chat.id);
    const text = message.text.trim();

    if (env.TELEGRAM_CHAT_ID && chatId !== String(env.TELEGRAM_CHAT_ID)) {
      console.warn(`Unauthorized command from chat ID: ${chatId}`);
      await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, "⛔ *Unauthorized.*\nThis Personal OS engine is private.");
      return new Response("OK", { status: 200 });
    }

    if (text === "/start" || text === "/help") {
      const helpMsg = `🎙️ *Personal OS — Elliot Alderson Engine*\n\n` +
        `I speak directly to you to cut through the noise and hack your mental operating system.\n\n` +
        `*Commands:*\n` +
        `• \`/generate\` — Generate audio on a random stoic topic\n` +
        `• \`/generate <topic>\` — Generate audio on your custom topic\n` +
        `  _Example:_ \`/generate conquering imposter syndrome\`\n` +
        `• \`/ping\` — Live health check (Gemini, Hugging Face & Bot)\n` +
        `• \`/wake\` — Wake up sleeping Hugging Face TTS Space\n` +
        `• \`/status\` — View configuration & schedule\n\n` +
        `_Automated Schedule: Every 3 hours (00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00 WIB)._`;
      await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, helpMsg);
      return new Response("OK", { status: 200 });
    }

    if (text === "/status") {
      const statusMsg = `⚙️ *Personal OS Cloud Config*\n\n` +
        `• *Orchestrator:* Cloudflare Workers (Serverless)\n` +
        `• *Schedule:* Every 3 hours (00:00–21:00 WIB)\n` +
        `• *TTS Backend:* Hugging Face Spaces (F5-TTS)\n` +
        `• *LLM Engine:* Google Gemini Flash\n` +
        `• *Persona:* Elliot Alderson\n\n` +
        `_Tip: Send \`/ping\` for live health check or \`/wake\` if Hugging Face is sleeping._`;
      await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, statusMsg);
      return new Response("OK", { status: 200 });
    }

    if (text === "/wake" || text.startsWith("/wake@")) {
      await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, "⚡ *Sending wake-up signal to Hugging Face Space...*");
      ctx.waitUntil(
        wakeHuggingFaceSpace(env)
          .then(async (res) => {
            const icon = res.ok ? "🟢" : "🔴";
            const wakeMsg =
              `${icon} *Hugging Face Wake-Up*\n\n` +
              `• *Method:* ${escapeMarkdown(res.method)}\n` +
              `• *Result:* ${escapeMarkdown(res.message)}`;
            await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, wakeMsg);
          })
          .catch(async (err) => {
            await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, `⚠️ *Wake Failed:*\n${escapeMarkdown(err.message)}`);
          })
      );
      return new Response("OK", { status: 200 });
    }

    if (text === "/ping" || text.startsWith("/ping@")) {
      await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, "🔍 *Pinging Personal OS cloud services...*");
      ctx.waitUntil(
        runHealthDiagnostics(env)
          .then(async (diag) => {
            const reportMsg =
              `📡 *Personal OS — Live Ping Report*\n\n` +
              `• *Cloudflare Worker:* ${escapeMarkdown(diag.cloudflare)}\n` +
              `• *Env Secrets:* ${escapeMarkdown(diag.secrets)}\n` +
              `• *Google Gemini AI:* ${escapeMarkdown(diag.gemini)}\n` +
              `• *Hugging Face TTS:* ${escapeMarkdown(diag.huggingface)}\n` +
              `• *Telegram Bot API:* ${escapeMarkdown(diag.telegram)}\n\n` +
              `*Summary:* ${escapeMarkdown(diag.summary)}\n` +
              `_Checked at: ${diag.checkedAtWib} WIB_`;
            await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, reportMsg);
          })
          .catch(async (err) => {
            await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, `⚠️ *Ping Failed:*\n\`${err.message}\``);
          })
      );
      return new Response("OK", { status: 200 });
    }

    const targetTopic = (text.startsWith("/generate") || text.startsWith("/gen"))
      ? text.replace(/^\/(generate|gen)(@\w+)?/i, "").trim()
      : text;

    const ackText = targetTopic
      ? `⏳ *Initiating Neural Synthesis...*\n\n_"Hello, friend. Preparing your 1-minute audio on: '${targetTopic}'... Please wait ~45–60s."_`
      : `⏳ *Initiating Neural Synthesis...*\n\n_"Hello, friend. Generating dynamic context & 1-minute audio... Please wait ~45–60s."_`;

    await sendTelegramText(env.TELEGRAM_BOT_TOKEN, chatId, ackText);
    await sendTelegramChatAction(env.TELEGRAM_BOT_TOKEN, chatId, "record_voice");

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
