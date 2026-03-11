from typing import List, Dict, Any
from langchain_community.vectorstores import Milvus
from src.modules.rag.domain.ports.vector_store_port import VectorStorePort
from src.modules.rag.domain.ports.document import DocumentDomain
from src.modules.shared.infrastructure.settings import settings
from langchain_core.documents import Document

class MilvusVectorStoreAdapter(VectorStorePort):
    """
    Milvus vector store adapter implementing VectorStorePort.
    
    Provides persistent vector storage and similarity search capabilities.
    """
    
    def __init__(self, embedding_function):
        """
        Initialize Milvus vector store.
        
        Args:
            embedding_function: Embedding function for vector generation.
        """
        self.vector_store = Milvus(
            embedding_function=embedding_function,
            connection_args={
                "host": settings.MILVUS_HOST,
                "port": str(settings.MILVUS_PORT)
            },
            collection_name=settings.MILVUS_COLLECTION,
            auto_id=True
        )

    async def add_documents(self, documents: List[DocumentDomain]) -> None:
        """Add documents to the Milvus vector store."""
        langchain_docs = [
            Document(page_content=d.page_content, metadata=d.metadata)
            for d in documents
        ]
        self.vector_store.add_documents(langchain_docs)

    async def similarity_search(
        self,
        query: str,
        k: int = 4
    ) -> List[DocumentDomain]:
        """Perform similarity search in Milvus."""
        results = self.vector_store.similarity_search(query, k=k)
        return [
            DocumentDomain(page_content=r.page_content, metadata=r.metadata)
            for r in results
        ]
