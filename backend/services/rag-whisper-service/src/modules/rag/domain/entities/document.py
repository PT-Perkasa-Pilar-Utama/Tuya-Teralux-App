from dataclasses import dataclass, field
from typing import Dict, Any
from datetime import datetime
import hashlib


@dataclass
class Document:
    """
    Rich domain entity representing a document in the RAG system.
    
    This is a proper domain entity with behavior, not just a data container.
    """
    content: str
    metadata: Dict[str, Any] = field(default_factory=dict)
    id: str = field(default_factory=lambda: "")
    created_at: datetime = field(default_factory=datetime.utcnow)
    
    def __post_init__(self):
        """Generate ID if not provided."""
        if not self.id:
            self.id = self._generate_id()
    
    def _generate_id(self) -> str:
        """Generate a unique ID based on content and timestamp."""
        content_hash = hashlib.md5(
            f"{self.content}{self.created_at.isoformat()}".encode()
        ).hexdigest()
        return f"doc_{content_hash[:12]}"
    
    @property
    def word_count(self) -> int:
        """Return the number of words in the document."""
        return len(self.content.split())
    
    @property
    def char_count(self) -> int:
        """Return the number of characters in the document."""
        return len(self.content)
    
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
        if len(self.content) > max_length:
            self.content = self.content[:max_length] + "..."
    
    def is_empty(self) -> bool:
        """Check if document content is empty."""
        return len(self.content.strip()) == 0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert document to dictionary."""
        return {
            "id": self.id,
            "content": self.content,
            "metadata": self.metadata,
            "created_at": self.created_at.isoformat(),
            "word_count": self.word_count,
            "char_count": self.char_count,
        }
