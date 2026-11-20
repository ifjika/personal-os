# Changelog
All notable changes to the **Personal OS: AI Audio Motivation Engine** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
