from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from typing import Optional, Dict, Any

from src.modules.shared.infrastructure.auth import get_current_user
from src.modules.shared.infrastructure.di.container import container
from src.modules.rag.application.service import RAGApplicationService
from src.modules.rag.infrastructure.factory import LLMAdapterFactory
from src.modules.shared.infrastructure.settings import settings

router = APIRouter(prefix="/v1/rag", tags=["RAG"])


def get_rag_service() -> RAGApplicationService:
    """Get RAG service from DI container."""
    return container.rag_service()

class IngestRequest(BaseModel):
    text: str
    metadata: Optional[Dict[str, Any]] = None

class QueryRequest(BaseModel):
    question: str
    model_id: Optional[str] = None

class TranslateRequest(BaseModel):
    text: str
    target_lang: str
    model_id: Optional[str] = None

class SummaryRequest(BaseModel):
    text: str
    model_id: Optional[str] = None

@router.post("/ingest", dependencies=[Depends(get_current_user)])
async def ingest(request: IngestRequest, service: RAGApplicationService = Depends(get_rag_service)):
    chunks = await service.ingest_text(request.text, request.metadata)
    return {"status": "success", "chunks_ingested": chunks}

@router.post("/query", dependencies=[Depends(get_current_user)])
async def query(request: QueryRequest, service: RAGApplicationService = Depends(get_rag_service)):
    llm_client = None
    if request.model_id:
        llm_client = LLMAdapterFactory.create_llm(request.model_id)
    answer = await service.query(request.question, llm_client)
    return {"answer": answer}

@router.post("/chat", dependencies=[Depends(get_current_user)])
async def chat(request: QueryRequest, service: RAGApplicationService = Depends(get_rag_service)):
    llm_client = None
    if request.model_id:
        llm_client = LLMAdapterFactory.create_llm(request.model_id)
    answer = await service.chat(request.question, llm_client)
    return {"answer": answer}

@router.post("/control", dependencies=[Depends(get_current_user)])
async def control(request: QueryRequest, service: RAGApplicationService = Depends(get_rag_service)):
    # Reusing QueryRequest for simplicity (question mapping to command)
    llm_client = None
    if request.model_id:
        llm_client = LLMAdapterFactory.create_llm(request.model_id)
    result = await service.control(request.question, llm_client)
    return {"control_action": result}

@router.get("/status/{task_id}", dependencies=[Depends(get_current_user)])
async def get_status(task_id: str):
    # Placeholder status check
    return {"task_id": task_id, "status": "COMPLETED", "result": "Task finished successfully"}

@router.post("/translate", dependencies=[Depends(get_current_user)])
async def translate(request: TranslateRequest, service: RAGApplicationService = Depends(get_rag_service)):
    llm_client = None
    if request.model_id:
        llm_client = LLMAdapterFactory.create_llm(request.model_id)
    result = await service.translate(request.text, request.target_lang, llm_client)
    return {"translated_text": result}

@router.post("/summary", dependencies=[Depends(get_current_user)])
async def summarize(request: SummaryRequest, service: RAGApplicationService = Depends(get_rag_service)):
    llm_client = None
    if request.model_id:
        llm_client = LLMAdapterFactory.create_llm(request.model_id)
    result = await service.summarize(request.text, llm_client)
    return {"summary": result}
