from fastapi import APIRouter, Depends, UploadFile, File, HTTPException, BackgroundTasks
import shutil
import os
import uuid
from typing import Optional

from src.modules.shared.infrastructure.auth import get_current_user
from src.modules.shared.infrastructure.di.container import container
from src.modules.whisper.application.service import WhisperApplicationService
from src.modules.shared.infrastructure.settings import settings

router = APIRouter(prefix="/v1/whisper", tags=["Whisper"])


def get_whisper_service() -> WhisperApplicationService:
    """Get Whisper service from DI container."""
    return container.whisper_service()

@router.post("/transcribe", dependencies=[Depends(get_current_user)])
async def transcribe_sync(
    file: UploadFile = File(...),
    model_id: Optional[str] = None,
    service: WhisperApplicationService = Depends(get_whisper_service)
):
    temp_dir = "tmp/uploads"
    os.makedirs(temp_dir, exist_ok=True)
    file_path = f"{temp_dir}/{uuid.uuid4()}_{file.filename}"
    
    try:
        with open(file_path, "wb") as buffer:
            shutil.copyfileobj(file.file, buffer)
        
        text = await service.transcribe_sync(file_path, model_id)
        return {"text": text, "model_id": model_id or settings.DEFAULT_WHISPER_MODEL}
    finally:
        if os.path.exists(file_path):
            os.remove(file_path)

@router.post("/transcribe/async", dependencies=[Depends(get_current_user)])
async def transcribe_async(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    model_id: Optional[str] = None,
    service: WhisperApplicationService = Depends(get_whisper_service)
):
    temp_dir = "tmp/uploads"
    os.makedirs(temp_dir, exist_ok=True)
    file_path = f"{temp_dir}/{uuid.uuid4()}_{file.filename}"
    
    with open(file_path, "wb") as buffer:
        shutil.copyfileobj(file.file, buffer)
    
    job_id = str(uuid.uuid4())
    background_tasks.add_task(service.run_async_job, job_id, file_path, model_id)
    
    return {"job_id": job_id, "status": "pending"}

@router.get("/jobs/{job_id}", dependencies=[Depends(get_current_user)])
async def get_job_status(job_id: str, service: WhisperApplicationService = Depends(get_whisper_service)):
    job = await service.get_job_status(job_id)
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    return job
