# Implementation Plan & Execution Roadmap

## Personal OS: AI Audio Motivation Engine

**Current Version:** `v0.2.3-alpha`  
**Last Updated:** `2026-11-20`

> **Operational Directive:**
> Use the active checklist below to track features currently under development. In accordance with the protocol, **completed and verified features are eliminated from the active checklist** to ensure the document remains concise, clean, and exclusively reflective of active work-in-progress (WIP).

---

## 1. Phase-Based Roadmap & Milestone Status

| Phase       | Core Focus                             | Target Deliverables                                                                                    | Status           | Version        |
| :---------- | :------------------------------------- | :----------------------------------------------------------------------------------------------------- | :--------------- | :------------- |
| **Phase 1** | **Acoustic & Voice Validation**        | Clean vocal reference isolation, Elliot Alderson voice persona lock, target language alignment (`en`). | `COMPLETED` ✅   | `v0.1.0-alpha` |
| **Phase 2** | **Core Orchestrator & Chunker**        | Go domain boundaries, boundary chunker (100–250 chars), prompt generator, SQLite storage.              | `IN_PROGRESS` ⏳ | `v0.2.0-alpha` |
| **Phase 3** | **Audio Post-Processing & TTS Client** | FFmpeg 100ms crossfade, EBU R128 mastering, Hugging Face 24/7 F5-TTS cloud deployment.                 | `COMPLETED` ✅   | `v0.2.0-alpha` |
| **Phase 4** | **Web Studio & Telemetry**             | Terminal Noir UI (Next.js/Tailwind), canvas waveform visualizer, SSE log streamer.                     | `PENDING` 📋     | `v0.3.0-alpha` |
| **Phase 5** | **Automation & Dispatch**              | Scheduled cron trigger, Telegram Bot audio dispatcher, Docker containerization.                        | `PENDING` 📋     | `v0.4.0-alpha` |

---

## 2. Active Feature Checklist (Work-in-Progress)

Items currently in progress or queued for immediate implementation _(Completed items have been purged per protocol)_:

- [ ] **[Phase 2.4]** Implement SQLite persistence layer (`internal/repository`) with WAL mode (`PRAGMA journal_mode=WAL;`) for `personas`, `render_jobs`, and `job_chunks`.
- [ ] **[Phase 4.1]** Scaffold Web Studio frontend (`apps/web-console`) using Next.js App Router and _Terminal Noir_ Tailwind design tokens.
- [ ] **[Phase 4.2]** Develop interactive HTML5 Canvas Waveform Visualizer & Audio Scrubber component.
- [ ] **[Phase 4.3]** Implement Server-Sent Events (SSE) hook (`useJobLogs`) for real-time log streaming from Go orchestrator to Web Console.
- [ ] **[Phase 5.1]** Build Telegram Bot Dispatcher (`internal/dispatcher`) to automatically deliver mastered MP3 audio to mobile devices.
- [ ] **[Phase 5.2]** Set up 24/7 cloud cron trigger (Cloudflare Worker Cron / GitHub Actions) to run daily generation cycles automatically.

---

## 3. Verified & Completed Milestones (Archive)

The following components have been implemented, tested, and verified:

1. **Vocal Persona Lock (`internal/domain/persona.go`):**
   - Locked to **Elliot Alderson** (`storage/references/processed/elliot_ref_15s.wav`).
   - Standardized target language to English (`en`) for native phoneme fluidity and zero-hallucination cadence.
2. **Boundary-Aware Acoustic Chunker (`internal/chunker`):**
   - Implemented regex-based clause boundary chunker (`100–250 characters`).
   - Verified 100% PASS with table-driven tests (`chunker_test.go`).
3. **Dynamic Daily Monologue Generator (`internal/prompt`):**
   - Implemented `GeminiScriptGenerator` querying Google Gemini API (`gemini-3-flash-preview` / `gemini-flash-latest`) with custom system prompt rules: societal systemic loops, reality as a game, and stoic mental control.
   - Fixed reasoning budget exhaustion bug by setting `thinkingBudget: 0` and `maxOutputTokens: 2500` to prevent monologue cutoff.
   - Added real-time token telemetry (`UsageMetadata`) logging prompt, candidate, and total token counts.
   - Built automatic markdown scrubbing pipeline to deliver pure speech text to TTS.
   - Maintained `BuiltinScriptGenerator` as deterministic offline fallback.
4. **FFmpeg Post-Processor (`internal/audio`):**
   - Implemented dynamic `filter_complex` builder for 100ms linear crossfade (`acrossfade=d=0.1:c1=tri:c2=tri`).
   - Implemented multi-track ambient layer mixing (`amix=inputs=2:duration=first`) with -18 dB ducking.
   - Enforced broadcast loudness compliance: **EBU R128 (`loudnorm=I=-16:LRA=7:TP=-1.5`)** and master MP3 320kbps encoding.
   - Verified 100% PASS with synthetic sine wave test suite (`mixer_test.go`).
5. **Dual-Mode Synthesizer Client (`internal/synthesizer`):**
   - Implemented `HTTPSynthesizer` for streaming cloud WAV audio directly to disk (`io.Copy`).
   - Implemented `EdgeTTSSynthesizer` (local zero-GPU fallback via `python -m edge_tts`).
   - Implemented `MockSynthesizer` for deterministic offline development.
6. **24/7 Cloud Zero-Shot Neural TTS Worker (`deploy/hf-space`):**
   - Successfully deployed to Hugging Face Spaces (CPU 16GB RAM) running **F5-TTS**.
   - Dual interface: Web GUI (Gradio) + REST API endpoint (`POST /v1/tts`).
