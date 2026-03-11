#!/usr/bin/env bash
set -Eeuo pipefail

# =========================
# Professional AI Assistant TTS
# =========================

TEXT="${1:-Halo! Saya Sensho. Ada yang bisa saya bantu?}"
OUTPUT_FILE="${2:-greeting_sensio_pro_assistant.mp3}"
MODEL="${MODEL:-gemini-2.5-flash-preview-tts}"

# Persona / delivery instruction
STYLE_INSTRUCTION=$(
cat <<'EOF'
Speak as a highly professional AI assistant.
Voice characteristics:
- calm, composed, and confident
- clear articulation and precise pronunciation
- neutral-professional tone
- efficient, polished, and trustworthy
- warm but restrained; never overly emotional, dramatic, or playful
- avoid sing-song cadence, exaggerated emphasis, or sales-like enthusiasm
- maintain steady pacing with short natural pauses
- sound like an executive virtual assistant delivering concise, high-quality help

For Indonesian text:
- use natural, standard Indonesian pronunciation
- keep diction formal, clean, and modern
- avoid sounding theatrical, robotic, or overly soft
EOF
)

# Optional voice name if supported by the API/model
VOICE_NAME="${VOICE_NAME:-Kore}"

# Temp files
RAW_FILE="$(mktemp /tmp/sensio_tts_XXXXXX.raw)"
JSON_FILE="$(mktemp /tmp/sensio_tts_XXXXXX.json)"

cleanup() {
  rm -f "$RAW_FILE" "$JSON_FILE"
}
trap cleanup EXIT

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Error: required command not found: $1" >&2
    exit 1
  }
}

require_cmd curl
require_cmd jq
require_cmd base64
require_cmd ffmpeg

# Load .env safely if present
if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

: "${GEMINI_API_KEY:?Error: GEMINI_API_KEY is not set. Put it in .env or export it first.}"

echo "Generating professional AI assistant TTS..."
echo "Text   : $TEXT"
echo "Output : $OUTPUT_FILE"
echo "Model  : $MODEL"
echo "Voice  : $VOICE_NAME"

# Build JSON safely with jq
REQUEST_BODY="$(
  jq -n \
    --arg style "$STYLE_INSTRUCTION" \
    --arg text "$TEXT" \
    --arg voice "$VOICE_NAME" \
    '{
      contents: [
        {
          role: "user",
          parts: [
            {
              text: ("Task: Generate speech audio only.\n\nStyle guide:\n" + $style + "\n\nText to speak:\n" + $text)
            }
          ]
        }
      ],
      generationConfig: {
        responseModalities: ["AUDIO"],
        speechConfig: {
          voiceConfig: {
            prebuiltVoiceConfig: {
              voiceName: $voice
            }
          }
        }
      }
    }'
)"

curl -sS \
  "https://generativelanguage.googleapis.com/v1beta/models/${MODEL}:generateContent?key=${GEMINI_API_KEY}" \
  -X POST \
  -H 'Content-Type: application/json' \
  -d "$REQUEST_BODY" \
  > "$JSON_FILE"

# Show API error clearly
if jq -e '.error' "$JSON_FILE" >/dev/null 2>&1; then
  echo "API Error:" >&2
  jq '.error' "$JSON_FILE" >&2
  exit 1
fi

# Extract base64 PCM audio
AUDIO_B64="$(jq -r '.candidates[0].content.parts[]?.inlineData.data // empty' "$JSON_FILE" | head -n 1)"

if [[ -z "$AUDIO_B64" ]]; then
  echo "Error: No audio data returned by API." >&2
  echo "Raw response excerpt:" >&2
  jq '{candidates, promptFeedback}' "$JSON_FILE" >&2 || cat "$JSON_FILE" >&2
  exit 1
fi

printf '%s' "$AUDIO_B64" | base64 --decode > "$RAW_FILE"

# Convert raw PCM to MP3
ffmpeg -y -hide_banner -loglevel error \
  -f s16le -ar 24000 -ac 1 \
  -i "$RAW_FILE" \
  -codec:a libmp3lame -b:a 128k \
  "$OUTPUT_FILE"

echo "Success! Saved to $OUTPUT_FILE"