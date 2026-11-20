
# 🎙️ Personal OS

> _"Hello, friend. It's happening again. They have us convinced that if we stop moving, we stop existing..."_

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Version](https://img.shields.io/badge/Version-v0.4.0-brightgreen)](docs/CHANGELOG.md)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Telegram Bot](https://img.shields.io/badge/Telegram-Bot%20API-2CA5E0?logo=telegram)](https://core.telegram.org/bots/api)
[![Cloudflare Workers](https://img.shields.io/badge/Cloudflare-Serverless%20Cron-F38020?logo=cloudflare)](https://workers.cloudflare.com/)
[![Hugging Face](https://img.shields.io/badge/Hugging%20Face-F5--TTS%20Diffusion-FFD21E?logo=huggingface)](https://huggingface.co/)

**Personal OS** is an autonomous, 24/7 AI Audio Motivation Engine. It deconstructs modern fatigue, digital anxiety, and daily friction through the intimate, calculated, stoic consciousness of **Elliot Alderson (_Mr. Robot_)**.

Twice every day—or whenever you ask—it synthesizes a 1-minute voice monologue and delivers mastered, podcast-grade audio directly to your private Telegram chat.

---

## ✨ Highlights

- 🧠 **100% Dynamic Consciousness:** Powered by Google Gemini Flash. No static lists or repetitive clichés. It invents novel philosophical perspectives on treating reality as an open-world simulation and hacking your daily mental state.
- 🎙️ **Neural Voice Synthesis:** Uses deep diffusion voice cloning (**F5-TTS**) to recreate Elliot's distinctive quiet, intimate, analytical whisper.
- ⛅ **True $0 Serverless Infrastructure:** Runs 24/7 on free-tier edge computing (Cloudflare Workers Cron + Hugging Face Spaces + Telegram Bot API).
- 📲 **Interactive On-Demand Companion:** In addition to automated morning & evening broadcasts, you can send any thought or topic in Telegram to get an immediate, personalized 1-minute audio debrief.

---

## 🏗️ How It Works

```mermaid
flowchart LR
    A["⏰ Autonomous Cron\n(06:00 & 18:00 WIB)"] -->|Trigger| B["⛅ Serverless Orchestrator\n(Cloudflare Worker)"]
    U["📱 Telegram Message\n(Any topic or /generate)"] -->|Webhook| B

    B -->|Dynamic Scripting| C["🧠 Google Gemini AI\n(Stoic Reasoning Engine)"]
    C -->|1-Min Monologue| B

    B -->|Async Queue| D["🤗 Cloud Neural Worker\n(F5-TTS Voice Model)"]
    D -->|Audio Synthesis| D

    D -->|Direct Delivery| T["📲 Your Telegram Chat\n(Mastered Audio Message)"]
```

1. **The Spark:** Triggered automatically twice a day (06:00 & 18:00 WIB) or instantly when you text the bot.
1. **The Mind:** Gemini Flash crafts a tightly calibrated ~120-word spoken monologue addressing modern friction.
1. **The Voice:** Hugging Face Spaces synthesizes speech via neural diffusion and optimizes loudness for mobile speakers.
1. **The Delivery:** The master audio file and telemetry arrive directly in your private Telegram feed.

---

## 📱 The Experience

Every delivery arrives cleanly formatted with actionable metadata:

```text
🎙️ Personal OS — Motivation Engine

📌 Topic: The Algorithmic Echo of Validation
📝 Summary: True autonomy is found by severing the feedback loop between your worth and digital metrics.
⏱️ Process Time: 48.2s
🪙 Tokens: 512 (120 in / 392 out)

"Hello, friend..."
▶️ [━━━━━━━●────────] 01:04 (High-Fidelity Audio)
```

### Interactive Commands

| Command / Input        | What Happens                                                                             |
| :--------------------- | :--------------------------------------------------------------------------------------- |
| `/generate`            | Generates a fresh audio on a novel, dynamically generated stoic topic.                   |
| `/generate <topic>`    | Generates a targeted audio around your chosen topic (e.g., `/generate burnout`).         |
| _`<any text message>`_ | Send any raw thought or vent; the engine deconstructs it into a 1-minute audio response. |
| `/status`              | Checks engine uptime, AI health, and schedule information.                               |

---


## ⏰ Delivery Protocol

| Cadence | Jakarta Time (WIB) | Mindset Objective |
| :--- | :--- | :--- |
| **Morning Protocol** | **06:00 WIB** | Calibrate mental clarity and strip manufactured urgency before the workday. |
| **Evening Reflection** | **18:00 WIB** | Mental debrief, disconnect from the loop, and reclaim quiet autonomy. |

  _(Need a mental reset midday? Just send `/generate` anytime)._

---

## 📚 Technical Documentation

For engineers interested in the implementation details, deep-dive into our modular architecture and code:

- 🏛️ **[Architecture & Data Flow](docs/ARCHITECTURE.md)** — In-depth breakdown of Go concurrency, FFmpeg mastering, and async queues.
- 📋 **[Product Requirements (PRD)](docs/PRD.md)** — Acoustic guidelines, word pacing, and target specifications.
- 🗺️ **[Implementation Roadmap](docs/IMPLEMENTATION_PLAN.md)** — Completed phases, milestones, and system checklist.
- 📜 **[Version Changelog](docs/CHANGELOG.md)** — Semantic version history and release notes.

---

## ⚖️ Ethical Disclaimer

  _Personal OS is an experimental, non-commercial educational research project exploring zero-shot neural voice diffusion, LLM reasoning, and serverless architectures. The voice persona is inspired by the fictional character Elliot Alderson from Mr. Robot._

---

## 📄 License

Distributed under the [MIT License](LICENSE). © 2026 Personal OS.
