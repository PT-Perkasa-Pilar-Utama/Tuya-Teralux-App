from typing import List, Optional
from langchain_openai import OpenAIEmbeddings
from langchain_google_genai import GoogleGenerativeAIEmbeddings
from src.modules.rag.domain.ports import EmbeddingPort
from src.modules.shared.infrastructure.settings import settings

class OpenAIEmbeddingAdapter(EmbeddingPort):
    def __init__(self, model_name: str):
        self.embeddings = OpenAIEmbeddings(
            model=model_name,
            openai_api_key=settings.OPENAI_API_KEY
        )

    def embed_text(self, text: str) -> List[float]:
        return self.embeddings.embed_query(text)

    def embed_documents(self, documents: List[str]) -> List[List[float]]:
        return self.embeddings.embed_documents(documents)

class GeminiEmbeddingAdapter(EmbeddingPort):
    def __init__(self, model_name: str):
        self.embeddings = GoogleGenerativeAIEmbeddings(
            model=model_name,
            google_api_key=settings.GEMINI_API_KEY
        )

    def embed_text(self, text: str) -> List[float]:
        return self.embeddings.embed_query(text)

    def embed_documents(self, documents: List[str]) -> List[List[float]]:
        return self.embeddings.embed_documents(documents)
