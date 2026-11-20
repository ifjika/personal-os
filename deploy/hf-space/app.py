try:
    import spaces
except ImportError:
    pass

import os
import time
import tempfile
import urllib.request
import urllib.parse
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
                f"📌 *Topic:* _{resolved_topic}_\n"
                f"📝 *Summary:* _{resolved_summary}_\n"
                f"⏱️ *Process Time:* {elapsed_sec:.1f}s\n"
                f"{token_line}\n"
                f"_\"Hello, friend...\"_"
            )

            boundary = "----PersonalOSBoundary" + os.urandom(8).hex()
            with open(out_path, "rb") as f:
                audio_bytes = f.read()

            body = (
                f"--{boundary}\r\nContent-Disposition: form-data; name=\"chat_id\"\r\n\r\n{chat_id}\r\n"
                f"--{boundary}\r\nContent-Disposition: form-data; name=\"title\"\r\n\r\nMotivation: {resolved_topic}\r\n"
                f"--{boundary}\r\nContent-Disposition: form-data; name=\"performer\"\r\n\r\nElliot Alderson\r\n"
                f"--{boundary}\r\nContent-Disposition: form-data; name=\"caption\"\r\n\r\n{caption}\r\n"
                f"--{boundary}\r\nContent-Disposition: form-data; name=\"parse_mode\"\r\n\r\nMarkdown\r\n"
                f"--{boundary}\r\nContent-Disposition: form-data; name=\"audio\"; filename=\"daily_{int(time.time())}.wav\"\r\nContent-Type: audio/wav\r\n\r\n"
            ).encode("utf-8") + audio_bytes + f"\r\n--{boundary}--\r\n".encode("utf-8")

            tg_url = f"https://api.telegram.org/bot{tg_token}/sendAudio"
            tg_req = urllib.request.Request(tg_url, data=body, headers={"Content-Type": f"multipart/form-data; boundary={boundary}"})
            with urllib.request.urlopen(tg_req, timeout=60) as tg_resp:
                print(f"📲 Telegram delivery complete! Status: {tg_resp.status} (Total time: {elapsed_sec:.1f}s)")

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
