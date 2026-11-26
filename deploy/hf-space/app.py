try:
    import spaces
except ImportError:
    pass

import os
import time
import tempfile
import urllib.request
import urllib.parse
import urllib.error
import gradio as gr
from f5_tts.api import F5TTS

# 1. Initialize F5-TTS voice model into memory
print("⏳ Loading F5-TTS voice model into memory...")
f5tts = F5TTS(device="cpu")
print("✅ F5-TTS voice model loaded successfully!")

REF_AUDIO = "elliot_ref_15s.wav"
REF_TEXT = (
    "Hello, friend. Hello, friend? That's lame. Maybe I should give you a name, "
    "but that's a slippery slope. You're only in my head. We have to remember that."
)

def synthesize_voice(
    text: str,
    topic: str = "",
    summary: str = "",
    chat_id: str = "",
    tg_token: str = "",
    total_tokens: float = 0,
    prompt_tokens: float = 0,
    candidate_tokens: float = 0
) -> str:
    """Inference pipeline for Elliot Alderson voice persona with optional direct Telegram dispatch."""
    start_time = time.time()
    if not text.strip():
        text = "Hello, friend. Stay focused on what you can control."

    with tempfile.NamedTemporaryFile(delete=False, suffix=".wav") as tmp:
        out_path = tmp.name

    print(f"🎙️ Starting neural synthesis ({len(text.split())} words)...")
    try:
        wav, sr, _ = f5tts.infer(
            ref_file=REF_AUDIO,
            ref_text=REF_TEXT,
            gen_text=text,
            nfe_step=24,
        )
    except TypeError:
        wav, sr, _ = f5tts.infer(
            ref_file=REF_AUDIO,
            ref_text=REF_TEXT,
            gen_text=text,
        )

    # Boost whisper vocal amplitude for clear mobile phone speaker listening
    try:
        import numpy as np
        if hasattr(wav, "cpu"):
            wav_np = wav.cpu().numpy()
        else:
            wav_np = np.array(wav)
        max_val = float(np.max(np.abs(wav_np)))
        if max_val > 1e-4:
            gain = min(3.5, 0.94 / max_val)
            wav = wav * gain
    except Exception as norm_err:
        print(f"Peak normalization skipped: {norm_err}")

    f5tts.export_wav(wav, out_path)
    elapsed_sec = time.time() - start_time
    print(f"🎉 Speech synthesis complete in {elapsed_sec:.1f}s!")

def _escape_md(text: str) -> str:
    for ch in ("_", "*", "`", "["):
        text = text.replace(ch, "\\" + ch)
    return text

def _build_multipart(fields: dict, audio_bytes: bytes, filename: str, boundary: str) -> bytes:
    parts = []
    for k, v in fields.items():
        parts.append(f"--{boundary}\r\nContent-Disposition: form-data; name=\"{k}\"\r\n\r\n{v}\r\n".encode("utf-8"))
    parts.append(
        f"--{boundary}\r\nContent-Disposition: form-data; name=\"audio\"; filename=\"{filename}\"\r\nContent-Type: audio/wav\r\n\r\n".encode("utf-8")
        + audio_bytes
        + f"\r\n--{boundary}--\r\n".encode("utf-8")
    )
    return b"".join(parts)

    # Direct Telegram dispatch if credentials are provided by caller
    if chat_id and tg_token:
        try:
            resolved_topic = topic.strip() if topic else "Seeing Through the Illusion"
            resolved_summary = summary.strip() if summary else "Execute your next deliberate move."
            tot_tok = int(total_tokens)
            prm_tok = int(prompt_tokens)
            cnd_tok = int(candidate_tokens)
            token_line = f"🪙 *Tokens:* {tot_tok} ({prm_tok} in / {cnd_tok} out)\n" if tot_tok > 0 else ""

            caption = (
                f"🎙️ *Personal OS — Motivation Engine*\n\n"
                f"📌 *Topic:* _{_escape_md(resolved_topic)}_\n"
                f"📝 *Summary:* _{_escape_md(resolved_summary)}_\n"
                f"⏱️ *Process Time:* {elapsed_sec:.1f}s\n"
                f"{token_line}\n"
                f"_\"Hello, friend...\"_"
            )

            with open(out_path, "rb") as f:
                audio_bytes = f.read()

            boundary = "----PersonalOSBoundary" + os.urandom(8).hex()
            filename = f"daily_{int(time.time())}.wav"
            fields = {
                "chat_id": chat_id,
                "title": f"Motivation: {resolved_topic}",
                "performer": "Elliot Alderson",
                "caption": caption,
                "parse_mode": "Markdown",
            }
            tg_url = f"https://api.telegram.org/bot{tg_token}/sendAudio"

            try:
                body = _build_multipart(fields, audio_bytes, filename, boundary)
                tg_req = urllib.request.Request(tg_url, data=body, headers={"Content-Type": f"multipart/form-data; boundary={boundary}"})
                with urllib.request.urlopen(tg_req, timeout=60) as tg_resp:
                    print(f"📲 Telegram delivery complete! Status: {tg_resp.status} (Total time: {elapsed_sec:.1f}s)")
            except urllib.error.HTTPError as http_err:
                # Fallback to plain text caption if Telegram rejects Markdown
                if http_err.code == 400:
                    print(f"⚠️ Telegram Markdown rejected (400), retrying plain text caption...")
                    fields.pop("parse_mode", None)
                    fields["caption"] = f"🎙️ Personal OS — Motivation Engine\n\nTopic: {resolved_topic}\nSummary: {resolved_summary}\nProcess Time: {elapsed_sec:.1f}s\n\n\"Hello, friend...\""
                    body = _build_multipart(fields, audio_bytes, filename, boundary)
                    tg_req = urllib.request.Request(tg_url, data=body, headers={"Content-Type": f"multipart/form-data; boundary={boundary}"})
                    with urllib.request.urlopen(tg_req, timeout=60) as tg_resp:
                        print(f"📲 Telegram plain fallback complete! Status: {tg_resp.status}")
                else:
                    raise

        except Exception as tg_err:
            print(f"❌ Telegram delivery failed: {tg_err}")

    return out_path

# 2. Native Gradio UI & Explicit synthesize_voice API Endpoint
with gr.Blocks(title="🎙️ Personal OS: Elliot Alderson TTS Engine") as demo:
    gr.Markdown("# 🎙️ Personal OS: Elliot Alderson TTS Engine\n24/7 Cloud Zero-Shot Neural Voice Cloning powered by F5-TTS.")
    with gr.Row():
        with gr.Column():
            inp_text = gr.Textbox(
                label="Input Monologue Text", 
                value="Hello, friend. It's happening again. Focus on what you can control."
            )
            btn = gr.Button("Synthesize Audio", variant="primary")
        with gr.Column():
            out_audio = gr.Audio(label="Synthesized Speech (Elliot Alderson)", type="filepath")

    inp_topic = gr.Textbox(visible=False)
    inp_summary = gr.Textbox(visible=False)
    inp_chat_id = gr.Textbox(visible=False)
    inp_token = gr.Textbox(visible=False)
    inp_tot_tok = gr.Number(value=0, visible=False)
    inp_prm_tok = gr.Number(value=0, visible=False)
    inp_cnd_tok = gr.Number(value=0, visible=False)

    btn.click(
        fn=synthesize_voice,
        inputs=[inp_text, inp_topic, inp_summary, inp_chat_id, inp_token, inp_tot_tok, inp_prm_tok, inp_cnd_tok],
        outputs=out_audio,
        api_name="synthesize_voice"
    )

if __name__ == "__main__":
    demo.queue().launch()
