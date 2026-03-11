from abc import ABC
from typing import List, Dict, Optional
import httpx
from langchain_openai import ChatOpenAI
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_groq import ChatGroq
from src.modules.rag.domain.ports import LLMClientPort
from src.modules.shared.infrastructure.settings import settings

class OpenAIAdapter(LLMClientPort):
    def __init__(self, model_name: str):
        self.llm = ChatOpenAI(
            model_name=model_name,
            openai_api_key=settings.OPENAI_API_KEY
        )

    async def generate_response(self, prompt: str, history: Optional[List[Dict[str, str]]] = None) -> str:
        # History not yet fully integrated in ChatOpenAI call for simplicity
        response = self.llm.invoke(prompt)
        return response.content

class GeminiAdapter(LLMClientPort):
    def __init__(self, model_name: str):
        self.llm = ChatGoogleGenerativeAI(
            model=model_name,
            google_api_key=settings.GEMINI_API_KEY
        )

    async def generate_response(self, prompt: str, history: Optional[List[Dict[str, str]]] = None) -> str:
        response = self.llm.invoke(prompt)
        return response.content

class GroqAdapter(LLMClientPort):
    def __init__(self, model_name: str):
        self.llm = ChatGroq(
            model_name=model_name,
            groq_api_key=settings.GROQ_API_KEY
        )

    async def generate_response(self, prompt: str, history: Optional[List[Dict[str, str]]] = None) -> str:
        response = self.llm.invoke(prompt)
        return response.content

class OrionAdapter(LLMClientPort):
    # Custom adapter since LangChain might not have direct Orion support
    def __init__(self, model_name: str):
        self.model_name = model_name

    async def generate_response(self, prompt: str, history: Optional[List[Dict[str, str]]] = None) -> str:
        # Placeholder for Orion REST API call
        # Mocking for now as typical for Orion
        return f"Orion ({self.model_name}) response to: {prompt}"

class LocalLLMAdapter(LLMClientPort):
    def __init__(self, model_name: str):
        self.model_name = model_name

    async def generate_response(self, prompt: str, history: Optional[List[Dict[str, str]]] = None) -> str:
        # Placeholder for local LLM (llama.cpp/ollama)
        return f"Local ({self.model_name}) response to: {prompt}"
