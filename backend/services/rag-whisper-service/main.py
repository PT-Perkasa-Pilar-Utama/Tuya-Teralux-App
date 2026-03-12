from fastapi import FastAPI
from concurrent import futures
import threading
import logging

from src.modules.rag.interfaces.rest.router import router as rag_router
from src.modules.whisper.interfaces.rest.router import router as whisper_router
from src.modules.whisper.interfaces.grpc import whisper_pb2_grpc
from src.modules.whisper.services.grpc_server import GrpcServer, WhisperServicer
from src.modules.whisper.controllers import WhisperController
from src.modules.whisper.usecases import WhisperTranscribeUseCase
from src.modules.whisper.infrastructure.adapters.whisper import (
    OpenAIWhisperAdapter,
    GroqWhisperAdapter,
    GeminiWhisperAdapter,
    OrionWhisperAdapter,
    LocalWhisperAdapter,
)
from src.modules.shared.infrastructure.settings import settings

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="RAG-Whisper Service (Modular)", version="0.2.0")

# Register REST routers (for backward compatibility)
app.include_router(rag_router)
app.include_router(whisper_router)

@app.get("/health")
async def health():
    return {"status": "ok"}


def create_whisper_client():
    """Create Whisper client based on settings."""
    provider = settings.WHISPER_PROVIDER.lower() if hasattr(settings, 'WHISPER_PROVIDER') else "openai"
    
    if provider == "openai":
        return OpenAIWhisperAdapter(settings.OPENAI_WHISPER_MODEL if hasattr(settings, 'OPENAI_WHISPER_MODEL') else "whisper-1")
    elif provider == "groq":
        return GroqWhisperAdapter(settings.GROQ_WHISPER_MODEL if hasattr(settings, 'GROQ_WHISPER_MODEL') else "whisper-large-v3")
    elif provider == "gemini":
        return GeminiWhisperAdapter(settings.GEMINI_MODEL if hasattr(settings, 'GEMINI_MODEL') else "gemini-1.5-flash")
    elif provider == "orion":
        return OrionWhisperAdapter(settings.ORION_MODEL if hasattr(settings, 'ORION_MODEL') else "orion-whisper")
    elif provider == "local":
        return LocalWhisperAdapter(settings.WHISPER_MODEL_PATH if hasattr(settings, 'WHISPER_MODEL_PATH') else "")
    else:
        logger.warning(f"Unknown Whisper provider: {provider}, using OpenAI")
        return OpenAIWhisperAdapter("whisper-1")


def start_grpc_server(port: int = 50051):
    """Start gRPC server in a separate thread."""
    # Create dependencies
    whisper_client = create_whisper_client()
    use_case = WhisperTranscribeUseCase(whisper_client)
    controller = WhisperController(use_case)
    
    # Start gRPC server
    grpc_server = GrpcServer(controller, port)
    grpc_server.start()
    logger.info(f"gRPC server running on port {port}")
    
    return grpc_server


if __name__ == "__main__":
    import uvicorn
    
    # Start gRPC server in background thread
    grpc_port = 50051
    grpc_thread = threading.Thread(
        target=lambda: start_grpc_server(grpc_port).wait_for_termination(),
        daemon=True
    )
    grpc_thread.start()
    
    # Start FastAPI server
    logger.info("Starting RAG-Whisper Service with gRPC + REST")
    uvicorn.run(app, host="0.0.0.0", port=8000)
