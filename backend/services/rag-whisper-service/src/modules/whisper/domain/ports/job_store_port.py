from abc import ABC, abstractmethod
from typing import Optional
from src.modules.whisper.domain.ports.transcription_job import TranscriptionJob


class JobStorePort(ABC):
    """Port defining the interface for job state storage."""

    @abstractmethod
    async def save_job(self, job: TranscriptionJob) -> None:
        """
        Save job state.

        Args:
            job: The transcription job to save.
        """
        pass

    @abstractmethod
    async def get_job(self, job_id: str) -> Optional[TranscriptionJob]:
        """
        Get job by ID.

        Args:
            job_id: The job identifier.

        Returns:
            The transcription job if found, None otherwise.
        """
        pass
