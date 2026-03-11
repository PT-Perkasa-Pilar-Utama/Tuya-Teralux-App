FROM python:3.10-slim

WORKDIR /app

# Install system dependencies for audio processing if needed
RUN apt-get update && apt-get install -y \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

COPY pyproject.toml .
# Create dummy requirements to use pip install if pyproject.toml is not enough
RUN pip install .

COPY . .

# Ensure tmp dir exists
RUN mkdir -p tmp/uploads

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
