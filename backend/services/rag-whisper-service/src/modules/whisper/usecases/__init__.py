"""Whisper Use Cases - Business Logic Layer."""

import asyncio
import os
import uuid
import time
from typing import Optional, Dict
from dataclasses import dataclass, field
import tempfile

from ..dtos import (
    TranscribeRequestDTO,
    TranscribeResponseDTO,
    JobStatusDTO,
    CreateUploadSessionRequestDTO,
    UploadSessionDTO,
    UploadChunkRequestDTO,
    UploadChunkResponseDTO,
    FinalizeUploadSessionResponseDTO,
    JobStatus,
    UploadSessionStatus,
)
from src.modules.shared.infrastructure.settings import settings


@dataclass
class TranscriptionTask:
    """Internal model for transcription task."""
    id: str
    file_name: str
    file_path: str
    status: str = JobStatus.PENDING
    result: Optional[str] = None
    error: Optional[str] = None
    language: str = "id"
    diarize: bool = False
    created_at: int = field(default_factory=lambda: int(time.time()))
    updated_at: int = field(default_factory=lambda: int(time.time()))


class WhisperTranscribeUseCase:
    """Use case for handling transcription requests."""
    
    def __init__(self, whisper_client, job_store: Optional[Dict] = None):
        self.whisper_client = whisper_client
        self.job_store = job_store if job_store is not None else {}
        self.upload_sessions: Dict[str, UploadSessionDTO] = {}
    
    async def transcribe(self, request: TranscribeRequestDTO) -> TranscribeResponseDTO:
        """
        Transcribe audio file.
        Returns task_id for async processing.
        """
        task_id = str(uuid.uuid4())
        
        # Save audio file temporarily
        temp_dir = settings.WHISPER_TEMP_DIR
        os.makedirs(temp_dir, exist_ok=True)
        file_path = f"{temp_dir}/{task_id}_{request.file_name}"
        
        try:
            with open(file_path, "wb") as f:
                f.write(request.audio_data)
            
            # Create task entry
            task = TranscriptionTask(
                id=task_id,
                file_name=request.file_name,
                file_path=file_path,
                status=JobStatus.PENDING,
                language=request.language,
                diarize=request.diarize,
            )
            self.job_store[task_id] = task
            
            # Start async transcription
            asyncio.create_task(self._process_transcription(task_id))
            
            return TranscribeResponseDTO(
                task_id=task_id,
                status=JobStatus.PENDING,
            )
        except Exception as e:
            return TranscribeResponseDTO(
                task_id=task_id,
                status=JobStatus.FAILED,
                error=str(e),
            )
    
    async def _process_transcription(self, task_id: str):
        """Process transcription asynchronously."""
        task: TranscriptionTask = self.job_store.get(task_id)
        if not task:
            return
        
        try:
            task.status = JobStatus.PROCESSING
            task.updated_at = int(time.time())
            
            # Call whisper client
            result = await self.whisper_client.transcribe(
                task.file_path,
                task.language,
                task.diarize
            )
            
            task.status = JobStatus.COMPLETED
            task.result = result
        except Exception as e:
            task.status = JobStatus.FAILED
            task.error = str(e)
        finally:
            task.updated_at = int(time.time())
            # Clean up file
            if os.path.exists(task.file_path):
                os.remove(task.file_path)
    
    def get_status(self, job_id: str) -> Optional[JobStatusDTO]:
        """Get transcription job status."""
        task: TranscriptionTask = self.job_store.get(job_id)
        if not task:
            return None
        
        return JobStatusDTO(
            job_id=task.id,
            status=task.status,
            result=task.result,
            error=task.error,
            file_name=task.file_name,
            created_at=task.created_at,
            updated_at=task.updated_at,
        )
    
    def create_upload_session(self, request: CreateUploadSessionRequestDTO) -> UploadSessionDTO:
        """Create upload session for chunked upload."""
        session_id = str(uuid.uuid4())
        session = UploadSessionDTO(
            session_id=session_id,
            file_name=request.file_name,
            total_size=request.total_size,
            chunk_count=request.chunk_count,
        )
        self.upload_sessions[session_id] = session
        return session
    
    def upload_chunk(self, request: UploadChunkRequestDTO) -> UploadChunkResponseDTO:
        """Upload a chunk of audio file."""
        session = self.upload_sessions.get(request.session_id)
        if not session:
            return UploadChunkResponseDTO(
                session_id=request.session_id,
                chunk_index=request.chunk_index,
                success=False,
                error="Session not found",
            )
        
        if session.status != UploadSessionStatus.UPLOADING:
            return UploadChunkResponseDTO(
                session_id=request.session_id,
                chunk_index=request.chunk_index,
                success=False,
                error="Session not in uploading state",
            )
        
        # Store chunk
        session.chunks[request.chunk_index] = request.chunk_data
        session.uploaded_chunks = len(session.chunks)
        
        # Check if all chunks received
        if session.uploaded_chunks >= session.chunk_count:
            session.status = UploadSessionStatus.READY
        
        return UploadChunkResponseDTO(
            session_id=request.session_id,
            chunk_index=request.chunk_index,
            success=True,
            uploaded_chunks=session.uploaded_chunks,
        )
    
    def get_session_status(self, session_id: str) -> Optional[UploadSessionDTO]:
        """Get upload session status."""
        return self.upload_sessions.get(session_id)
    
    def finalize_session(self, session_id: str) -> FinalizeUploadSessionResponseDTO:
        """Finalize upload session and merge chunks."""
        session = self.upload_sessions.get(session_id)
        if not session:
            return FinalizeUploadSessionResponseDTO(
                session_id=session_id,
                merged_file_path="",
                file_name="",
                total_size=0,
                success=False,
                error="Session not found",
            )
        
        if session.status != UploadSessionStatus.READY:
            return FinalizeUploadSessionResponseDTO(
                session_id=session_id,
                merged_file_path="",
                file_name=session.file_name,
                total_size=session.total_size,
                success=False,
                error="Session not ready for finalization",
            )
        
        try:
            # Merge chunks
            temp_dir = settings.WHISPER_TEMP_DIR
            os.makedirs(temp_dir, exist_ok=True)
            merged_path = f"{temp_dir}/merged_{session_id}_{session.file_name}"
            
            with open(merged_path, "wb") as f:
                # Write chunks in order
                for i in range(session.chunk_count):
                    chunk = session.chunks.get(i)
                    if chunk:
                        f.write(chunk)
            
            session.status = UploadSessionStatus.CONSUMED
            
            return FinalizeUploadSessionResponseDTO(
                session_id=session_id,
                merged_file_path=merged_path,
                file_name=session.file_name,
                total_size=session.total_size,
                success=True,
            )
        except Exception as e:
            return FinalizeUploadSessionResponseDTO(
                session_id=session_id,
                merged_file_path="",
                file_name=session.file_name,
                total_size=session.total_size,
                success=False,
                error=str(e),
            )
