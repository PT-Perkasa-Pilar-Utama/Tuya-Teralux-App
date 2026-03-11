"""
Milvus Document Repository Implementation.

Implements DocumentRepository using Milvus vector store.
"""
from typing import List, Optional, Dict, Any

from src.modules.rag.domain.repositories import DocumentRepository
from src.modules.rag.domain.ports.document import DocumentDomain
from src.modules.rag.domain.ports.vector_store_port import VectorStorePort


class MilvusDocumentRepository(DocumentRepository):
    """
    Document repository implementation using Milvus vector store.

    Provides CRUD operations and semantic search for documents.
    """

    def __init__(self, vector_store: VectorStorePort):
        """
        Initialize the Milvus document repository.

        Args:
            vector_store: Vector store port for persistence.
        """
        self.vector_store = vector_store
        # In-memory index for ID-based lookups
        # In production, this should be backed by a database
        self._index: Dict[str, DocumentDomain] = {}

    async def save(self, doc: DocumentDomain) -> str:
        """
        Save a document to Milvus.

        Args:
            doc: The document to save.

        Returns:
            The ID of the saved document.
        """
        await self.vector_store.add_documents([doc])
        self._index[doc.id] = doc
        return doc.id

    async def find_by_id(self, doc_id: str) -> Optional[DocumentDomain]:
        """
        Find a document by its ID.

        Args:
            doc_id: The document ID to search for.

        Returns:
            The document if found, None otherwise.
        """
        return self._index.get(doc_id)

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
        return [
            doc for doc in self._index.values()
            if doc.has_metadata(key) and doc.get_metadata(key) == value
        ]

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
        results = await self.vector_store.similarity_search(query, k=k)
        # Also index results for future ID-based lookups
        for doc in results:
            if hasattr(doc, 'id') and doc.id:
                self._index[doc.id] = doc
        return results

    async def delete(self, doc_id: str) -> bool:
        """
        Delete a document from the repository.

        Note: Milvus doesn't support direct deletion by ID in this implementation.
        This only removes from the local index.

        Args:
            doc_id: The ID of the document to delete.

        Returns:
            True if deleted, False if document didn't exist.
        """
        if doc_id in self._index:
            del self._index[doc_id]
            return True
        return False

    async def exists(self, doc_id: str) -> bool:
        """
        Check if a document exists in the repository.

        Args:
            doc_id: The document ID to check.

        Returns:
            True if document exists, False otherwise.
        """
        return doc_id in self._index
