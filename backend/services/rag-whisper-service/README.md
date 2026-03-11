# RAG-Whisper Service

FastAPI-based service for RAG (Retrieval-Augmented Generation) and Whisper transcription, integrated with LangChain and Milvus.

## Architecture

- **Framework**: FastAPI
- **RAG Engine**: LangChain
- **Vector Database**: Milvus
- **Transcription**: OpenAI Whisper (via providers)
- **Security**: JWT validation (shared secret with Go backend)

## Endpoints

### RAG
- `POST /v1/rag/ingest`: Ingest text documents into Milvus.
- `POST /v1/rag/query`: Query the vector store and get LLM-generated answers.

### Whisper
- `POST /v1/whisper/transcribe`: Upload audio files for transcription.

## Setup

1. Copy `.env.example` to `.env` and fill in API keys.
2. Install dependencies: `pip install .`
3. Run the service: `python main.py`

## Local Development (Milvus)

Use the provided `docker-compose.yml` to start a local Milvus instance:
```bash
docker-compose up -d
```
