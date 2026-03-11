"""
Domain layer for the RAG module.

Contains entities, value objects, aggregates, events, and repository interfaces.
"""
from src.modules.rag.domain.ports.document import DocumentDomain
from src.modules.rag.domain.ports.text_splitter_port import TextSplitterPort
from src.modules.rag.domain.aggregates import RAGSession, ChatMessage, QueryResult
from src.modules.rag.domain.events import (
    DocumentAddedEvent,
    DocumentIngestedEvent,
    QueryExecutedEvent,
    ChatMessageAddedEvent,
    TranslationExecutedEvent,
    SummaryGeneratedEvent,
    DeviceControlAnalyzedEvent,
)
from src.modules.rag.domain.repositories import DocumentRepository, SessionRepository

__all__ = [
    # Entities
    "DocumentDomain",
    # Ports
    "TextSplitterPort",
    # Aggregates
    "RAGSession",
    "ChatMessage",
    "QueryResult",
    # Events
    "DocumentAddedEvent",
    "DocumentIngestedEvent",
    "QueryExecutedEvent",
    "ChatMessageAddedEvent",
    "TranslationExecutedEvent",
    "SummaryGeneratedEvent",
    "DeviceControlAnalyzedEvent",
    # Repositories
    "DocumentRepository",
    "SessionRepository",
]
