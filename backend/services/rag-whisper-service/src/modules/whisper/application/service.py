import os
import uuid
from typing import Optional

from src.modules.whisper.domain.ports import (
    WhisperClientPort,
    JobStorePort,
    TranscriptionJob,
    JobStatus,
)

class WhisperApplicationService:
    def __init__(self, default_client: WhisperClientPort, job_store: JobStorePort):
        self.default_client = default_client
        self.job_store = job_store

    async def transcribe_sync(self, file_path: str, model_id: Optional[str] = None) -> str:
        return await self.default_client.transcribe(file_path, model_id)

    async def run_async_job(self, job_id: str, file_path: str, model_id: Optional[str] = None):
        job = TranscriptionJob(id=job_id, filename=file_path, status=JobStatus.PROCESSING)
        await self.job_store.save_job(job)
        
        try:
            text = await self.default_client.transcribe(file_path, model_id)
            job.status = JobStatus.COMPLETED
            job.result = text
        except Exception as e:
            job.status = JobStatus.FAILED
            job.error = str(e)
        finally:
            if os.path.exists(file_path):
                os.remove(file_path)
        
        await self.job_store.save_job(job)

    async def get_job_status(self, job_id: str) -> Optional[TranscriptionJob]:
        return await self.job_store.get_job(job_id)
