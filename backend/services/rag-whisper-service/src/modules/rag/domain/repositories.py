"""
Repository Pattern interfaces for the RAG module.

Repositories provide an abstraction over data access, allowing the domain
layer to remain agnostic of persistence mechanisms.
"""
from abc import ABC, abstractmethod
from typing import List, Optional, Dict, Any
from datetime import datetime

from src.modules.rag.domain.ports.document import DocumentDomain
from src.modules.rag.domain.aggregates import RAGSession


class DocumentRepository(ABC):
    """
    Repository interface for Document persistence.

    Provides CRUD operations and domain-specific queries for documents.
    """

    @abstractmethod
    async def save(self, doc: DocumentDomain) -> str:
        """
        Save a document to the repository.

        Args:
            doc: The document to save.

        Returns:
            The ID of the saved document.
        """
        pass

    @abstractmethod
    async def find_by_id(self, doc_id: str) -> Optional[DocumentDomain]:
        """
        Find a document by its ID.

        Args:
            doc_id: The document ID to search for.

        Returns:
            The document if found, None otherwise.
        """
        pass

    @abstractmethod
    async def find_by_metadata(
        self,
        key: str,
        value: Any
    ) -> List[DocumentDomain]:
        """
        Find documents by metadata key-value pair.

        Args:
            key: Metadata key to search for.
            value: Value to match.

        Returns:
            List of matching documents.
        """
        pass

    @abstractmethod
    async def find_similar(
        self,
        query: str,
        k: int = 4
    ) -> List[DocumentDomain]:
        """
        Find documents similar to the query (semantic search).

        Args:
            query: The query text for similarity search.
            k: Number of results to return.

        Returns:
            List of similar documents ordered by relevance.
        """
        pass

    @abstractmethod
    async def delete(self, doc_id: str) -> bool:
        """
        Delete a document from the repository.

        Args:
            doc_id: The ID of the document to delete.

        Returns:
            True if deleted, False if document didn't exist.
        """
        pass

    @abstractmethod
    async def exists(self, doc_id: str) -> bool:
        """
        Check if a document exists in the repository.

        Args:
            doc_id: The document ID to check.

        Returns:
            True if document exists, False otherwise.
        """
        pass


class SessionRepository(ABC):
    """
    Repository interface for RAGSession persistence.

    Provides CRUD operations for RAG session aggregates.
    """

    @abstractmethod
    async def save(self, session: RAGSession) -> str:
        """
        Save a session to the repository.

        Args:
            session: The session to save.

        Returns:
            The ID of the saved session.
        """
        pass

    @abstractmethod
    async def find_by_id(self, session_id: str) -> Optional[RAGSession]:
        """
        Find a session by its ID.

        Args:
            session_id: The session ID to search for.

        Returns:
            The session if found, None otherwise.
        """
        pass

    @abstractmethod
    async def find_by_user(self, user_id: str) -> List[RAGSession]:
        """
        Find all sessions for a specific user.

        Args:
            user_id: The user ID to search for.

        Returns:
            List of sessions owned by the user.
        """
        pass

    @abstractmethod
    async def find_active_sessions(
        self,
        user_id: str,
        since: Optional[datetime] = None
    ) -> List[RAGSession]:
        """
        Find active sessions for a user.

        Args:
            user_id: The user ID to search for.
            since: Optional cutoff datetime for "active" sessions.

        Returns:
            List of active sessions.
        """
        pass

    @abstractmethod
    async def delete(self, session_id: str) -> bool:
        """
        Delete a session from the repository.

        Args:
            session_id: The ID of the session to delete.

        Returns:
            True if deleted, False if session didn't exist.
        """
        pass

    @abstractmethod
    async def update_last_accessed(
        self,
        session_id: str,
        timestamp: Optional[datetime] = None
    ) -> bool:
        """
        Update the last accessed timestamp for a session.

        Args:
            session_id: The session ID to update.
            timestamp: Optional timestamp (defaults to now).

        Returns:
            True if updated, False if session didn't exist.
        """
        pass
