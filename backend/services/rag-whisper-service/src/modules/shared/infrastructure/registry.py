import yaml
from pathlib import Path
from typing import Dict, List, Optional
from pydantic import BaseModel

class ModelConfig(BaseModel):
    id: str
    provider: str
    model_name: str
    description: Optional[str] = None

class ModelRegistry:
    def __init__(self, config_path: str):
        self.config_path = config_path
        self.llm_models: Dict[str, ModelConfig] = {}
        self.embedding_models: Dict[str, ModelConfig] = {}
        self.whisper_models: Dict[str, ModelConfig] = {}
        self.load_config()

    def load_config(self):
        with open(self.config_path, 'r') as f:
            config = yaml.safe_load(f)
            
        for m in config.get('llm_models', []):
            self.llm_models[m['id']] = ModelConfig(**m)
        for m in config.get('embedding_models', []):
            self.embedding_models[m['id']] = ModelConfig(**m)
        for m in config.get('whisper_models', []):
            self.whisper_models[m['id']] = ModelConfig(**m)

    def get_llm(self, model_id: str) -> Optional[ModelConfig]:
        return self.llm_models.get(model_id)

    def get_embedding(self, model_id: str) -> Optional[ModelConfig]:
        return self.embedding_models.get(model_id)

    def get_whisper(self, model_id: str) -> Optional[ModelConfig]:
        return self.whisper_models.get(model_id)

# Global registry instance
config_dir = Path(__file__).parent.parent / "config"
registry = ModelRegistry(str(config_dir / "models.yaml"))
