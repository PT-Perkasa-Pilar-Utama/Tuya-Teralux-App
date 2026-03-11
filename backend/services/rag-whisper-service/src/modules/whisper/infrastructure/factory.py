from typing import Optional
from src.modules.whisper.domain.ports import WhisperClientPort
from src.modules.whisper.infrastructure.adapters.whisper import OpenAIWhisperAdapter, GroqWhisperAdapter, GeminiWhisperAdapter, OrionWhisperAdapter, LocalWhisperAdapter
from src.modules.shared.infrastructure.registry import registry

class WhisperAdapterFactory:
    @staticmethod
    def create_whisper(model_id: str) -> WhisperClientPort:
        config = registry.get_whisper(model_id)
        if not config:
            raise ValueError(f"Whisper model {model_id} not found in registry")
        
        provider = config.provider.lower()
        model_name = config.model_name

        if provider == "openai":
            return OpenAIWhisperAdapter(model_name)
        elif provider == "groq":
            return GroqWhisperAdapter(model_name)
        elif provider == "gemini":
            return GeminiWhisperAdapter(model_name)
        elif provider == "orion":
            return OrionWhisperAdapter(model_name)
        elif provider == "local":
            return LocalWhisperAdapter(model_name)
        else:
            raise NotImplementedError(f"Provider {provider} not supported for Whisper yet")
