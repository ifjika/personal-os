# Changelog

All notable changes to the **Personal OS: AI Audio Motivation Engine** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v0.4.0] - 2026-11-21 (Live Production Release)

### Added

- **24/7 Automated $0 Cloud Pipeline (Cloudflare Cron + Hugging Face F5-TTS + Telegram):**
  - **Full Automation Schedule:** Cloudflare Workers Cron Trigger firing every 3 hours WIB (`0 2,5,8,11,14,17,20,23 * * *`), broadcasting at 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, and 21:00 WIB.
  - **Asynchronous Cloud Queue Integration:** Solved the Cloudflare Workers Free Tier 30-second execution kill by dispatching jobs to Hugging Face Spaces' native Gradio Queue (`POST /gradio_api/call/synthesize_voice`) in <200ms. Hugging Face synthesizes the full 1-minute audio and dispatches directly to Telegram Bot API with zero timeout constraints.
  - **Interactive Telegram Bot Webhook:** Integrated `/webhook` receiver enabling on-demand audio generation via `/generate` or custom topic input (`/generate <topic>`).
  - **Native Telegram Command Menu:** Registered Telegram `setMyCommands` and `setChatMenuButton`, rendering native auto-complete recommendations and the blue `[Menu]` button.
- **100% Dynamic Topic Generation (Eliminated Static Lists):**
  - Completely purged static `DEFAULT_TOPICS` hardcoded array. Google Gemini (`gemini-3-flash-preview`) now invents novel, unpredictable daily friction topics correlated with resisting the societal matrix and treating life as an open-world simulation.
  - Contextual Adaptation: Prompts dynamically adapt to user-provided input (e.g. `/generate lelah dengan rutinitas`) to tailor philosophical stoic solutions specifically around user friction.
- **Process Time & Token Telemetry in Telegram Captions:**
  - Real-time synthesis elapsed timer (`Process Time: <time>s`).
  - Gemini API token tracking (`Tokens: <total> (<prompt> in / <candidate> out)`).
  - Clean minimalist layout displaying solely **Topic**, **1-sentence Executive Summary**, **Process Time**, and **Tokens**.
- **Acoustic Loudness & Synthesis Acceleration:**
  - Accelerated F5-TTS inference by setting `nfe_step=24` (cutting CPU synthesis latency by ~40% with zero loss of whisper fidelity).
  - Integrated peak amplitude gain scaling (`min(3.5, 0.94 / max_val)`) in `app.py` boosting quiet speech up to +9.8 dB for clear phone speaker listening.

---

## [v0.3.5-alpha] - 2026-11-20

### Added

- **100% Dynamic Topic Generation (Eliminated Static Lists):**
  - Removed static `DEFAULT_TOPICS` hardcoded array. Gemini now generates novel, unpredictable daily friction topics correlated with the system matrix and game of life on every trigger.
  - Contextual Adaptation: If the user provides specific context (e.g. `/generate lelah dengan rutinitas`), Gemini tailors the topic and stoic philosophy specifically around that friction.
- **Process Time & Token Metrics in Telegram Captions:**
  - Added wall-clock synthesis timer (`Process Time`) and Gemini API token consumption tracking (`Tokens: total (prompt in / candidate out)`) to the delivery caption.
- **Audio Synthesis Acceleration & Whisper Loudness Scaling:**
  - Calibrated F5-TTS sampling to `nfe_step=24` (cutting CPU inference latency by ~40% with zero loss of whisper fidelity).
  - Integrated peak amplitude gain scaling (`min(3.5, 0.94 / max_val)`) directly in `app.py` for loud, crisp playback on phone speakers.

---

## [v0.3.4-alpha] - 2026-11-20

### Added

- **Asynchronous Cloud Task Delegation (`deploy/hf-space/app.py` & `deploy/cloudflare-cron`):**
  - Added background execution endpoint `POST /v1/trigger` on Hugging Face Spaces using FastAPI `BackgroundTasks`, decoupling heavy CPU speech synthesis (~60s) from serverless execution limits.
  - Resolved Cloudflare Workers Free Tier 30-second execution kill where prolonged F5-TTS inference aborted audio delivery silently.
  - Updated Cloudflare Worker to delegate on-demand generation in <300ms, letting Hugging Face Spaces finish neural synthesis and dispatch audio directly to Telegram without timeout risk.

