"""
Helper functions for LLM client management in REST routers.

Follows DRY principle by centralizing common LLM client logic.
"""
from typing import Optional

from src.modules.rag.domain.ports.llm_port import LLMClientPort
from src.modules.rag.infrastructure.factory import LLMAdapterFactory


async def get_llm_client(
    model_id: Optional[str] = None
) -> Optional[LLMClientPort]:
    """
    Get LLM client by model ID.
    
    Args:
        model_id: Optional model identifier. If None, returns None.
        
    Returns:
        LLM client instance or None.
    """
    if model_id:
        return LLMAdapterFactory.create_llm(model_id)
    return None


def create_llm_client_sync(
    model_id: Optional[str] = None
) -> Optional[LLMClientPort]:
    """
    Synchronous version of get_llm_client.
    
    Args:
        model_id: Optional model identifier. If None, returns None.
        
    Returns:
        LLM client instance or None.
    """
    if model_id:
        return LLMAdapterFactory.create_llm(model_id)
    return None
