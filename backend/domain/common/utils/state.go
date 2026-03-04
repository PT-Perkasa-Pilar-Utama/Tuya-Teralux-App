package utils

import "sync"

// ActiveTranscriptions tracks which Terminal IDs currently have an active transcription task.
// This is used to debounce dual signals (audio + text) from hardware devices.
var ActiveTranscriptions sync.Map
