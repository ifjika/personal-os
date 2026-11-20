# 24/7 Cloudflare Cron Trigger & Telegram Dispatcher
## Personal OS: AI Audio Motivation Engine

This module provides a **100% free, 24/7 autonomous cloud cron job** that executes every 2 hours, synthesizes daily 1-minute introspective motivation monologues in Elliot Alderson's voice, and delivers the audio directly to your Telegram app.

---

## Architecture Overview

```text
┌────────────────────────────────┐
│  Cloudflare Worker Cron Trigger│
│  - Schedule: "0 */2 * * *"     │
│  - Triggers every 2 hours UTC  │
└──────────────┬─────────────────┘
               │ 1. Calls Google Gemini Flash API (1-minute script)
               ▼
┌────────────────────────────────┐
│   Google Gemini Flash AI API   │
│  - Systemic/Game Philosophy    │
│  - ~120 words / 60 seconds     │
└──────────────┬─────────────────┘
               │ 2. Dispatches acoustic chunks via SSE Queue
               ▼
┌────────────────────────────────┐
│  Hugging Face Spaces (F5-TTS)  │
│  - 24/7 Neural Voice Inference │
│  - Elliot Alderson Voice Clone │
└──────────────┬─────────────────┘
               │ 3. Stitches WAVs with Equal-Power Crossfade
               ▼
┌────────────────────────────────┐
│      Telegram Bot API          │
│  - POST /bot<TOKEN>/sendAudio  │
│  - Streams directly to Mobile  │
└────────────────────────────────┘
```

---

## 1. Prerequisites: Telegram Bot Setup

1. **Create Bot Token:**
   - Open Telegram and message [@BotFather](https://t.me/BotFather).
   - Send `/newbot`, name your bot (e.g., `Personal OS Audio`), and pick a username (e.g., `my_personal_os_bot`).
   - Copy the HTTP API token (e.g., `7123456789:AAH...`).

2. **Get your Chat ID:**
   - Send a message (e.g. `hello`) to your newly created bot.
   - Message [@userinfobot](https://t.me/userinfobot) on Telegram to get your numeric `Id` (e.g., `123456789`).
   - *(Optional: If using a Channel, add the bot as Admin to the channel and use `@your_channel_name` as the Chat ID).*

---

## 2. Deploying to Cloudflare

### Method A: Deploy via Wrangler CLI (Fastest)

1. Open your terminal and navigate to this folder:
   ```bash
   cd deploy/cloudflare-cron
   ```

2. Login to Cloudflare:
   ```bash
   npx wrangler login
   ```

3. Deploy the worker:
   ```bash
   npx wrangler deploy
   ```

4. Set the required secrets:
   ```bash
   npx wrangler secret put GEMINI_API_KEY
   npx wrangler secret put TELEGRAM_BOT_TOKEN
   npx wrangler secret put TELEGRAM_CHAT_ID
   ```
   *(Enter your values when prompted).*

---

### Method B: Deploy via Cloudflare Dashboard (No CLI Required)

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com/) -> **Workers & Pages** -> **Create application** -> **Create Worker**.
2. Name it `personal-os-cron` and click **Deploy**.
3. Click **Edit code**, paste the entire code from `src/index.js`, and click **Deploy**.
4. Go to **Settings** -> **Variables and Secrets**:
   - Add Secret: `GEMINI_API_KEY` (Your Google Gemini key)
   - Add Secret: `TELEGRAM_BOT_TOKEN` (From @BotFather)
   - Add Secret: `TELEGRAM_CHAT_ID` (Your Telegram ID)
   - Add Variable (optional): `TTS_URL` = `https://ifjika-personal-os.hf.space`
5. Go to **Settings** -> **Triggers** -> **Cron Triggers**:
   - Click **Add Cron Trigger**.
   - Set Cron: `0 */2 * * *` (Runs every 2 hours at minute 0).

---

## 3. Activate Interactive Telegram Bot Commands

To enable on-demand generation directly inside your Telegram chat:

1. Open this URL once in your web browser:
   ```text
   https://personal-os-cron.<your-subdomain>.workers.dev/set-webhook
   ```
   *(Telegram will respond with `{"ok": true, "result": true, "description": "Webhook was set"}`).*

2. Now you can chat with your bot anytime on your phone:
   - `/generate` — Generates a 1-minute audio on a random stoic topic.
   - `/generate <custom topic>` — Generates audio on your exact topic (e.g. `/generate conquering daily overthinking` or `/generate breaking out of the loop`).
   - `/status` — View cloud worker health and schedule info.
   - Or send **any text message** (e.g. `how to stop overthinking work`) and the bot will immediately treat it as a topic and generate a custom Elliot Alderson audio response!

---

## 4. Immediate Testing via Browser / Curl

You can also trigger a run directly without opening Telegram:

```bash
curl https://personal-os-cron.<your-subdomain>.workers.dev/trigger
```
Or with a custom topic:
```bash
curl "https://personal-os-cron.<your-subdomain>.workers.dev/trigger?topic=escaping+the+matrix"
```
