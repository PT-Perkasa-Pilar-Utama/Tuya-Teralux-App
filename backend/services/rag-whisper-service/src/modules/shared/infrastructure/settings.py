from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import Optional

class Settings(BaseSettings):
    # FastAPI Settings
    PORT: int = 8000
    HOST: str = "0.0.0.0"
    DEBUG: bool = True

    # Security
    JWT_SECRET: str
    JWT_ALGORITHM: str = "HS256"

    # Milvus Settings
    MILVUS_HOST: str = "localhost"
    MILVUS_PORT: int = 19530
    MILVUS_COLLECTION: str = "rag_docs"

    # API Keys
    OPENAI_API_KEY: Optional[str] = None
    GEMINI_API_KEY: Optional[str] = None
    GROQ_API_KEY: Optional[str] = None
    ORION_API_KEY: Optional[str] = None

    # Default Models
    DEFAULT_LLM_MODEL: str = "openai/gpt-4o"
    DEFAULT_EMBEDDING_MODEL: str = "openai/text-embedding-3-small"
    DEFAULT_WHISPER_MODEL: str = "openai/whisper-1"

    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

settings = Settings()