---

## [v0.3.3-alpha] - 2026-11-20

### Added

- **Native Telegram Bot Autocomplete Menu (`setMyCommands`):**
  - Integrated Telegram `setMyCommands` API registration, enabling native in-app command recommendations and descriptions when typing `/` or `/ge`.
  - Bundled command menu registration into the 1-Click `/set-webhook` endpoint.
  - Activated Telegram Webhook link connecting Cloudflare Worker to Telegram Bot API.

---

## [v0.3.2-alpha] - 2026-11-20

### Added

- **Interactive Telegram Bot Webhook & On-Demand Commands (`deploy/cloudflare-cron`):**
  - Integrated native Telegram Webhook receiver (`POST /webhook`) allowing direct on-demand generation from smartphone or desktop chat.
  - Added command `/generate` for random stoic monologues and `/generate <topic>` for custom user themes.
  - Supported free-form text input: sending any message to the bot automatically converts it into a targeted philosophical motivation topic.
  - Added immediate user acknowledgment (`sendTelegramText`) and realistic "recording voice..." status animation (`sendChatAction: record_voice`) during synthesis.
  - Added 1-Click Webhook registration route (`GET /set-webhook`) to automatically connect Telegram to Cloudflare.
  - Implemented owner authorization gate preventing unauthorized third parties from exhausting Gemini and TTS compute quotas.

---

## [v0.3.1-alpha] - 2026-11-20

### Added

- **Mobile Audio Peak Loudness Normalization & Gain Boost:**
  - Implemented dynamic digital peak normalization in Cloudflare Worker (`deploy/cloudflare-cron/src/index.js`), scaling quiet F5-TTS whisper samples up to 3.5x (+9.8 dB) to target -0.5 dBFS (30,800/32,767 PCM peak), ensuring punchy mobile speaker volume.
  - Upgraded Go FFmpeg mastering pipeline (`internal/audio/mixer.go`) with `volume=1.8` pre-gain and mobile podcast broadcast standard **-14 LUFS** (`loudnorm=I=-14:LRA=7:TP=-1.0`).
- **Structured Executive Summary & Minimalist Telegram Caption:**
  - Stripped non-essential metadata (`Cadence`, `Persona`) from Telegram delivery captions.
  - Configured Gemini API to return structured JSON containing a 1-sentence executive takeaway (`summary`) and the spoken script (`monologue`).
  - Telegram delivery now displays clean **Topic** and **Summary**:

    ```text
    🎙️ Personal OS — Motivation Engine

    📌 Topic: <Topic>
    📝 Summary: <Executive Summary>

    "Hello, friend..."
    ```

---

## [v0.3.0-alpha] - 2026-11-20

### Added

- **Telegram Bot Audio Dispatcher (`internal/dispatcher`):**
  - Implemented `TelegramDispatcher` streaming master audio directly to Telegram mobile/desktop via `multipart/form-data` to Telegram Bot API (`POST /bot<TOKEN>/sendAudio`).
  - Wired into CLI with `--telegram`, `--tg-token`, and `--tg-chat-id` flags and zero-config `.env` loading (`TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`).
  - Added table-driven unit test suite (`telegram_test.go`) validating mock payload parsing and parameter contracts.
- **24/7 Cloudflare Cron Trigger Worker (`deploy/cloudflare-cron`):**
  - Developed autonomous serverless cron trigger worker scheduled for every 2 hours (`0 */2 * * *`).
  - Embedded complete 24/7 cloud orchestration pipeline: dynamic stoic topic selection, Gemini Flash monologue scripting (110–130 words), acoustic chunking, Hugging Face Spaces F5-TTS neural synthesis, native 16-bit PCM equal-power quarter-sine WAV stitching, and Telegram delivery.
  - Implemented `GET /trigger` route for immediate on-demand manual verification.
  - Added comprehensive deployment guide in `deploy/cloudflare-cron/README.md`.

