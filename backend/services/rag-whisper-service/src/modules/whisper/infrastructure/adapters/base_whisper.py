"""
Base Whisper Adapter providing common functionality.

Follows DRY principle by centralizing common Whisper adapter logic.
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
        model_id: Optional[str] = None
    ) -> str:
        """
        Transcribe audio file using OpenAI-compatible API.

        Args:
            file_path: Path to the audio file.
            model_id: Optional model identifier override.

        Returns:
            Transcribed text.
        """
        model = model_id or self.model_name

        with open(file_path, "rb") as audio_file:
            transcript = await self.client.audio.transcriptions.create(
                model=model,
                file=audio_file
            )

        return transcript.text
