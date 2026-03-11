from typing import Optional
from src.modules.rag.domain.ports import EmbeddingPort
from src.modules.rag.infrastructure.adapters.embeddings import OpenAIEmbeddingAdapter, GeminiEmbeddingAdapter
from src.modules.shared.infrastructure.registry import registry

class EmbeddingAdapterFactory:
    @staticmethod
    def create_embedding(model_id: str) -> EmbeddingPort:
        config = registry.get_embedding(model_id)
        if not config:
            raise ValueError(f"Embedding model {model_id} not found in registry")
        
        provider = config.provider.lower()
        model_name = config.model_name

        if provider == "openai":
            return OpenAIEmbeddingAdapter(model_name)
        elif provider == "gemini":
            return GeminiEmbeddingAdapter(model_name)
        else:
            raise NotImplementedError(f"Provider {provider} not supported for embeddings yet")
