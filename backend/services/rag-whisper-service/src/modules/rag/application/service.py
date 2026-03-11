"""
Legacy RAG Application Service.

DEPRECATED: This monolithic service violates SRP.
Use the specialized services instead:
- IngestService for document ingestion
- QueryService for queries/chat
- TranslateService for translation
- SummarizeService for summarization

This service is kept for backward compatibility only.
"""
from typing import List, Optional, Dict, Any

from src.modules.rag.domain.ports import VectorStorePort, LLMClientPort, DocumentDomain
from src.modules.rag.domain.ports.text_splitter_port import TextSplitterPort
from src.modules.rag.application.prompts import RAGPrompts


class RAGApplicationService:
    """Legacy monolithic RAG service (DEPRECATED)."""

    def __init__(
        self,
        vector_store: VectorStorePort,
        default_llm: LLMClientPort,
        text_splitter: Optional[TextSplitterPort] = None,
    ):
        self.vector_store = vector_store
        self.default_llm = default_llm
        # Default text splitter if not provided
        from src.modules.rag.infrastructure.adapters.text_splitter import LangChainTextSplitterAdapter
        self.text_splitter = text_splitter or LangChainTextSplitterAdapter()

    async def ingest_text(self, text: str, metadata: Optional[Dict[str, Any]] = None):
        chunks = self.text_splitter.split_text(text)
        docs = [DocumentDomain(page_content=chunk, metadata=metadata or {}) for chunk in chunks]
        await self.vector_store.add_documents(docs)
        return len(docs)

    async def query(self, question: str, llm_client: Optional[LLMClientPort] = None) -> str:
        client = llm_client or self.default_llm
        relevant_docs = await self.vector_store.similarity_search(question)
        context = RAGPrompts.build_context([d.page_content for d in relevant_docs])
        prompt = RAGPrompts.format_query(context, question)
        return await client.generate_response(prompt)

    async def translate(self, text: str, target_lang: str, llm_client: Optional[LLMClientPort] = None) -> str:
        client = llm_client or self.default_llm
        prompt = RAGPrompts.format_translate(text, target_lang)
        return await client.generate_response(prompt)

    async def summarize(self, text: str, llm_client: Optional[LLMClientPort] = None) -> str:
        client = llm_client or self.default_llm
        prompt = RAGPrompts.format_summarize(text)
        return await client.generate_response(prompt)

    async def chat(self, question: str, llm_client: Optional[LLMClientPort] = None) -> str:
        client = llm_client or self.default_llm
        relevant_docs = await self.vector_store.similarity_search(question)
        context = RAGPrompts.build_context([d.page_content for d in relevant_docs])
        prompt = RAGPrompts.format_chat(context, question)
        return await client.generate_response(prompt)

    async def control(self, command: str, llm_client: Optional[LLMClientPort] = None) -> str:
        client = llm_client or self.default_llm
        prompt = RAGPrompts.format_control(command)
        return await client.generate_response(prompt)
