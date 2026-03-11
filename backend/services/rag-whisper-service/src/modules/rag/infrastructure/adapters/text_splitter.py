"""
LangChain Text Splitter Adapter.

Implements TextSplitterPort using LangChain's RecursiveCharacterTextSplitter.
This moves the LangChain dependency from application layer to infrastructure.
"""
from typing import List

from langchain_text_splitters import RecursiveCharacterTextSplitter

from src.modules.rag.domain.ports.text_splitter_port import TextSplitterPort


class LangChainTextSplitterAdapter(TextSplitterPort):
    """
    LangChain-based text splitter implementation.

    Uses RecursiveCharacterTextSplitter for intelligent text chunking
    with configurable chunk size and overlap.
    """

    def __init__(self, chunk_size: int = 1000, chunk_overlap: int = 200):
        """
        Initialize the LangChain text splitter.

        Args:
            chunk_size: Maximum size of each text chunk.
            chunk_overlap: Number of characters to overlap between chunks.
        """
        self.splitter = RecursiveCharacterTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap,
            length_function=len,
            add_start_index=False,
        )

    def split_text(self, text: str) -> List[str]:
        """
        Split text into chunks using LangChain's splitter.

        Args:
            text: The text to split.

        Returns:
            List of text chunks.
        """
        return self.splitter.split_text(text)
