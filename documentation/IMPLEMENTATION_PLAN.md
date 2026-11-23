# Implementation Plan & Execution Roadmap

## Personal OS: AI Audio Motivation Engine

**Current Version:** `v0.4.0` (Live Production)  
**Last Updated:** `2026-11-21`

> **Operational Directive:**
> Use the active checklist below to track features currently under development. In accordance with the protocol, **completed and verified features are eliminated from the active checklist** to ensure the document remains concise, clean, and exclusively reflective of active work-in-progress (WIP).

---

## 1. Phase-Based Roadmap & Milestone Status

| Phase       | Core Focus                             | Target Deliverables                                                                                    | Status           | Version        |
| :---------- | :------------------------------------- | :----------------------------------------------------------------------------------------------------- | :--------------- | :------------- |
| **Phase 1** | **Acoustic & Voice Validation**        | Clean vocal reference isolation, Elliot Alderson voice persona lock, target language alignment (`en`). | `COMPLETED` ✅   | `v0.1.0-alpha` |
| **Phase 2** | **Core Orchestrator & Chunker**        | Go domain boundaries, boundary chunker (120–280 chars), prompt generator, SQLite storage.              | `IN_PROGRESS` ⏳ | `v0.2.0-alpha` |
| **Phase 3** | **Audio Post-Processing & TTS Client** | FFmpeg 250ms crossfade, EBU R128 mastering, Hugging Face 24/7 F5-TTS cloud deployment.                 | `COMPLETED` ✅   | `v0.2.0-alpha` |
| **Phase 4** | **Web Studio & Telemetry**             | Terminal Noir UI (Next.js/Tailwind), canvas waveform visualizer, SSE log streamer.                     | `PENDING` 📋     | `v0.5.0-alpha` |
| **Phase 5** | **Automation & Dispatch**              | 6-hourly Cloudflare Cron trigger, Telegram Bot audio dispatcher, mobile delivery, dynamic topics.        | `COMPLETED` ✅   | `v0.4.0` |

---

## 2. Active Feature Checklist (Work-in-Progress)

Items currently in progress or queued for immediate implementation _(Completed items have been purged per protocol)_:

- [ ] **[Phase 2.4]** Implement SQLite persistence layer (`internal/repository`) with WAL mode (`PRAGMA journal_mode=WAL;`) for `personas`, `render_jobs`, and `job_chunks`.
- [ ] **[Phase 4.1]** Scaffold Web Studio frontend (`apps/web-console`) using Next.js App Router and _Terminal Noir_ Tailwind design tokens.
- [ ] **[Phase 4.2]** Develop interactive HTML5 Canvas Waveform Visualizer & Audio Scrubber component.
- [ ] **[Phase 4.3]** Implement Server-Sent Events (SSE) hook (`useJobLogs`) for real-time log streaming from Go orchestrator to Web Console.

---

## 3. Verified & Completed Milestones (Archive)

The following components have been implemented, tested, and verified:

1. **Vocal Persona Lock (`internal/domain/persona.go`):**
   - Locked to **Elliot Alderson** (`storage/references/processed/elliot_ref_15s.wav`).
   - Standardized target language to English (`en`) for native phoneme fluidity and zero-hallucination cadence.
2. **Boundary-Aware Acoustic Chunker (`internal/chunker`):**
   - Implemented regex-based clause boundary chunker (`120–280 characters`).
   - Verified 100% PASS with table-driven tests (`chunker_test.go`).
3. **Dynamic Daily Monologue Generator (`internal/prompt`):**
   - Implemented `GeminiScriptGenerator` querying Google Gemini API (`gemini-3-flash-preview` / `gemini-flash-latest`) with custom system prompt rules: societal systemic loops, reality as an open-world simulation/game, and stoic mental control.
   - Fixed reasoning budget exhaustion bug by setting `thinkingBudget: 0` and `maxOutputTokens: 950` for 1-minute audio (~110–130 words).
   - Added real-time token telemetry (`UsageMetadata`) tracking prompt, candidate, and total token counts.
   - Built automatic markdown scrubbing pipeline to deliver pure speech text to TTS.
   - Maintained `BuiltinScriptGenerator` as deterministic offline fallback.
4. **FFmpeg Post-Processor (`internal/audio`):**
   - Implemented dynamic `filter_complex` builder for 250ms equal-power quarter-sine crossfade (`acrossfade=d=0.25:c1=qsin:c2=qsin`).
   - Implemented multi-track ambient layer mixing (`amix=inputs=2:duration=first`) with -18 dB ducking.
   - Enforced mobile podcast broadcast loudness compliance: **-14 LUFS** (`loudnorm=I=-14:LRA=7:TP=-1.0`) with `volume=1.8` pre-gain and master MP3 320kbps encoding.
   - Verified 100% PASS with synthetic sine wave test suite (`mixer_test.go`).
5. **Dual-Mode Synthesizer Client (`internal/synthesizer`):**
   - Implemented `HTTPSynthesizer` supporting direct `/v1/tts` and Gradio Queue protocol (`POST /gradio_api/call/synthesize_voice`).
   - Implemented `EdgeTTSSynthesizer` (local zero-GPU fallback via `python -m edge_tts`).
   - Implemented `MockSynthesizer` for deterministic offline development.
6. **24/7 Cloud Zero-Shot Neural TTS Worker (`deploy/hf-space`):**
   - Deployed to Hugging Face Spaces (CPU 16GB RAM) running **F5-TTS** with native Gradio Blocks queueing (`/gradio_api/call/synthesize_voice`).
   - Integrated inference acceleration (`nfe_step=24`) cutting synthesis latency by ~40%.
   - Integrated peak volume normalization gain scaling (`min(3.5, 0.94 / max_val)`) for clear phone speaker listening.
   - Added direct Telegram Bot API dispatch bypassing third-party serverless timeouts.
7. **Telegram Bot Dispatcher (`internal/dispatcher`):**
   - Implemented `TelegramDispatcher` streaming audio files via `multipart/form-data` to Telegram Bot API (`/sendAudio`).
   - Verified 100% PASS with unit test suite (`telegram_test.go`).
8. **24/7 Cloudflare Worker Cron Automation (`deploy/cloudflare-cron`):**
   - Automated serverless cron worker scheduled every 6 hours WIB (`0 5,11,17,23 * * *`) firing at 00:00, 06:00, 12:00 & 18:00 WIB (17:00, 23:00, 05:00 & 11:00 UTC).
   - Interactive Telegram bot webhook receiver (`POST /webhook`) supporting `/generate` on-demand and free-form topic input.
   - Native Telegram autocomplete command menu (`setMyCommands`) and blue `[Menu]` button registration (`/set-webhook`).
   - 100% Dynamic Topic Generation via Gemini Flash, eliminating static topic lists entirely.
   - Real-time telemetry in Telegram captions: **Topic**, **1-sentence Executive Summary**, **Process Time**, and **Token Counts**.
    - Successfully deployed to Hugging Face Spaces (CPU 16GB RAM) running **F5-TTS**.
    - Dual interface: Web GUI (Gradio) + REST API endpoint (`POST /v1/tts`).
