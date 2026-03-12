from typing import Optional
from src.modules.whisper.domain.ports import WhisperClientPort
from src.modules.shared.infrastructure.settings import settings
from src.modules.whisper.infrastructure.adapters.base_whisper import BaseWhisperAdapter


class OpenAIWhisperAdapter(BaseWhisperAdapter):
    """OpenAI Whisper adapter using the base class."""

    def __init__(self, model_name: str = "whisper-1"):
        super().__init__(
            model_name=model_name,
            api_key=settings.OPENAI_API_KEY
        )


class GroqWhisperAdapter(BaseWhisperAdapter):
    """Groq Whisper adapter using OpenAI-compatible API."""

    def __init__(self, model_name: str = "whisper-large-v3"):
        super().__init__(
            model_name=model_name,
            api_key=settings.GROQ_API_KEY,
            base_url="https://api.groq.com/openai/v1"
        )

class GeminiWhisperAdapter(WhisperClientPort):
    """Gemini 1.5 handles audio natively."""
    
    def __init__(self, model_name: str = "gemini-1.5-flash"):
        self.model_name = model_name
    
    async def transcribe(self, file_path: str, language: str = "id", diarize: bool = False) -> str:
        # TODO: Implement Gemini Audio transcription
        return f"Gemini ({self.model_name}) transcribed audio at {file_path}"

class OrionWhisperAdapter(WhisperClientPort):
    """Orion Whisper API adapter."""
    
    def __init__(self, model_name: str = "orion-whisper"):
        self.model_name = model_name
    
    async def transcribe(self, file_path: str, language: str = "id", diarize: bool = False) -> str:
        # TODO: Implement Orion Whisper API call
        return f"Orion ({self.model_name}) transcribed audio at {file_path}"

class LocalWhisperAdapter(WhisperClientPort):
    """Local whisper.cpp adapter."""
    
    def __init__(self, model_name: str = ""):
        self.model_name = model_name
    
    async def transcribe(self, file_path: str, language: str = "id", diarize: bool = False) -> str:
        # TODO: Implement whisper.cpp local transcription
        return f"Local ({self.model_name}) transcribed audio at {file_path}"