---

## [v0.2.3-alpha] - 2026-11-20

### Added

- **Job Completion Telemetry Summary (`cmd/cli/main.go`):**
  - Integrated high-resolution process stopwatch (`time.Since(startTime)`) reporting exact wall-clock rendering latency upon completion (`Total Process Time`).
  - Added comprehensive token expenditure audit to final completion banner (`Gemini Tokens Spent : X tokens (Y prompt / Z output)`).
  - Maintained zero-crash graceful fallback for offline builtin template executions (`Offline Builtin template used`).

---

## [v0.2.2-alpha] - 2026-11-20

### Added

- **1-Minute Precision Pacing Engine (`internal/prompt`):**
  - Calibrated script generation constraints to strictly 110–130 words (~700–850 characters), precisely targeting a 60-second total runtime at Elliot Alderson's calm, deliberate 115–125 WPM whisper cadence.
  - Reduced token overhead to `maxOutputTokens: 800`, speeding up script generation by ~300% and halving API token costs.
- **Equal-Power Acoustic Crossfade & Transition Smoothing (`internal/audio`):**
  - Replaced linear triangle crossfade (`tri`) with constant-power quarter-sine curve (`c1=qsin:c2=qsin`), eliminating the -3dB volume dip and energy drop between stitched vocal fragments.
  - Extended transition overlap from 100ms to 250ms (`d=0.25`), naturally blending room acoustic tails, breath pauses, and ambient floor noise.
  - Optimized chunk boundary sizing in CLI (`NewBoundaryChunker(120, 280)`), compressing 1-minute monologues into just 3 coherent acoustic units instead of 11 micro-fragments (reducing splice seams by ~73%).

---

## [v0.2.1-alpha] - 2026-11-20

### Added

- **Gemini API Dynamic Philosophical Script Generator (`internal/prompt`):**
  - Integrated Google Gemini (`gemini-3-flash-preview` / `gemini-flash-latest`) for fully dynamic, topic-driven monologues.
  - Injected Elliot Alderson persona lore: deconstructing the societal matrix, life as a rigged simulation/game, and hacking personal programming through stoic execution.
  - Added real-time token telemetry (`UsageMetadata`) tracking input, output, and total token usage on each generation.
  - Implemented zero-dependency `.env` environment loader in CLI entrypoint for secure `GEMINI_API_KEY` and `TTS_URL` loading.
- **Gradio 6 SSE Queue Protocol Support (`internal/synthesizer`):**
  - Upgraded `HTTPSynthesizer` to automatically fallback to Hugging Face Spaces Gradio API (`POST /gradio_api/call/synthesize_voice` -> SSE stream event listener -> WAV downloader) when direct `/v1/tts` endpoint is unavailable.

### Fixed

- **Gemini 1.5 Sunset / 404 Model Resolution:**
  - Updated model routing from deprecated `gemini-1.5-flash` to `gemini-3-flash-preview` and `gemini-flash-latest`.
- **Reasoning Token Budget Monologue Cutoff (9-Second Audio Bug):**
  - Resolved `finishReason: MAX_TOKENS` cutoff where Gemini 3 reasoning tokens consumed the output budget. Set `thinkingConfig.thinkingBudget: 0` and expanded `maxOutputTokens: 2500`, successfully producing full 350-word scripts (~2.5 to 3 minutes of mastered audio).

---

## [v0.2.0-alpha] - 2026-11-20

### Added

- **24/7 Cloud Zero-Shot Neural Voice Worker (`deploy/hf-space`):**
  - Containerized F5-TTS model on Hugging Face Spaces (CPU 16GB RAM) running 24/7 at $0 cost.
  - Implemented Dual Interface: Gradio web GUI for manual browser playback + FastAPI REST endpoint (`POST /v1/tts`) for automated orchestrator calls.
  - Embedded clean reference vocal of Elliot Alderson (`elliot_ref_15s.wav`) and fixed reference transcript.
- **Go Orchestrator Scaffolding & CLI (`cmd/cli/main.go`):**
  - Built zero-trust argument dispatcher supporting `--topic`, `--persona`, `--tts-url`, `--ambient`, and `--output`.
