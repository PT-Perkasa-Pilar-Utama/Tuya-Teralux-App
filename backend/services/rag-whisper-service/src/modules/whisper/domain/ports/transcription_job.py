from pydantic import BaseModel
from typing import Optional
from src.modules.whisper.domain.ports.job_status import JobStatus


class TranscriptionJob(BaseModel):
    """Domain model representing a transcription job."""
    id: str
    filename: str
    status: JobStatus
    result: Optional[str] = None
    error: Optional[str] = None

    def is_complete(self) -> bool:
        """Check if job is completed."""
        return self.status == JobStatus.COMPLETED

    def is_failed(self) -> bool:
        """Check if job has failed."""
        return self.status == JobStatus.FAILED

    def is_pending(self) -> bool:
        """Check if job is still pending or processing."""
        return self.status in [JobStatus.PENDING, JobStatus.PROCESSING]
