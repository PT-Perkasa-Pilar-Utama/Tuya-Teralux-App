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

# If no TEXT arg provided, try RAG_TEXT env var or STDIN (allow piping)
if [ -z "$TEXT" ]; then
  if [ -n "${RAG_TEXT:-}" ]; then
    TEXT="$RAG_TEXT"
  elif [ ! -t 0 ]; then
    # read from stdin (supports piping)
    TEXT=$(cat -)
  fi
fi

# If still empty, try to find RAG_TEXT in .env files (current dir, parents, backend/)
if [ -z "$TEXT" ]; then
  for f in .env; do
    dir="."
    for i in 0 1 2 3; do
      if [ -f "$dir/$f" ]; then
        val=$(grep '^RAG_TEXT=' "$dir/$f" | head -n1 | sed 's/RAG_TEXT=//') || true
        if [ -n "$val" ]; then
          TEXT="$val"
          break 2
        fi
      fi
      dir="../$dir"
    done
  done
fi

if [ -z "$TEXT" ]; then
  # also try backend/.env file
  if [ -f "backend/.env" ]; then
    TEXT=$(grep '^RAG_TEXT=' backend/.env | head -n1 | sed 's/RAG_TEXT=//') || true
  fi
fi

# If still empty, fall back to interactive prompt
if [ -z "$TEXT" ]; then
  echo -n "Enter text for RAG (or press Ctrl+C to cancel): "
  read -r TEXT
fi

if [ -z "$TEXT" ]; then
  echo "No text provided. Exiting." >&2
  exit 1
fi

# If the second argument looks like a URL (starts with http), treat it as BASE_URL
if [ -n "$API_KEY_ARG" ] && echo "$API_KEY_ARG" | grep -qE '^https?://'; then
  BASE_URL="$API_KEY_ARG"
  API_KEY_ARG=""
fi

# Resolve API_KEY precedence: CLI arg > exported ENV var > .env in parents
API_KEY=""
if [ -n "$API_KEY_ARG" ]; then
  API_KEY="$API_KEY_ARG"
elif [ -n "${API_KEY:-}" ]; then
  API_KEY="${API_KEY:-}"
fi

# Search for .env up to 3 parent levels if API_KEY still empty
if [ -z "$API_KEY" ]; then
  for f in .env; do
    dir="."
    for i in 0 1 2 3; do
      if [ -f "$dir/$f" ]; then
        val=$(grep '^API_KEY=' "$dir/$f" | head -n1 | sed 's/API_KEY=//') || true
        if [ -n "$val" ]; then
          API_KEY="$val"
          break 2
        fi
      fi
      dir="../$dir"
    done
  done
fi

# Also check backend/ for .env file if still not found (common when running from repo root)
if [ -z "$API_KEY" ] && [ -f "backend/.env" ]; then
  API_KEY=$(grep '^API_KEY=' backend/.env | head -n1 | sed 's/API_KEY=//') || true
fi

if [ -z "$API_KEY" ]; then
  echo "Error: API_KEY not provided. Pass as second arg or set in .env or export API_KEY env var" >&2
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
# Safely build JSON body using jq
BODY=$(jq -nc --arg text "$TEXT" '{text: $text}')
SUB_RES=$(curl -s -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -X POST "$BASE_URL/api/rag" -d "$BODY")
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
    STATUS=$(echo "$RES" | jq -r '.data.status.status // (.data.status // empty)')
    if [ "$STATUS" = "done" ] || [ "$STATUS" = "error" ]; then
      echo "\nStatus: $STATUS"
      echo "Result (status DTO):"
      # pretty print the status DTO if available
      echo "$RES" | jq -r '.data.status'

      # fallback: also print legacy .data.result if present
      RAW_LEGACY=$(echo "$RES" | jq -r '.data.result // empty')
      if [ -n "$RAW_LEGACY" ]; then
        echo "\nLegacy Result field:\n$RAW_LEGACY"
      fi

      # Attempt to extract specific fields if structured
      ENDPOINT=$(echo "$RES" | jq -r '.data.status.endpoint // empty')
      METHOD=$(echo "$RES" | jq -r '.data.status.method // empty')
      BODY=$(echo "$RES" | jq -r '.data.status.body // empty')
      if [ -n "$ENDPOINT" ] || [ -n "$METHOD" ] || [ -n "$BODY" ]; then
        echo "\nLLM decision:"
        [ -n "$ENDPOINT" ] && echo "  endpoint: $ENDPOINT"
        [ -n "$METHOD" ] && echo "  method: $METHOD"
        if [ -n "$BODY" ]; then
          echo "  body:"
          echo "$BODY" | jq -C . 2>/dev/null || echo "    $BODY"
        fi
      fi

  fi

done

echo "\nTimeout waiting for task completion"
exit 1
