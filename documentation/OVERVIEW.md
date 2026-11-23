# Project Overview: Personal OS (AI Audio Motivation Engine)

## 1. Executive Summary

**Personal OS** is an automated, hybrid-architecture multimedia pipeline designed to generate long-form, high-immersion motivational audio. By combining the raw, psychological intensity of characters like Elliot Anderson (_Mr. Robot_) or Batman with ASMR auditory textures, the system produces long-duration motivational soundscapes.

The project uses a **hybrid architecture**—leveraging **Golang** for high-performance backend orchestration, concurrency, and media post-processing, alongside **Python** as a specialized microservice for running GPU-bound voice-cloning models.

---

## 2. Core Objectives

- **Character Voice Synthesis:** Zero-shot voice cloning utilizing clean reference audio of specific personas (e.g., whispering, monotone, gritty timber).
- **Long-Form Stability:** Overcoming LLM and TTS length limitations by implementing intelligent text chunking and seamless audio stitching.
- **ASMR Immersion:** Post-processing raw audio with spatial pacing, breath pauses, and background noise textures (e.g., brown noise, rain, subtle room tone).
- **System Efficiency:** Low-overhead orchestration capable of managing background rendering jobs without choking system resources.

---

## 3. System Architecture & Tech Stack

```text
[ Golang Orchestrator (Backend / CLI / API) ]
  ├── 1. Scheduler / Trigger Engine
  ├── 2. LLM Prompt Engine (Script Generation)
  ├── 3. Text Chunking & Concurrency Controller (Goroutines)
  └── 4. FFmpeg Audio Post-Processor (Stitching & ASMR Mixing)
         ▲
         │ (HTTP / gRPC via JSON)
         ▼
[ Python Inference Worker (FastAPI) ]
  ├── PyTorch Runtime
  └── F5-TTS / Fish Speech Voice Cloning Model (GPU VRAM)
```

### Component Breakdown

| Layer                     | Technology                    | Responsibilities                                                                                                   |
| :------------------------ | :---------------------------- | :----------------------------------------------------------------------------------------------------------------- |
| **Orchestration & Logic** | **Golang (Go)**               | Task scheduling, script generation via LLM APIs, text splitting, concurrent dispatch, and FFmpeg execution.        |
| **Inference Engine**      | **Python, FastAPI, PyTorch**  | Hosting the zero-shot voice cloning model (`F5-TTS` or `Fish Speech`) to convert text fragments into audio chunks. |
| **Post-Processing**       | **FFmpeg (via Go `os/exec`)** | Audio chunk concatenation, crossfading, loudness normalization (`loudnorm`), and ambient noise layering.           |

---

## 4. End-to-End Execution Pipeline

1. **Trigger & Scripting:**
   - A cron job or manual trigger starts the Go application.
   - Go queries an LLM (e.g., Llama 3 via Ollama or DeepSeek API) using a persona-engineered system prompt containing acoustic cues (`[sigh]`, `...`, `[whisper]`).
2. **Text Chunking:**
   - The generated long script is split into manageable fragments (150–250 characters) to prevent TTS hallucinations and intonation drift.
3. **Parallel Inference:**
   - Go fires concurrent requests (`goroutines`) to the Python FastAPI worker.
   - The Python worker processes each chunk through the voice-cloning model and returns raw `.wav` files.
4. **Audio Assembly & Mastering:**
   - Go collects the audio chunks and passes them to FFmpeg.
   - FFmpeg applies crossfades, mixes in a subtle background _brown noise_ track, and normalizes audio levels for an authentic ASMR experience.

---

## 5. Risk Assessment & Mitigations

- **Vulnerability: Acoustic Degradation on Long Audio**
  - _Cause:_ Forcing TTS to render 30 minutes in one pass leads to robotic pitch shifts.
  - _Mitigation:_ Strict sentence/paragraph chunking followed by millisecond-accurate crossfading in FFmpeg.
- **Vulnerability: Garbage In, Garbage Out (Reference Audio)**
  - _Cause:_ Using movie clips containing background music or sound effects ruins voice cloning fidelity.
  - _Mitigation:_ Pre-process reference audio using source separation tools (UVR5 / Demucs) to isolate clean vocals.
- **Vulnerability: Cross-Lingual Prosody Mismatch**
  - _Cause:_ English-trained character voices speaking Indonesian can sound unnatural or misaccented.
  - _Mitigation (Approach A Adopted):_ Standardize on **Fish Speech** as the primary speech microservice. Use clean 10–15s English reference vocal audio from target characters (Elliot Alderson, Batman) with their corresponding English transcripts to condition the model, then synthesize introspective Indonesian scripts via cross-lingual zero-shot voice cloning.

---

## 6. Development Milestones

- **Phase 1 (Validation):** Extract clean voice samples (10–15s) of the target character and test local voice cloning output quality using F5-TTS in Python.
- **Phase 2 (Core Engine):** Build the Golang script chunking engine and establish communication with the Python FastAPI worker.
- **Phase 3 (Post-Processing):** Implement the FFmpeg pipeline in Go for crossfading and ASMR sound layering.
- **Phase 4 (Automation):** Wrap the system into a scheduled CLI or background service for daily automated generation.
