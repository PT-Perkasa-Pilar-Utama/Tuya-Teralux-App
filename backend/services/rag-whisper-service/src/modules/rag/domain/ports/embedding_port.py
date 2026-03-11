from abc import ABC, abstractmethod
from typing import List


class EmbeddingPort(ABC):
    """Port defining the interface for embedding models.

    Note: Methods are synchronous as LangChain embedding implementations
    are blocking operations.
    """

    @abstractmethod
    def embed_text(self, text: str) -> List[float]:
        """
        Generate embedding for a single text.

        Args:
            text: The text to embed.

        Returns:
            List of floats representing the embedding vector.
        """
        pass

    @abstractmethod
    def embed_documents(self, documents: List[str]) -> List[List[float]]:
        """
        Generate embeddings for multiple documents.

        Args:
            documents: List of texts to embed.

        Returns:
            List of embedding vectors.
        """
        pass
