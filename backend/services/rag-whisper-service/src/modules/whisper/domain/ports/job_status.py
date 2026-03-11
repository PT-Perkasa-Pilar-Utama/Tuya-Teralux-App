from enum import Enum


class JobStatus(str, Enum):
    """Enum representing the status of a transcription job."""
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"
