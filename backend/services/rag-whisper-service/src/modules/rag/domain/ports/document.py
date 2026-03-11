from pydantic import BaseModel, Field
from typing import Dict, Any
from datetime import datetime
import hashlib


class DocumentDomain(BaseModel):
    """
    Domain model representing a document in the RAG system.

    This is the canonical document representation used throughout
    the application for vector storage and retrieval.
    """
    page_content: str = Field(..., description="The main text content of the document")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Document metadata")
    id: str = Field(default_factory="", description="Unique document identifier")
    created_at: datetime = Field(default_factory=datetime.utcnow, description="Creation timestamp")

    class Config:
        arbitrary_types_allowed = True

    def __init__(self, **data):
        super().__init__(**data)
        # Generate ID if not provided
        if not self.id:
            self.id = self._generate_id()

    def _generate_id(self) -> str:
        """Generate a unique ID based on content and timestamp."""
        content_hash = hashlib.md5(
            f"{self.page_content}{self.created_at.isoformat()}".encode()
        ).hexdigest()
        return f"doc_{content_hash[:12]}"

    @property
    def word_count(self) -> int:
        """Return the number of words in the document."""
        return len(self.page_content.split())

    @property
    def char_count(self) -> int:
        """Return the number of characters in the document."""
        return len(self.page_content)

    def has_metadata(self, key: str) -> bool:
        """Check if document has specific metadata key."""
        return key in self.metadata

    def get_metadata(self, key: str, default: Any = None) -> Any:
        """Get metadata value by key with optional default."""
        return self.metadata.get(key, default)

    def set_metadata(self, key: str, value: Any) -> None:
        """Set a metadata value."""
        self.metadata[key] = value

    def truncate(self, max_length: int) -> None:
        """Truncate content to maximum length."""
        if len(self.page_content) > max_length:
            self.page_content = self.page_content[:max_length] + "..."

    def is_empty(self) -> bool:
        """Check if document content is empty."""
        return len(self.page_content.strip()) == 0

    def to_dict(self) -> Dict[str, Any]:
        """Convert document to dictionary."""
        return {
            "id": self.id,
            "content": self.page_content,
            "metadata": self.metadata,
            "created_at": self.created_at.isoformat(),
            "word_count": self.word_count,
            "char_count": self.char_count,
        }
