"""Whisper Controllers - gRPC Request Handlers."""

import time
from typing import Optional

from ..dtos import (
    TranscribeRequestDTO,
    TranscribeResponseDTO,
    JobStatusDTO,
    CreateUploadSessionRequestDTO,
    UploadSessionDTO,
    UploadChunkRequestDTO,
    UploadChunkResponseDTO,
    FinalizeUploadSessionResponseDTO,
)
from ..usecases import WhisperTranscribeUseCase


class WhisperController:
    """Controller for handling gRPC Whisper requests."""
    
    def __init__(self, use_case: WhisperTranscribeUseCase):
        self.use_case = use_case
    
    async def transcribe(self, request: TranscribeRequestDTO) -> TranscribeResponseDTO:
        """Handle transcription request."""
        return await self.use_case.transcribe(request)
    
    def get_job_status(self, job_id: str) -> Optional[JobStatusDTO]:
        """Handle job status request."""
        return self.use_case.get_status(job_id)
    
    def create_upload_session(self, request: CreateUploadSessionRequestDTO) -> UploadSessionDTO:
        """Handle upload session creation."""
        return self.use_case.create_upload_session(request)
    
    def upload_chunk(self, request: UploadChunkRequestDTO) -> UploadChunkResponseDTO:
        """Handle chunk upload."""
        return self.use_case.upload_chunk(request)
    
    def get_session_status(self, session_id: str) -> Optional[UploadSessionDTO]:
        """Handle session status request."""
        return self.use_case.get_session_status(session_id)
    
    def finalize_session(self, session_id: str) -> FinalizeUploadSessionResponseDTO:
        """Handle session finalization."""
        return self.use_case.finalize_session(session_id)
