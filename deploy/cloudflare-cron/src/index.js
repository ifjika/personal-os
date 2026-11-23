/**
 * Personal OS: 24/7 Automated Motivation Engine (Cloudflare Worker Entrypoint)
 *
 * Schedule: Twice daily at 06:00 & 18:00 WIB (UTC: "0 23 * * *" & "0 11 * * *").
 * Architecture: Entrypoint -> Services (Pipeline, Gemini, Telegram, Diagnostics) -> Prompts.
 */
import { executePipeline } from "./services/pipeline.js";
import { runHealthDiagnostics } from "./services/diagnostics.js";
import { handleTelegramWebhook, registerTelegramWebhook, sendTelegramText } from "./services/telegram.js";

export default {
  async scheduled(event, env, ctx) {
    console.log(`⏰ Scheduled cron triggered at ${new Date().toISOString()} (cron: "${event.cron}")`);
    ctx.waitUntil(
      executePipeline(env).catch(async (err) => {
        console.error("Scheduled cron pipeline failed:", err);
        if (env.TELEGRAM_BOT_TOKEN && env.TELEGRAM_CHAT_ID) {
          await sendTelegramText(
            env.TELEGRAM_BOT_TOKEN,
            env.TELEGRAM_CHAT_ID,
            `⚠️ *Personal OS Cron Trigger Failed:*\n\`${err.message}\``
          );
        }
      })
    );
  },

  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    if (url.pathname === "/set-webhook") {
      const result = await registerTelegramWebhook(env, url.origin);
      return new Response(JSON.stringify(result, null, 2), {
        headers: { "Content-Type": "application/json" }
      });
    }

    if (url.pathname === "/webhook" && request.method === "POST") {
      return await handleTelegramWebhook(request, env, ctx);
    }

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

    if (url.pathname === "/ping") {
      const diag = await runHealthDiagnostics(env);
      return new Response(JSON.stringify(diag, null, 2), {
        headers: { "Content-Type": "application/json" }
      });
    }

    const diagnostics = {
      service: "Personal OS 24/7 Audio Motivation Engine",
      status: "ACTIVE",
      cronSchedule: "Twice daily: 06:00 & 18:00 WIB (UTC: 0 23 * * *, 0 11 * * *)",
      webhookUrl: `${url.origin}/webhook`,
      oneClickWebhookSetup: `${url.origin}/set-webhook`,
      manualTrigger: `${url.origin}/trigger`,
      livePing: `${url.origin}/ping`,
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
