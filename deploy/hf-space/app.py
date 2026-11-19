try:
    import spaces
except ImportError:
    pass

import os
import tempfile
from fastapi.responses import FileResponse
from pydantic import BaseModel
import gradio as gr
from f5_tts.api import F5TTS

# 1. Initialize F5-TTS model into memory
print("⏳ Loading F5-TTS voice model into memory...")
f5tts = F5TTS(device="cpu")
print("✅ F5-TTS voice model loaded successfully!")

REF_AUDIO = "elliot_ref_15s.wav"
REF_TEXT = (
    "Hello, friend. Hello, friend? That's lame. Maybe I should give you a name, "
    "but that's a slippery slope. You're only in my head. We have to remember that."
)

def synthesize_voice(text: str) -> str:
    """Inference pipeline for Elliot Alderson voice persona."""
    if not text.strip():
        text = "Hello, friend. Stay focused on what you can control."
    
    with tempfile.NamedTemporaryFile(delete=False, suffix=".wav") as tmp:
        out_path = tmp.name

    wav, sr, _ = f5tts.infer(
        ref_file=REF_AUDIO,
        ref_text=REF_TEXT,
        gen_text=text,
    )
    f5tts.export_wav(wav, out_path)
    return out_path

# 2. Interactive Web GUI (Gradio)
demo = gr.Interface(
    fn=synthesize_voice,
    inputs=gr.Textbox(
        label="Input Monologue Text", 
        value="Hello, friend. It's happening again. Focus on what you can control."
    ),
    outputs=gr.Audio(label="Synthesized Speech (Elliot Alderson)", type="filepath"),
    title="🎙️ Personal OS: Elliot Alderson TTS Engine",
    description="24/7 Cloud Zero-Shot Neural Voice Cloning powered by F5-TTS."
)

# 3. Attach custom REST API to the underlying FastAPI app of Gradio
class TTSRequest(BaseModel):
    text: str

@demo.app.post("/v1/tts")
async def api_tts(req: TTSRequest):
    output_wav = synthesize_voice(req.text)
    return FileResponse(output_wav, media_type="audio/wav")

# 4. Launch Gradio using native Space binding
if __name__ == "__main__":
    demo.launch()
