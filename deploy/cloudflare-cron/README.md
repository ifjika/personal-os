# 24/7 Cloudflare Cron Trigger & Telegram Dispatcher
## Personal OS: AI Audio Motivation Engine

This module provides a **100% free, 24/7 autonomous cloud cron job** that executes every 3 hours WIB (`00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00 WIB`), synthesizes 1-minute introspective motivation monologues in Elliot Alderson's voice, and delivers the audio directly to your Telegram app.

---

## Architecture Overview

```text
┌───────────────────────────────────────────────────────────┐
│  Cloudflare Worker Cron Trigger                           │
│  - Schedule: "0 2,5,8,11,14,17,20,23 * * *"               │
│  - Triggers every 3 hours WIB (00:00–21:00 WIB)           │
└──────────────────────────────┬────────────────────────────┘
                               │ 1. Calls Google Gemini Flash API
                               ▼
┌───────────────────────────────────────────────────────────┐
│  Google Gemini Flash AI API                               │
│  - Mandatory Topic Grounding + Stoic/Game Philosophy      │
│  - ~110–130 words / 60 seconds                            │
└──────────────────────────────┬────────────────────────────┘
                               │ 2. Async Queue Dispatch (<200ms, 3x retry)
                               ▼
┌───────────────────────────────────────────────────────────┐
│  Hugging Face Spaces (F5-TTS)                             │
│  - 24/7 Neural Voice Inference + Peak Loudness Boost      │
│  - Direct Multipart Audio Upload to Telegram Bot API      │
└──────────────────────────────┬────────────────────────────┘
                               │ 3. POST /bot<TOKEN>/sendAudio
                               ▼
┌───────────────────────────────────────────────────────────┐
│  Telegram Bot API (Mobile & Desktop Delivery)             │
└───────────────────────────────────────────────────────────┘
```

### Modular Source Structure (`src/`)
- `index.js` — Worker entrypoint (`scheduled()` cron & `fetch()` HTTP router)
- `prompts/elliot.js` — Elliot Alderson system instruction & prompt builder
- `services/gemini.js` — Gemini API client with `503`/timeout fallback
- `services/pipeline.js` — Async Gradio Queue orchestrator with 3x cold-start retry
- `services/diagnostics.js` — Live parallel probes (`/ping`) & Space wake-up (`/wake`)
- `services/telegram.js` — Webhook command router & Markdown-safe Bot API helpers

---

## 1. Prerequisites: Telegram Bot Setup

1. **Create Bot Token:**
   - Open Telegram and message [@BotFather](https://t.me/BotFather).
   - Send `/newbot`, name your bot, and copy the HTTP API token.
2. **Get your Chat ID:**
   - Message your bot, then check [@userinfobot](https://t.me/userinfobot) for your numeric `Id`.

---

## 2. Deploying to Cloudflare

```bash
cd deploy/cloudflare-cron
npx wrangler login
npx wrangler deploy
```

Set the required encrypted secrets:

```bash
npx wrangler secret put GEMINI_API_KEY
npx wrangler secret put TELEGRAM_BOT_TOKEN
npx wrangler secret put TELEGRAM_CHAT_ID
npx wrangler secret put HF_TOKEN   # Recommended: enables /wake hard-restart via Hugging Face Hub API
```

---

## 3. Activate Interactive Telegram Bot Commands

1. Open this URL once in your web browser to register the webhook and `[Menu]` commands:
   ```text
   https://personal-os-cron.<your-subdomain>.workers.dev/set-webhook
   ```
2. Available commands in Telegram:
   - `/generate` — Generate a 1-minute audio on a novel, auto-generated stoic topic.
   - `/generate <topic>` — Generate audio directly addressing your custom topic.
   - `/ping` — Live health check across Cloudflare, Gemini AI, Hugging Face TTS & Bot API.
   - `/wake` — Wake up or restart a sleeping Hugging Face TTS Space remotely.
   - `/status` — View worker configuration and 3-hour schedule.
   - Or send **any text message** to treat it as an on-demand topic!

---

## 4. HTTP Endpoints (Testing & Diagnostics)

- `GET /` — Static worker configuration & secret presence overview
- `GET /ping` — Live parallel health check JSON report
- `GET /wake` — Trigger remote Hugging Face Space wake-up/restart
- `GET /trigger` or `GET /trigger?topic=...` — Manual pipeline trigger
