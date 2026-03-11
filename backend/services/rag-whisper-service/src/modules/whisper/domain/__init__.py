"""
Domain layer for the Whisper module.

Contains ports (interfaces), entities, and value objects.
"""
from src.modules.whisper.domain.ports.whisper_client_port import WhisperClientPort
from src.modules.whisper.domain.ports.job_store_port import JobStorePort
from src.modules.whisper.domain.ports.transcription_job import TranscriptionJob
from src.modules.whisper.domain.ports.job_status import JobStatus

__all__ = [
    "WhisperClientPort",
    "JobStorePort",
    "TranscriptionJob",
    "JobStatus",
]
