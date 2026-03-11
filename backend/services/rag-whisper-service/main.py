from fastapi import FastAPI
from src.modules.rag.interfaces.rest.router import router as rag_router
from src.modules.whisper.interfaces.rest.router import router as whisper_router
import os

app = FastAPI(title="RAG-Whisper Service (Modular)", version="0.2.0")

# Register routers
app.include_router(rag_router)
app.include_router(whisper_router)

@app.get("/health")
async def health():
    return {"status": "ok"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
