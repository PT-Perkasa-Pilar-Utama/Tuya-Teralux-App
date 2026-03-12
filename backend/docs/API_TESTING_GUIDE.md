# API Testing Guide

This guide explains how to test the Teralux API using Scalar API Reference or Postman with environment variables.

## 🚀 Scalar API Reference

### Access Scalar API Reference

Once the backend is running, open your browser and navigate to:

```
http://localhost:8080/openapi
```

Or if using alternate port:
```
http://localhost:8081/openapi
```

Scalar API Reference provides:
- Modern, beautiful dark theme documentation
- Interactive sidebar navigation
- Try-it-out functionality for all endpoints
- Server selector to switch between ports
- Authentication helpers for API Key and Bearer token
- Clean, professional layout

## 🔧 Environment Variables

### Setup for Testing

1. **Copy `.env.example` to `.env`**:
   ```bash
   cp .env.example .env
   ```

2. **Fill in the test variables**:
   ```bash
   # Required
   API_KEY=your-api-key-here
   JWT_SECRET=your-jwt-secret-here
   
   # Optional (for testing)
   TEST_API_KEY=your-api-key-here
   TEST_JWT_TOKEN=your-jwt-token-here
   TEST_TERMINAL_MAC=AA:BB:CC:DD:EE:FF
   ```

### Available Test Variables

| Variable | Description | Used For |
|----------|-------------|----------|
| `TEST_API_KEY` | API Key for public endpoints | `X-API-KEY` header |
| `TEST_JWT_TOKEN` | JWT Bearer token | `Authorization: Bearer` header |
| `TEST_TUYA_ACCESS_TOKEN` | Tuya access token | Tuya endpoints |
| `TEST_TERMINAL_MAC` | Terminal MAC address | Terminal lookup |
| `TEST_TERMINAL_ID` | Terminal ID | Terminal/Device/Scene endpoints |
| `TEST_DEVICE_ID` | Device ID | Device status endpoints |
| `TEST_SCENE_ID` | Scene ID | Scene control endpoints |
| `TEST_EMAIL` | Test email address | Mail endpoints |

## 🔐 Authentication

### API Key (Public Endpoints)
```
X-API-KEY: your-api-key
```

### Bearer Token (Protected Endpoints)
```
Authorization: Bearer your-jwt-token
```

## 📝 Quick Reference

### 01. Tuya
- `GET /api/tuya/auth` - Authenticate
- `GET /api/tuya/devices` - Get all devices
- `POST /api/tuya/devices/:id/commands/switch` - Send switch command

### 02. Terminal
- `POST /api/terminal` - Create terminal (API Key)
- `GET /api/terminal` - Get all terminals (API Key)
- `GET /api/terminal/mac/:mac` - Get by MAC (API Key)

### 02. Terminal - Devices
- `POST /api/devices` - Create device
- `GET /api/devices` - Get all devices
- `PUT /api/devices/:id/status` - Update status

### 03. Scenes
- `GET /api/scenes` - Get all scenes
- `POST /api/terminal/:id/scenes` - Create scene
- `GET /api/terminal/:id/scenes/:scene_id/control` - Control scene

### 04. Speech
- `POST /api/models/whisper/transcribe` - Transcribe audio

### 05. RAG
- `POST /api/models/rag/chat` - AI Chat
- `POST /api/models/rag/control` - AI Control
- `POST /api/models/rag/summary` - Generate summary

### 06. Models
- `GET /api/models` - Get available models
- `POST /api/models/gemini` - Gemini
- `POST /api/models/openai` - OpenAI

### 07. Recordings
- `GET /api/recordings` - List recordings
- `POST /api/recordings` - Create recording

### 08. Mail
- `POST /api/mail/send` - Send email
- `POST /api/mail/send/mac/:mac` - Send by MAC

### 09. Common
- `GET /api/health` - Health check
- `DELETE /api/cache/flush` - Flush cache

## 💡 Tips

1. Use Scalar "Test Request" button to try endpoints directly
2. Set authentication tokens in the request interface
3. Save resource IDs in `.env` after creating them
4. Use response data from one endpoint as input for another
