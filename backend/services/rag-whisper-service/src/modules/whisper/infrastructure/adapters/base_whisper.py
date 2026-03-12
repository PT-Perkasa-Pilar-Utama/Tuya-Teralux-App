"""
Base Whisper Adapter providing common functionality.

Follows DDRY principle by centralizing common Whisper adapter logic.
"""
from typing import Optional
from abc import ABC

from openai import AsyncOpenAI
from src.modules.whisper.domain.ports import WhisperClientPort


class BaseWhisperAdapter(WhisperClientPort, ABC):
    """
    Base class for Whisper transcription adapters.

    Provides common initialization and transcription logic
    for OpenAI-compatible Whisper APIs.
    """

    def __init__(
        self,
        model_name: str,
        api_key: str,
        base_url: Optional[str] = None
    ):
        """
        Initialize the base Whisper adapter.

        Args:
            model_name: The Whisper model name to use.
            api_key: API key for the service.
            base_url: Optional base URL for the API endpoint.
        """
        self.model_name = model_name
        self.client = AsyncOpenAI(
            api_key=api_key,
            base_url=base_url
        )

    async def transcribe(
        self,
        file_path: str,
        language: str = "id",
        diarize: bool = False
    ) -> str:
        """
        Transcribe audio file using OpenAI-compatible API.

        Args:
            file_path: Path to the audio file.
            language: Language code (e.g., "id", "en").
            diarize: Enable speaker diarization (not supported by all providers).

        Returns:
            Transcribed text.
        """
        with open(file_path, "rb") as audio_file:
            transcript = await self.client.audio.transcriptions.create(
                model=self.model_name,
                file=audio_file,
                language=language if language else "id"
            )

        return transcript.text
