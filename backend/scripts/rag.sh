#!/usr/bin/env bash
set -euo pipefail

# Helper: authenticate (X-API-KEY) -> submit RAG request -> poll until done -> print result
# Usage: ./scripts/rag.sh "turn on the lamp" [API_KEY] [BASE_URL]

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: 'jq' is required. Install it and retry."
  exit 1
fi

TEXT="${1:-}"
API_KEY_ARG="${2:-}"
BASE_URL="${3:-http://localhost:8081}"

if [ -z "$TEXT" ]; then
  echo "Usage: $0 \"text to process\" [API_KEY] [BASE_URL]"
  exit 1
fi

# Try to get API key from arg or .env.dev/.env
API_KEY="$API_KEY_ARG"
if [ -z "$API_KEY" ]; then
  if [ -f .env.dev ]; then
    API_KEY=$(grep '^API_KEY=' .env.dev | head -n1 | sed 's/API_KEY=//') || true
  fi
fi
if [ -z "$API_KEY" ] && [ -f .env ]; then
  API_KEY=$(grep '^API_KEY=' .env | head -n1 | sed 's/API_KEY=//') || true
fi

if [ -z "$API_KEY" ]; then
  echo "Error: API_KEY not provided. Pass as second arg or set in .env/.env.dev" >&2
  exit 1
fi

echo "Authenticating against $BASE_URL/api/tuya/auth..."
AUTH_RES=$(curl -s -H "X-API-KEY: $API_KEY" "$BASE_URL/api/tuya/auth")
TOKEN=$(echo "$AUTH_RES" | jq -r '.data.access_token // empty')
if [ -z "$TOKEN" ]; then
  echo "Failed to authenticate. Response:" >&2
  echo "$AUTH_RES" | jq -C .
  exit 1
fi

echo "Token obtained: ${TOKEN}"

echo "Submitting RAG request..."
SUB_RES=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X POST "$BASE_URL/api/rag" -d "{\"text\":\"${TEXT//"/\"}\"}")
TASK_ID=$(echo "$SUB_RES" | jq -r '.data.task_id // empty')
if [ -z "$TASK_ID" ]; then
  echo "Failed to submit RAG request. Response:" >&2
  echo "$SUB_RES" | jq -C .
  exit 1
fi

echo "Task submitted: $TASK_ID"

# Poll for result
echo -n "Polling task status"
for i in $(seq 1 120); do
  sleep 1
  echo -n "."
  RES=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/rag/$TASK_ID")
  STATUS=$(echo "$RES" | jq -r '.data.status // empty')
  if [ "$STATUS" = "done" ] || [ "$STATUS" = "error" ]; then
    echo "\nStatus: $STATUS"
    echo "Result:"
    # pretty print result field
    echo "$RES" | jq -r '.data.result'
    # try best-effort to extract JSON from result and pretty print
    RAW=$(echo "$RES" | jq -r '.data.result')
    if echo "$RAW" | grep -q '{'; then
      CAND=$(echo "$RAW" | tr -d '\r')
      # try to pretty print JSON inside the result
      echo "\nParsed JSON (best-effort):"
      echo "$CAND" | sed 's/\\n/\n/g' | awk 'BEGIN{ORS=""}{print $0}' | sed 's/\"/"/g' | jq -C . 2>/dev/null || echo "(not valid JSON)";
    fi
    exit 0
  fi

done

echo "\nTimeout waiting for task completion"
exit 1
