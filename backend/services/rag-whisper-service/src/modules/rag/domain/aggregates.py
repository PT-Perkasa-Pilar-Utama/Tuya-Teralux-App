"""
Aggregate Roots for the RAG module.

Aggregates are clusters of domain objects that are treated as a single
unit and have a lifecycle. They enforce invariants and raise domain events.
"""
from dataclasses import dataclass, field
from datetime import datetime
from typing import List, Dict, Any, Optional
import uuid

from src.modules.rag.domain.ports.document import DocumentDomain
from src.modules.rag.domain.events import (
    DocumentAddedEvent,
    ChatMessageAddedEvent,
    QueryExecutedEvent,
)


@dataclass
class ChatMessage:
    """Value object representing a chat message."""
    content: str
    role: str  # 'user' or 'assistant'
    timestamp: datetime = field(default_factory=datetime.utcnow)
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert message to dictionary."""
        return {
            "content": self.content,
            "role": self.role,
            "timestamp": self.timestamp.isoformat(),
            "metadata": self.metadata,
        }


@dataclass
class QueryResult:
    """Value object representing a query result."""
    query: str
    answer: str
    retrieved_documents: List[DocumentDomain] = field(default_factory=list)
    latency_ms: float = 0.0
    timestamp: datetime = field(default_factory=datetime.utcnow)

    @property
    def document_count(self) -> int:
        """Return the number of retrieved documents."""
        return len(self.retrieved_documents)

    def to_dict(self) -> Dict[str, Any]:
        """Convert result to dictionary."""
        return {
            "query": self.query,
            "answer": self.answer,
            "document_count": self.document_count,
            "latency_ms": self.latency_ms,
            "timestamp": self.timestamp.isoformat(),
        }


class RAGSession:
    """
    Aggregate root representing a RAG session.

    A session encapsulates a user's interaction with the RAG system,
    including documents added, chat history, and query results.
    """

    def __init__(self, user_id: str, session_id: Optional[str] = None):
        """
        Initialize a new RAG session.

        Args:
            user_id: The ID of the user owning this session.
            session_id: Optional session ID. Generated if not provided.
        """
        self.session_id = session_id or str(uuid.uuid4())
        self.user_id = user_id
        self.documents: List[DocumentDomain] = []
        self.history: List[ChatMessage] = []
        self.created_at = datetime.utcnow()
        self.updated_at = self.created_at
        self._pending_events: List[Any] = []

    @property
    def document_count(self) -> int:
        """Return the number of documents in the session."""
        return len(self.documents)

    @property
    def message_count(self) -> int:
        """Return the number of messages in the session."""
        return len(self.history)

    def add_document(self, doc: DocumentDomain) -> None:
        """
        Add a document to the session.

        Args:
            doc: The document to add.
        """
        self.documents.append(doc)
        self.updated_at = datetime.utcnow()
        self._raise_event(DocumentAddedEvent(
            document_id=doc.id,
            content_preview=doc.content[:100],
            metadata=doc.metadata,
        ))

    def add_user_message(self, message: str) -> ChatMessage:
        """
        Add a user message to the chat history.

        Args:
            message: The user's message.

        Returns:
            The created ChatMessage instance.
        """
        chat_msg = ChatMessage(content=message, role="user")
        self.history.append(chat_msg)
        self.updated_at = datetime.utcnow()
        self._raise_event(ChatMessageAddedEvent(
            session_id=self.session_id,
            message=message,
            role="user",
        ))
        return chat_msg

    def add_assistant_message(self, message: str) -> ChatMessage:
        """
        Add an assistant message to the chat history.

        Args:
            message: The assistant's response.

        Returns:
            The created ChatMessage instance.
        """
        chat_msg = ChatMessage(content=message, role="assistant")
        self.history.append(chat_msg)
        self.updated_at = datetime.utcnow()
        self._raise_event(ChatMessageAddedEvent(
            session_id=self.session_id,
            message=message,
            role="assistant",
        ))
        return chat_msg

    def record_query(self, result: QueryResult) -> None:
        """
        Record a query execution in the session.

        Args:
            result: The query result to record.
        """
        # Add query as user message
        self.add_user_message(result.query)
        # Add answer as assistant message
        self.add_assistant_message(result.answer)

    def get_recent_messages(self, limit: int = 10) -> List[Dict[str, str]]:
        """
        Get recent chat messages formatted for LLM context.

        Args:
            limit: Maximum number of messages to return.

        Returns:
            List of message dictionaries with role and content.
        """
        recent = self.history[-limit:]
        return [{"role": msg.role, "content": msg.content} for msg in recent]

    def find_documents_by_metadata(
        self,
        key: str,
        value: Any
    ) -> List[DocumentDomain]:
        """
        Find documents in the session by metadata key-value pair.

        Args:
            key: Metadata key to search for.
            value: Value to match.

        Returns:
            List of matching documents.
        """
        return [
            doc for doc in self.documents
            if doc.has_metadata(key) and doc.get_metadata(key) == value
        ]

    def _raise_event(self, event: Any) -> None:
        """
        Record a domain event for later dispatch.

        Args:
            event: The domain event to record.
        """
        self._pending_events.append(event)

    def get_pending_events(self) -> List[Any]:
        """
        Get and clear pending domain events.

        Returns:
            List of pending events.
        """
        events = self._pending_events.copy()
        self._pending_events.clear()
        return events

    def to_dict(self) -> Dict[str, Any]:
        """Convert session to dictionary representation."""
        return {
            "session_id": self.session_id,
            "user_id": self.user_id,
            "document_count": self.document_count,
            "message_count": self.message_count,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
            "documents": [doc.to_dict() for doc in self.documents],
            "history": [msg.to_dict() for msg in self.history],
        }
