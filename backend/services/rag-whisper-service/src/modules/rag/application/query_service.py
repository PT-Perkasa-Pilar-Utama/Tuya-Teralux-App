"""
Query Service for RAG operations.

Responsible for querying the vector store and generating responses.
Follows Single Responsibility Principle (SRP).
"""
from typing import Optional, List, Dict, Any

from src.modules.rag.domain.ports import VectorStorePort, LLMClientPort
from src.modules.rag.application.prompts import RAGPrompts


class QueryService:
    """Service for querying documents and generating responses."""

    def __init__(
        self,
        vector_store: VectorStorePort,
        default_llm: LLMClientPort,
        k: int = 4,
    ):
        """
        Initialize the query service.

        Args:
            vector_store: Vector store port for document retrieval.
            default_llm: Default LLM client for response generation.
            k: Number of documents to retrieve for context.
        """
        self.vector_store = vector_store
        self.default_llm = default_llm
        self.k = k
    
    async def query(
        self,
        question: str,
        llm_client: Optional[LLMClientPort] = None
    ) -> str:
        """
        Query the vector store and generate an answer.

        Args:
            question: The question to answer.
            llm_client: Optional LLM client override.

        Returns:
            Generated answer based on retrieved context.
        """
        client = llm_client or self.default_llm
        relevant_docs = await self.vector_store.similarity_search(
            question, k=self.k
        )

        context = RAGPrompts.build_context([d.page_content for d in relevant_docs])
        prompt = RAGPrompts.format_query(context, question)

        return await client.generate_response(prompt)
    
    async def chat(
        self,
        question: str,
        llm_client: Optional[LLMClientPort] = None,
        history: Optional[List[Dict[str, str]]] = None
    ) -> str:
        """
        Chat with the RAG system (conversational query).

        Args:
            question: The user's question.
            llm_client: Optional LLM client override.
            history: Optional conversation history.

        Returns:
            Generated conversational response.
        """
        client = llm_client or self.default_llm
        relevant_docs = await self.vector_store.similarity_search(
            question, k=self.k
        )
        context = RAGPrompts.build_context([d.page_content for d in relevant_docs])
        prompt = RAGPrompts.format_chat(context, question)

        return await client.generate_response(prompt, history=history)

    async def control(
        self,
        command: str,
        llm_client: Optional[LLMClientPort] = None
    ) -> str:
        """
        Analyze user command for device control.

        Args:
            command: The control command from user.
            llm_client: Optional LLM client override.

        Returns:
            Analyzed action and device logic.
        """
        client = llm_client or self.default_llm
        prompt = RAGPrompts.format_control(command)
        return await client.generate_response(prompt)
