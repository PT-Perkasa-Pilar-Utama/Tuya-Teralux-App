"""
Domain Events for the RAG module.

These events represent significant occurrences in the domain
and are used for event-driven architecture and domain logic.
"""
from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, Any, Optional


@dataclass
class DocumentAddedEvent:
    """Event raised when a document is added to the system."""
    document_id: str
    content_preview: str
    metadata: Dict[str, Any] = field(default_factory=dict)
    timestamp: datetime = field(default_factory=datetime.utcnow)

    @property
    def content_length(self) -> int:
        """Return the length of the content preview."""
        return len(self.content_preview)


@dataclass
class DocumentIngestedEvent:
    """Event raised when documents are ingested into the vector store."""
    document_ids: list[str] = field(default_factory=list)
    chunks_count: int = 0
    metadata: Dict[str, Any] = field(default_factory=dict)
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class QueryExecutedEvent:
    """Event raised when a query is executed against the vector store."""
    query: str
    results_count: int
    latency_ms: float
    retrieved_document_ids: list[str] = field(default_factory=list)
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class ChatMessageAddedEvent:
    """Event raised when a chat message is added to a session."""
    session_id: str
    message: str
    role: str  # 'user' or 'assistant'
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class TranslationExecutedEvent:
    """Event raised when text is translated."""
    source_text: str
    target_language: str
    translated_text: str
    latency_ms: float
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class SummaryGeneratedEvent:
    """Event raised when a summary is generated."""
    source_length: int
    summary_length: int
    compression_ratio: float
    latency_ms: float
    timestamp: datetime = field(default_factory=datetime.utcnow)

    def __post_init__(self):
        """Calculate compression ratio if not provided."""
        if self.source_length > 0 and self.compression_ratio == 0:
            self.compression_ratio = self.summary_length / self.source_length


@dataclass
class DeviceControlAnalyzedEvent:
    """Event raised when a device control command is analyzed."""
    command: str
    inferred_action: Optional[str] = None
    target_device: Optional[str] = None
    confidence: float = 0.0
    timestamp: datetime = field(default_factory=datetime.utcnow)
