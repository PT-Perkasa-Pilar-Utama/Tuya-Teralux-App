"""Whisper gRPC Server Implementation."""

import asyncio
import logging
from concurrent import futures
from typing import Optional

import grpc

from ..dtos import (
    TranscribeRequestDTO,
    CreateUploadSessionRequestDTO,
    UploadChunkRequestDTO,
)
from ..controllers import WhisperController
from ..interfaces.grpc import whisper_pb2, whisper_pb2_grpc

logger = logging.getLogger(__name__)


class WhisperServicer(whisper_pb2_grpc.WhisperServiceServicer):
    """gRPC servicer for Whisper service."""
    
    def __init__(self, controller: WhisperController):
        self.controller = controller
    
    async def Transcribe(self, request, context):
        """Handle gRPC Transcribe request."""
        try:
            dto = TranscribeRequestDTO(
                audio_data=request.audio_data,
                file_name=request.file_name,
                language=request.language or "id",
                diarize=request.diarize,
                model_id=request.model_id if request.model_id else None,
                correlation_id=request.correlation_id if request.correlation_id else None,
            )
            
            response = await self.controller.transcribe(dto)
            
            return whisper_pb2.TranscribeResponse(
                task_id=response.task_id,
                status=response.status,
                transcript=response.transcript or "",
                error=response.error or "",
                duration_ms=response.duration_ms,
            )
        except Exception as e:
            logger.error(f"Transcribe error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return whisper_pb2.TranscribeResponse(
                task_id="",
                status="failed",
                transcript="",
                error=str(e),
                duration_ms=0,
            )
    
    def GetJobStatus(self, request, context):
        """Handle gRPC GetJobStatus request."""
        try:
            status = self.controller.get_job_status(request.job_id)
            
            if not status:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details("Job not found")
                return whisper_pb2.JobStatusResponse()
            
            return whisper_pb2.JobStatusResponse(
                job_id=status.job_id,
                status=status.status,
                result=status.result or "",
                error=status.error or "",
                file_name=status.file_name,
                created_at=status.created_at,
                updated_at=status.updated_at,
            )
        except Exception as e:
            logger.error(f"GetJobStatus error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return whisper_pb2.JobStatusResponse()
    
    def CreateUploadSession(self, request, context):
        """Handle gRPC CreateUploadSession request."""
        try:
            dto = CreateUploadSessionRequestDTO(
                file_name=request.file_name,
                total_size=request.total_size,
                chunk_count=request.chunk_count,
                correlation_id=request.correlation_id if request.correlation_id else None,
            )
            
            session = self.controller.create_upload_session(dto)
            
            return whisper_pb2.UploadSessionResponse(
                session_id=session.session_id,
                file_name=session.file_name,
                total_size=session.total_size,
                chunk_count=session.chunk_count,
                uploaded_chunks=session.uploaded_chunks,
                status=session.status,
                created_at=session.created_at,
                expires_at=session.expires_at,
            )
        except Exception as e:
            logger.error(f"CreateUploadSession error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return whisper_pb2.UploadSessionResponse()
    
    def UploadChunk(self, request_iterator, context):
        """Handle gRPC UploadChunk streaming request."""
        try:
            last_response = None
            for request in request_iterator:
                dto = UploadChunkRequestDTO(
                    session_id=request.session_id,
                    chunk_index=request.chunk_index,
                    chunk_data=request.chunk_data,
                    correlation_id=request.correlation_id if request.correlation_id else None,
                )
                
                response = self.controller.upload_chunk(dto)
                last_response = response
            
            return whisper_pb2.UploadChunkResponse(
                session_id=last_response.session_id,
                chunk_index=last_response.chunk_index,
                success=last_response.success,
                error=last_response.error or "",
                uploaded_chunks=last_response.uploaded_chunks,
            )
        except Exception as e:
            logger.error(f"UploadChunk error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return whisper_pb2.UploadChunkResponse(
                session_id="",
                chunk_index=0,
                success=False,
                error=str(e),
                uploaded_chunks=0,
            )
    
    def GetUploadSessionStatus(self, request, context):
        """Handle gRPC GetUploadSessionStatus request."""
        try:
            session = self.controller.get_session_status(request.session_id)
            
            if not session:
                context.set_code(grpc.StatusCode.NOT_FOUND)
                context.set_details("Session not found")
                return whisper_pb2.UploadSessionResponse()
            
            return whisper_pb2.UploadSessionResponse(
                session_id=session.session_id,
                file_name=session.file_name,
                total_size=session.total_size,
                chunk_count=session.chunk_count,
                uploaded_chunks=session.uploaded_chunks,
                status=session.status,
                created_at=session.created_at,
                expires_at=session.expires_at,
            )
        except Exception as e:
            logger.error(f"GetUploadSessionStatus error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return whisper_pb2.UploadSessionResponse()
    
    def FinalizeUploadSession(self, request, context):
        """Handle gRPC FinalizeUploadSession request."""
        try:
            result = self.controller.finalize_session(request.session_id)
            
            if not result.success:
                context.set_code(grpc.StatusCode.FAILED_PRECONDITION)
                context.set_details(result.error or "Failed to finalize session")
            
            return whisper_pb2.FinalizeUploadSessionResponse(
                session_id=result.session_id,
                merged_file_path=result.merged_file_path,
                file_name=result.file_name,
                total_size=result.total_size,
                success=result.success,
                error=result.error or "",
            )
        except Exception as e:
            logger.error(f"FinalizeUploadSession error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return whisper_pb2.FinalizeUploadSessionResponse(
                session_id="",
                merged_file_path="",
                file_name="",
                total_size=0,
                success=False,
                error=str(e),
            )


class GrpcServer:
    """gRPC Server for Whisper service."""
    
    def __init__(self, controller: WhisperController, port: int = 50051):
        self.controller = controller
        self.port = port
        self.server = None
    
    def start(self):
        """Start gRPC server."""
        self.server = grpc.server(
            futures.ThreadPoolExecutor(max_workers=10),
            options=[
                ('grpc.max_send_message_length', 50 * 1024 * 1024),  # 50MB
                ('grpc.max_receive_message_length', 50 * 1024 * 1024),
            ]
        )
        
        servicer = WhisperServicer(self.controller)
        whisper_pb2_grpc.add_WhisperServiceServicer_to_server(servicer, self.server)
        
        self.server.add_insecure_port(f'[::]:{self.port}')
        self.server.start()
        logger.info(f"gRPC server started on port {self.port}")
    
    def wait_for_termination(self):
        """Wait for server termination."""
        if self.server:
            self.server.wait_for_termination()
    
    def stop(self, grace=0):
        """Stop gRPC server."""
        if self.server:
            self.server.stop(grace)
            logger.info("gRPC server stopped")
