from abc import ABC, abstractmethod
from typing import Optional


class WhisperClientPort(ABC):
    """Port defining the interface for Whisper transcription clients."""

    @abstractmethod
    async def transcribe(
        self,
        file_path: str,
        model_id: Optional[str] = None
    ) -> str:
        """
        Transcribe audio file.

        Args:
            file_path: Path to the audio file.
            model_id: Optional model identifier.

        Returns:
            Transcribed text.
        """
        pass
