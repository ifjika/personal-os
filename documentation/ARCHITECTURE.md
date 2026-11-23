# Architectural Specification & System Topology

## Personal OS: AI Audio Motivation Engine

**Version:** `v0.2.0-alpha` | **Last Updated:** `2026-11-20`

---

## 1. High-Level Architectural Pattern

Personal OS employs an **Asynchronous Orchestration & Distributed Zero-Shot Inference** topology. High-concurrency script dispatching and digital signal processing (DSP) are decoupled from heavy PyTorch neural network execution.

```text
┌────────────────────────────────────────────────────────────────────────┐
│                        Client & Trigger Layer                          │
│  - Terminal CLI Operator (Go cmd/cli/main.go)                          │
│  - Scheduled Automation (Cloudflare Worker Cron / GitHub Actions)      │
│  - Mobile Delivery (Telegram Bot API)                                  │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ CLI Exec / HTTP
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Golang Orchestration Engine                          │
│  - Job Dispatcher & State Machine (internal/domain)                   │
│  - LLM Monologue Prompt Engine (internal/prompt)                       │
│  - Acoustic Boundary Chunker [120-280 chars] (internal/chunker)       │
│  - Synthesis Dispatcher [Cloud & Local] (internal/synthesizer)         │
│  - FFmpeg Audio Post-Processor & Mastering (internal/audio)            │
│  - SQLite Database [WAL Mode] (internal/repository)                   │
└──────────────────┬─────────────────────────────────┬───────────────────┘
                   │                                 │
     HTTP/REST     ▼                                 ▼ Local Subprocess Exec
┌──────────────────────────────────────┐  ┌──────────────────────────────┐
│  Hugging Face Spaces Worker (Cloud)  │  │   FFmpeg Media Processor     │
│ - 24/7 Free Container (2 vCPU, 16GB) │  │ - Multi-chunk Concatenation  │
│ - F5-TTS Zero-Shot Voice Cloning     │  │ - 250ms Equal-Power Crossfade│
│ - Elliot Alderson Vocal Persona Lock │  │ - EBU R128 Broadcast Master  │
│ - FastAPI + Gradio Port 7860         │  │ - ASMR Ambient Mixing (-18dB)│
└──────────────────────────────────────┘  └──────────────────────────────┘
```

---

## 2. Component Deep Dive

### 2.1 Go Orchestration Engine (`cmd/cli`, `internal/`)

- **Domain Modeling (`internal/domain`):** Pure Go entities defining the vocal persona (Elliot Alderson, reference audio at `storage/references/processed/elliot_ref_15s.wav`) and state machine transitions (`Pending -> GeneratingScript -> Chunking -> Synthesizing -> Mixing -> Completed`).
- **Boundary-Aware Acoustic Chunker (`internal/chunker`):** Restricts TTS input chunks to 100–250 characters, splitting strictly at natural punctuation and sentence boundaries (`. ! ? … \n`) to prevent model hallucinations and vocal glitches.
- **Synthesizer Client (`internal/synthesizer`):** Implements dual-mode abstraction:
  - `HTTPSynthesizer`: Dispatches JSON payloads to the cloud worker and streams raw WAV bytes directly to disk via `io.Copy` (Zero-Heap memory footprint).
  - `EdgeTTSSynthesizer`: Local zero-GPU fallback utilizing Microsoft Edge neural TTS (`en-US-ChristopherNeural`).
  - `MockSynthesizer`: Local offline deterministic testing with synthetic 432 Hz test tones.
- **FFmpeg Mastering Engine (`internal/audio`):** Subprocess controller generating dynamic multi-input `filter_complex` graphs for 250ms equal-power quarter-sine crossfading, ASMR ambient soundscape layering, and EBU R128 loudness compliance (-16 LUFS, 320kbps MP3).

### 2.2 Cloud Inference Worker (`deploy/hf-space/`)

- **F5-TTS Architecture:** Flow-matching zero-shot speech synthesis conditioned on Elliot Alderson's vocal timbre.
- **Runtime Environment:** Hugging Face Spaces (CPU Basic: 2 vCPU, 16 GB RAM).
- **Dual Interface:** Dual-mounted FastAPI REST endpoint (`POST /v1/tts`) + Gradio visual web player.

---

## 3. Directory Layout

```text
personal-os/
├── cmd/
│   └── cli/
│       └── main.go                   # CLI entrypoint for daily audio generation
├── deploy/
│   └── hf-space/                     # 24/7 Cloud Voice Cloning Worker
│       ├── app.py                    # FastAPI + Gradio F5-TTS application
│       ├── requirements.txt          # Python cloud dependencies
│       └── elliot_ref_15s.wav        # Studio reference audio
├── docs/                             # Engineering & Technical Specifications
│   ├── ARCHITECTURE.md               # System topology & components
│   ├── IMPLEMENTATION_PLAN.md        # Active roadmap & verified milestones
│   ├── INTEGRATION.md                # REST API contracts & schemas
│   ├── TSD.md                        # Technical Specification Document
│   └── CHANGELOG.md                  # Release version history
├── internal/
│   ├── audio/                        # FFmpeg crossfade & mastering engine
│   ├── chunker/                      # Boundary-aware acoustic text chunker
│   ├── domain/                       # Core domain entities & state machines
│   ├── prompt/                       # LLM monologue script generator
│   ├── repository/                   # SQLite persistence layer (WAL mode)
│   └── synthesizer/                  # TTS client (Cloud HTTP & Local Fallbacks)
├── storage/                          # Local Media & Database Cache
│   ├── ambient/                      # Brown noise & rain soundscapes
│   ├── cache/                        # Intermediate chunk WAV files
│   ├── output/                       # Mastered daily motivation MP3s
│   └── references/                   # Vocal persona reference files
├── go.mod
└── README.md
```
