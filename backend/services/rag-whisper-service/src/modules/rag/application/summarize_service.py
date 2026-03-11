"""
Summarization Service.

Responsible for summarizing text using LLM.
Follows Single Responsibility Principle (SRP).
"""
from typing import Optional

from src.modules.rag.domain.ports import LLMClientPort
from src.modules.rag.application.prompts import RAGPrompts


class SummarizeService:
    """Service for summarizing text."""

    def __init__(self, default_llm: LLMClientPort):
        """
        Initialize the summarization service.

        Args:
            default_llm: Default LLM client for summarization.
        """
        self.default_llm = default_llm

    async def summarize(
        self,
        text: str,
        llm_client: Optional[LLMClientPort] = None,
        detailed: bool = False
    ) -> str:
        """
        Summarize the given text.

        Args:
            text: The text to summarize.
            llm_client: Optional LLM client override.
            detailed: Whether to generate a detailed summary.

        Returns:
            Generated summary.
        """
        client = llm_client or self.default_llm
        prompt = RAGPrompts.format_summarize(text, detailed)
        return await client.generate_response(prompt)
