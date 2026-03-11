from abc import ABC, abstractmethod
from typing import List, Optional, Dict, Any


class LLMClientPort(ABC):
    """Port defining the interface for LLM clients."""

    @abstractmethod
    async def generate_response(
        self,
        prompt: str,
        history: Optional[List[Dict[str, str]]] = None
    ) -> str:
        """
        Generate response from LLM.

        Args:
            prompt: The input prompt/text to generate response for.
            history: Optional conversation history for context.

        Returns:
            Generated response as string.
        """
        pass
