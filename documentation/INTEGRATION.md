# Integration Specification & API Contracts
## Personal OS: AI Audio Motivation Engine
**Version:** `v0.2.1-alpha` | **Last Updated:** `2026-11-20`

---

## 1. Cloud TTS Inference API Contract

### 1.1 Endpoint Overview
The 24/7 cloud voice cloning worker deployed on Hugging Face Spaces exposes a unified REST API endpoint for raw audio synthesis.

```text
POST https://<space-domain>.hf.space/v1/tts
```

- **Authentication:** Unauthenticated (Public Space) or Bearer Token (`Authorization: Bearer <HF_TOKEN>`)
- **Transport:** HTTPS / TCP
- **Timeout:** 180 seconds (accommodates cold starts and multi-step flow matching inference)

### 1.2 Request Payload Schema (`application/json`)

```json
{
  "text": "Hello, friend. Focus only on the next honest move right before your eyes."
}
```

#### Field Specifications:
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `text` | `string` | **Yes** | Speech text to synthesize (max 250 characters per chunk). |
| `reference_audio` | `string` | No | Optional override path for custom acoustic prompt audio. |
| `reference_text` | `string` | No | Optional transcript matching the reference audio. |

### 1.3 Response Specification
- **Success Status:** `200 OK`
- **Content-Type:** `audio/wav`
- **Body:** Binary uncompressed PCM WAV stream (48,000 Hz, 16-bit, mono).
- **Streaming Behavior:** Streamed directly to disk cache using `io.Copy` to prevent heap buffer bloat.

### 1.4 Error Status Codes
| HTTP Status | Error Code | Description | Mitigation Strategy |
| :--- | :--- | :--- | :--- |
| `400` | `ERR_EMPTY_TEXT` | Input text string is blank. | Client validates text length prior to dispatch. |
| `404` | `ERR_REF_NOT_FOUND`| Reference audio missing on cloud disk. | Verify `elliot_ref_15s.wav` exists in Space root. |
| `500` | `ERR_CUDA_NVML` | CUDA initialization conflict. | Ensure `import spaces` is placed at Line 1 of `app.py`. |
| `503` | `ERR_CONTAINER_START` | Space is waking up from idle. | Client implements exponential backoff retry. |

---

## 2. Local Fallback Contract (`EdgeTTSSynthesizer`)

When `--tts-url` is omitted, the CLI orchestrator falls back to local Microsoft Edge neural synthesis:

```bash
python -m edge_tts --voice=en-US-ChristopherNeural --rate=-8% --text="<CHUNK_TEXT>" --write-media="<OUTPUT_PATH>"
```

- **Voice Identifier:** `en-US-ChristopherNeural`
- **Tempo Modulation:** `-8%` (emulates Elliot Alderson's calm, deliberate cadence)
- **Dependency:** `pip install edge-tts`
- **Hardware Requirement:** Zero GPU (pure CPU/network fallback).

---

## 3. Google Gemini API Dynamic Monologue Contract

For dynamic, unscripted motivation audio, the orchestrator invokes the Google Gemini REST API.

```text
POST https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent?key=<GEMINI_API_KEY>
```

- **Fallback Model:** `gemini-flash-latest` (if 404 occurs on preview endpoints).
- **Thinking Configuration:** Explicitly disabled (`thinkingBudget: 0`) to prevent internal reasoning tokens from depleting output limits.
- **Output Token Target:** `maxOutputTokens: 2500` (allowing ~300 to 450 spoken words).
- **Telemetry (`usageMetadata`):**
  - `promptTokenCount`: ~350–380 tokens (Elliot system instruction + topic prompt).
  - `candidatesTokenCount`: ~440–470 tokens (~350 words monologue).
  - `totalTokenCount`: ~800–830 tokens (~$0.0002 or Free Tier $0).

---

## 4. Hugging Face Gradio 6 SSE Queue Protocol

When connecting to Gradio Spaces, the client uses the native 3-step SSE Queue pipeline:
1. **Initiate Job:** `POST https://<space-domain>.hf.space/gradio_api/call/synthesize_voice` with `{"data": ["<chunk_text>"]}`. Returns `{"event_id": "<uuid>"}`.
2. **Listen to SSE:** `GET https://<space-domain>.hf.space/gradio_api/call/synthesize_voice/<event_id>`. Listens for `data: [{"url": "..."}]`.
3. **Download WAV:** `GET <audio_url>` and pipe bytes directly to `storage/cache/chunk_XX.wav`.
