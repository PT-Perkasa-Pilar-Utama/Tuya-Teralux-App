from typing import Optional
from src.modules.whisper.domain.ports import WhisperClientPort
from src.modules.shared.infrastructure.settings import settings
from src.modules.whisper.infrastructure.adapters.base_whisper import BaseWhisperAdapter


class OpenAIWhisperAdapter(BaseWhisperAdapter):
    """OpenAI Whisper adapter using the base class."""

    def __init__(self, model_name: str):
        super().__init__(
            model_name=model_name,
            api_key=settings.OPENAI_API_KEY
        )


class GroqWhisperAdapter(BaseWhisperAdapter):
    """Groq Whisper adapter using OpenAI-compatible API."""

    def __init__(self, model_name: str):
        super().__init__(
            model_name=model_name,
            api_key=settings.GROQ_API_KEY,
            base_url="https://api.groq.com/openai/v1"
        )

class GeminiWhisperAdapter(WhisperClientPort):
    # Gemini 1.5 handles audio natively
    def __init__(self, model_name: str):
        self.model_name = model_name

    async def transcribe(self, file_path: str, model_id: Optional[str] = None) -> str:
        # Mocking Gemini Audio-to-Text for now
        return f"Gemini ({self.model_name}) transcribed audio at {file_path}"

class OrionWhisperAdapter(WhisperClientPort):
    def __init__(self, model_name: str):
        self.model_name = model_name

    async def transcribe(self, file_path: str, model_id: Optional[str] = None) -> str:
        # Placeholder for Orion Whisper API call
        return f"Orion ({self.model_name}) transcribed audio at {file_path}"

class LocalWhisperAdapter(WhisperClientPort):
    def __init__(self, model_name: str):
        self.model_name = model_name

    async def transcribe(self, file_path: str, model_id: Optional[str] = None) -> str:
        # Placeholder for whisper.cpp / local model
        return f"Local ({self.model_name}) transcribed audio at {file_path}"
