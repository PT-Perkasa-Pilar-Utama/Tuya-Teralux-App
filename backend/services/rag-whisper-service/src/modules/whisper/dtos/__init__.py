"""Whisper DTOs for gRPC communication."""

from dataclasses import dataclass, field
from typing import Optional, List, Dict, Any
from enum import Enum
import time


class JobStatus(str, Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    COMPLETED = "completed"
    FAILED = "failed"


class UploadSessionStatus(str, Enum):
    UPLOADING = "uploading"
    READY = "ready"
    CONSUMED = "consumed"
    EXPIRED = "expired"


@dataclass
class TranscribeRequestDTO:
    audio_data: bytes
    file_name: str
    language: str = "id"
    diarize: bool = False
    model_id: Optional[str] = None
    correlation_id: Optional[str] = None


@dataclass
class TranscribeResponseDTO:
    task_id: str
    status: str
    transcript: Optional[str] = None
    error: Optional[str] = None
    duration_ms: int = 0


@dataclass
class JobStatusDTO:
    job_id: str
    status: str
    result: Optional[str] = None
    error: Optional[str] = None
    file_name: str = ""
    created_at: int = 0
    updated_at: int = 0


@dataclass
class CreateUploadSessionRequestDTO:
    file_name: str
    total_size: int
    chunk_count: int
    correlation_id: Optional[str] = None


@dataclass
class UploadSessionDTO:
    session_id: str
    file_name: str
    total_size: int
    chunk_count: int
    uploaded_chunks: int = 0
    status: str = "uploading"
    created_at: int = field(default_factory=lambda: int(time.time()))
    expires_at: int = 0
    chunks: Dict[int, bytes] = field(default_factory=dict)
    
    def __post_init__(self):
        if self.expires_at == 0:
            # Default 24 hours TTL
            self.expires_at = self.created_at + (24 * 60 * 60)


@dataclass
class UploadChunkRequestDTO:
    session_id: str
    chunk_index: int
    chunk_data: bytes
    correlation_id: Optional[str] = None


@dataclass
class UploadChunkResponseDTO:
    session_id: str
    chunk_index: int
    success: bool
    error: Optional[str] = None
    uploaded_chunks: int = 0


@dataclass
class FinalizeUploadSessionResponseDTO:
    session_id: str
    merged_file_path: str
    file_name: str
    total_size: int
    success: bool
    error: Optional[str] = None
