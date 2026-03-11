from typing import Optional
from src.modules.rag.domain.ports import LLMClientPort
from src.modules.rag.infrastructure.adapters.llm import OpenAIAdapter, GeminiAdapter, GroqAdapter, OrionAdapter, LocalLLMAdapter
from src.modules.shared.infrastructure.registry import registry

class LLMAdapterFactory:
    @staticmethod
    def create_llm(model_id: str) -> LLMClientPort:
        config = registry.get_llm(model_id)
        if not config:
            raise ValueError(f"LLM model {model_id} not found in registry")
        
        provider = config.provider.lower()
        model_name = config.model_name

        if provider == "openai":
            return OpenAIAdapter(model_name)
        elif provider == "gemini":
            return GeminiAdapter(model_name)
        elif provider == "groq":
            return GroqAdapter(model_name)
        elif provider == "orion":
            return OrionAdapter(model_name)
        elif provider == "local":
            return LocalLLMAdapter(model_name)
        else:
            raise NotImplementedError(f"Provider {provider} not supported for LLM yet")
