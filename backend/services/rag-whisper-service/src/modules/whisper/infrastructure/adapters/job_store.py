from typing import Optional, Dict
from src.modules.whisper.domain.ports import JobStorePort, TranscriptionJob

class InMemoryJobStore(JobStorePort):
    def __init__(self):
        self.jobs: Dict[str, TranscriptionJob] = {}

    async def save_job(self, job: TranscriptionJob):
        self.jobs[job.id] = job

    async def get_job(self, job_id: str) -> Optional[TranscriptionJob]:
        return self.jobs.get(job_id)
