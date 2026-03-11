"""
Port defining the interface for text splitters.

This port allows the application layer to remain agnostic of the
specific text splitting implementation (LangChain, custom, etc.).
"""
from abc import ABC, abstractmethod
from typing import List


class TextSplitterPort(ABC):
    """Port defining the interface for text splitting strategies."""

    @abstractmethod
    def split_text(self, text: str) -> List[str]:
        """
        Split text into chunks.

        Args:
            text: The text to split.

        Returns:
            List of text chunks.
        """
        pass
