from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import Optional

class Settings(BaseSettings):
    # FastAPI Settings
    PORT: int = 8000
    HOST: str = "0.0.0.0"
    DEBUG: bool = True

    # gRPC Settings
    GRPC_PORT: int = 50051

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

    # Whisper Settings
    WHISPER_PROVIDER: str = "openai"  # openai, groq, gemini, orion, local
    OPENAI_WHISPER_MODEL: str = "whisper-1"
    GROQ_WHISPER_MODEL: str = "whisper-large-v3"
    GEMINI_MODEL: str = "gemini-1.5-flash"
    ORION_MODEL: str = "orion-whisper"
    WHISPER_MODEL_PATH: Optional[str] = None
    WHISPER_TEMP_DIR: str = "tmp/uploads"

    # Default Models
    DEFAULT_LLM_MODEL: str = "openai/gpt-4o"
    DEFAULT_EMBEDDING_MODEL: str = "openai/text-embedding-3-small"
    DEFAULT_WHISPER_MODEL: str = "openai/whisper-1"

    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

settings = Settings()
