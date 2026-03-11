from abc import ABC, abstractmethod
from typing import List
from src.modules.rag.domain.ports.document import DocumentDomain


class VectorStorePort(ABC):
    """Port defining the interface for vector stores."""

    @abstractmethod
    async def add_documents(self, documents: List[DocumentDomain]) -> None:
        """
        Add documents to the vector store.

        Args:
            documents: List of documents to store.
        """
        pass

    @abstractmethod
    async def similarity_search(
        self,
        query: str,
        k: int = 4
    ) -> List[DocumentDomain]:
        """
        Perform similarity search in the vector store.

        Args:
            query: The query text to search for.
            k: Number of results to return.

        Returns:
            List of similar documents.
        """
        pass
