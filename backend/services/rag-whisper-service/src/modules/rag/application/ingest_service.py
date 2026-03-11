"""
Document Ingestion Service.

Responsible for ingesting and chunking text documents into the vector store.
Follows Single Responsibility Principle (SRP).
"""
from typing import List, Optional, Dict, Any

from src.modules.rag.domain.ports import VectorStorePort, DocumentDomain
from src.modules.rag.domain.ports.text_splitter_port import TextSplitterPort


class IngestService:
    """Service for ingesting documents into the vector store."""

    def __init__(
        self,
        vector_store: VectorStorePort,
        text_splitter: TextSplitterPort,
    ):
        """
        Initialize the ingest service.

        Args:
            vector_store: Vector store port for document storage.
            text_splitter: Text splitter port for chunking.
        """
        self.vector_store = vector_store
        self.text_splitter = text_splitter
    
    async def ingest_text(
        self,
        text: str,
        metadata: Optional[Dict[str, Any]] = None
    ) -> int:
        """
        Ingest text into the vector store.
        
        Args:
            text: The text content to ingest.
            metadata: Optional metadata to associate with the document.
            
        Returns:
            Number of chunks ingested.
        """
        chunks = self.text_splitter.split_text(text)
        docs = [
            DocumentDomain(page_content=chunk, metadata=metadata or {})
            for chunk in chunks
        ]
        await self.vector_store.add_documents(docs)
        return len(docs)
    
    async def ingest_documents(
        self,
        documents: List[DocumentDomain]
    ) -> int:
        """
        Ingest pre-chunked documents into the vector store.
        
        Args:
            documents: List of document domains to ingest.
            
        Returns:
            Number of documents ingested.
        """
        await self.vector_store.add_documents(documents)
        return len(documents)
