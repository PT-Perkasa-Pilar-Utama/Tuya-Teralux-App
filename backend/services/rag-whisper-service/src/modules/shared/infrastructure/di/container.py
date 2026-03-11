"""
Dependency Injection Container using dependency-injector library.

This container provides a centralized way to manage all dependencies
following the Dependency Inversion Principle (DIP).
"""
from dependency_injector import containers, providers

from src.modules.rag.infrastructure.adapters.embeddings import (
    OpenAIEmbeddingAdapter,
    GeminiEmbeddingAdapter,
)
from src.modules.rag.infrastructure.adapters.llm import (
    OpenAIAdapter,
    GeminiAdapter,
    GroqAdapter,
)
from src.modules.rag.infrastructure.adapters.milvus import MilvusVectorStoreAdapter
from src.modules.rag.infrastructure.adapters.text_splitter import LangChainTextSplitterAdapter
from src.modules.rag.application.service import RAGApplicationService
from src.modules.rag.application.ingest_service import IngestService
from src.modules.rag.application.query_service import QueryService
from src.modules.rag.application.translate_service import TranslateService
from src.modules.rag.application.summarize_service import SummarizeService
from src.modules.rag.infrastructure.repositories.milvus_document_repository import MilvusDocumentRepository
from src.modules.rag.infrastructure.repositories.session_repository import InMemorySessionRepository

from src.modules.whisper.infrastructure.adapters.whisper import (
    OpenAIWhisperAdapter,
    GroqWhisperAdapter,
)
from src.modules.whisper.infrastructure.adapters.job_store import InMemoryJobStore
from src.modules.whisper.application.service import WhisperApplicationService

from src.modules.shared.infrastructure.settings import settings


class Container(containers.DeclarativeContainer):
    """
    Main DI container for the RAG-Whisper service.
    
    All dependencies are configured as singletons to ensure
    consistent state and optimal resource usage.
    """
    
    # Configuration
    config = providers.Configuration()
    
    # ============ RAG Module Dependencies ============
    
    # Embeddings
    openai_embeddings = providers.Singleton(
        OpenAIEmbeddingAdapter,
        model_name=config.openai.embedding_model_name,
    )
    
    gemini_embeddings = providers.Singleton(
        GeminiEmbeddingAdapter,
        model_name=config.gemini.embedding_model_name,
    )
    
    # LLMs
    openai_llm = providers.Singleton(
        OpenAIAdapter,
        model_name=config.openai.llm_model_name,
    )

    gemini_llm = providers.Singleton(
        GeminiAdapter,
        model_name=config.gemini.llm_model_name,
    )

    groq_llm = providers.Singleton(
        GroqAdapter,
        model_name=config.groq.llm_model_name,
    )

    # Vector Store
    # Note: MilvusVectorStoreAdapter needs the LangChain embedding object,
    # which is exposed via the .embeddings attribute of our adapter
    # Using a lambda to extract the LangChain object from the adapter
    vector_store = providers.Singleton(
        lambda emb_adapter: MilvusVectorStoreAdapter(emb_adapter.embeddings),
        emb_adapter=openai_embeddings,
    )

    # Text Splitter
    text_splitter = providers.Singleton(
        LangChainTextSplitterAdapter,
        chunk_size=1000,
        chunk_overlap=200,
    )

    # RAG Application Services (SRP-compliant)
    ingest_service = providers.Factory(
        IngestService,
        vector_store=vector_store,
        text_splitter=text_splitter,
    )

    query_service = providers.Factory(
        QueryService,
        vector_store=vector_store,
        default_llm=openai_llm,
    )

    translate_service = providers.Factory(
        TranslateService,
        default_llm=openai_llm,
    )

    summarize_service = providers.Factory(
        SummarizeService,
        default_llm=openai_llm,
    )

    # Legacy RAG Service (for backward compatibility - TO BE DEPRECATED)
    rag_service = providers.Factory(
        RAGApplicationService,
        vector_store=vector_store,
        default_llm=openai_llm,
    )

    # Repositories
    document_repository = providers.Factory(
        MilvusDocumentRepository,
        vector_store=vector_store,
    )

    session_repository = providers.Singleton(InMemorySessionRepository)
    
    # ============ Whisper Module Dependencies ============
    
    # Whisper Clients
    openai_whisper = providers.Singleton(
        OpenAIWhisperAdapter,
        model_name=config.openai.whisper_model_name,
    )
    
    groq_whisper = providers.Singleton(
        GroqWhisperAdapter,
        model_name=config.groq.whisper_model_name,
    )
    
    # Job Store
    job_store = providers.Singleton(InMemoryJobStore)
    
    # Whisper Application Service
    whisper_service = providers.Factory(
        WhisperApplicationService,
        default_client=openai_whisper,  # Default to OpenAI Whisper
        job_store=job_store,
    )


# Create and wire the container
def create_container() -> Container:
    """
    Create and configure the DI container with settings.
    
    Returns:
        Configured Container instance.
    """
    container = Container()
    
    # Wire configuration from settings
    container.config.openai.embedding_model_name.from_value(
        "text-embedding-3-small"
    )
    container.config.openai.llm_model_name.from_value("gpt-4o")
    container.config.openai.whisper_model_name.from_value("whisper-1")
    
    container.config.gemini.embedding_model_name.from_value(
        "models/embedding-001"
    )
    container.config.gemini.llm_model_name.from_value("gemini-pro")
    
    container.config.groq.llm_model_name.from_value("llama3-70b-8192")
    container.config.groq.whisper_model_name.from_value("whisper-large-v3")
    
    return container


# Global container instance
container = create_container()