- **Acoustic Boundary Chunker Engine (`internal/chunker`):**
  - Developed boundary-aware sentence splitter (100–250 chars) preventing neural TTS hallucinations.
  - Added comprehensive table-driven unit tests (`chunker_test.go`) achieving 100% test pass.
- **Daily Monologue Prompt Generator (`internal/prompt`):**
  - Implemented `GeminiScriptGenerator` utilizing Google Gemini Flash API for dynamic, contextual monologue synthesis.
  - Formulated system prompt encoding Elliot Alderson's lore: seeing through the system's manufactured loops, treating reality as a game/simulation, and reclaiming agency through calculated stoic action.
  - Sanitized generated text by stripping all markdown noise (`**`, `##`, `_`) to guarantee pristine TTS pronunciation.
  - Maintained `BuiltinScriptGenerator` as deterministic offline fallback.
- **Dual-Mode Voice Synthesizer Client (`internal/synthesizer`):**
  - Implemented `HTTPSynthesizer` with direct socket-to-disk streaming (`io.Copy`) to prevent heap fragmentation.
  - Implemented `EdgeTTSSynthesizer` (`en-US-ChristopherNeural`, `-8%` rate) as local zero-GPU fallback.
  - Implemented `MockSynthesizer` for offline deterministic testing with 432 Hz test tones.
- **FFmpeg Mastering & Crossfade Engine (`internal/audio`):**
  - Dynamic `filter_complex` builder for 100ms linear crossfade (`acrossfade=d=0.1:c1=tri:c2=tri`).
  - Broadcast loudness compliance: EBU R128 normalization (`loudnorm=I=-16:LRA=7:TP=-1.5`).
  - Multi-track ASMR ambient soundscape layering (`volume=0.12` / -18 dB ducking).
  - High-fidelity 320kbps MP3 encoding (`libmp3lame`).
  - Verified with synthetic sine wave test suite (`mixer_test.go`).

### Fixed

- **PyArrow & NumPy Dependency Clash in Hugging Face (`ModuleNotFoundError`):**
  - Removed explicit `numpy<=1.26.4` and duplicate `torch` from `requirements.txt` to eliminate C-extension incompatibility with Python 3.12 and PyArrow.
- **Port 7860 Address Collision on Hugging Face (`[Errno 98] address already in use`):**
  - Removed duplicate `uvicorn.run()` call; switched to native `demo.launch()` with custom FastAPI routes mounted directly on `demo.app`.
- **CUDA Lifecycle Exception in ZeroGPU (`RuntimeError: CUDA has been initialized before importing spaces`):**
  - Guarded and hoisted `import spaces` to Line 1 of `app.py` before any PyTorch or CUDA submodules load.
- **Python `argparse` Negative Number Parsing (`error: argument --rate: expected one argument`):**
  - Reformatted CLI argument passing from `"--rate", "-8%"` to `"--rate=-8%"` to prevent minus signs from being misinterpreted as flags.
- **Indonesian Phoneme Distortion in Foreign Neural TTS:**
  - Standardized target generation language to English (`en`) for Elliot Alderson persona, completely eliminating accented/slurred speech artifacts.

---

## [v0.1.0-alpha] - 2026-11-16

### Added

- **Acoustic Profiling & Persona Selection:**
  - Tested Elliot Alderson (Mr. Robot ASMR) and Batman Indonesian narrator voice samples.
  - Extracted and isolated clean 15s studio snippet: `storage/references/processed/elliot_ref_15s.wav`.
  - Identified phonetic limitations of multilingual zero-shot models and finalized Elliot Alderson as the primary vocal identity.
- **Core Engineering Documentation:**
  - Created initial suite of 9 technical specifications in `docs/`: `PRD.md`, `ARCHITECTURE.md`, `TSD.md`, `INTEGRATION.md`, `APP_FLOW.md`, `IMPLEMENTATION_PLAN.md`, `TESTING_GUIDE.md`, `DESIGN_SYSTEM.md`, and `SECURITY_GUIDE.md`.
