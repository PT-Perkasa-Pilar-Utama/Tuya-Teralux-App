"""
Translation Service.

Responsible for translating text using LLM.
Follows Single Responsibility Principle (SRP).
"""
from typing import Optional

from src.modules.rag.domain.ports import LLMClientPort
from src.modules.rag.application.prompts import RAGPrompts


class TranslateService:
    """Service for translating text."""

    def __init__(self, default_llm: LLMClientPort):
        """
        Initialize the translation service.

        Args:
            default_llm: Default LLM client for translation.
        """
        self.default_llm = default_llm

    async def translate(
        self,
        text: str,
        target_lang: str,
        llm_client: Optional[LLMClientPort] = None
    ) -> str:
        """
        Translate text to target language.

        Args:
            text: The text to translate.
            target_lang: Target language for translation.
            llm_client: Optional LLM client override.

        Returns:
            Translated text.
        """
        client = llm_client or self.default_llm
        prompt = RAGPrompts.format_translate(text, target_lang)
        return await client.generate_response(prompt)
