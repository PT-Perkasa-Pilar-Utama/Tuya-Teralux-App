package contracts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensio/domain/models/rag/dtos"
)

func TestRAGChatRequestDTOContract(t *testing.T) {
	dto := dtos.RAGChatRequestDTO{
		RequestID:  "req-123",
		Prompt:     "Nyalakan AC di ruang meeting",
		Language:   "id",
		TerminalID: "term-123",
		UID:        "uid-456",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "req-123", decoded["request_id"])
	assert.Equal(t, "Nyalakan AC di ruang meeting", decoded["prompt"])
	assert.Equal(t, "id", decoded["language"])
	assert.Equal(t, "term-123", decoded["terminal_id"])
	assert.Equal(t, "uid-456", decoded["uid"])
}

func TestRAGChatResponseDTOContract(t *testing.T) {
	dto := dtos.RAGChatResponseDTO{
		Response:   "AC telah dinyalakan",
		IsControl:  true,
		IsBlocked:  false,
		RequestID:  "req-123",
		Source:     "HTTP_HANDLER",
		InstanceID: "instance-1",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "AC telah dinyalakan", decoded["response"])
	assert.Equal(t, true, decoded["is_control"])
	assert.Equal(t, false, decoded["is_blocked"])
	assert.Equal(t, "HTTP_HANDLER", decoded["source"])
}

func TestRAGSummaryRequestDTOContract(t *testing.T) {
	dto := dtos.RAGSummaryRequestDTO{
		Text:         "Meeting transcript...",
		Language:     "id",
		Context:      "technical meeting",
		Style:        "minutes",
		Location:     "Jakarta",
		Date:         "2026-04-15",
		Participants: []string{"Alice", "Bob", "Charlie"},
		MacAddress:   "AA:BB:CC:DD:EE:FF",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Meeting transcript...", decoded["text"])
	assert.Equal(t, "id", decoded["language"])
	assert.Equal(t, "technical meeting", decoded["context"])
	assert.Equal(t, "minutes", decoded["style"])
	assert.Equal(t, "Jakarta", decoded["location"])
	assert.Equal(t, "2026-04-15", decoded["date"])
}

func TestRAGStatusDTOContract(t *testing.T) {
	dto := dtos.RAGStatusDTO{
		Status:  "completed",
		Result:  "Summary text...",
		Summary: "Meeting Summary: ...",
		PDFUrl:  "https://example.com/summary.pdf",
		Error:   "",
		Trigger: "/api/models/rag/summary",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "completed", decoded["status"])
	assert.Equal(t, "Summary text...", decoded["result"])
	assert.Equal(t, "Meeting Summary: ...", decoded["summary"])
	assert.Equal(t, "https://example.com/summary.pdf", decoded["pdf_url"])
}

func TestRAGRequestDTOContract(t *testing.T) {
	dto := dtos.RAGRequestDTO{
		Text:       "Translate this text",
		Language:   "id",
		MacAddress: "AA:BB:CC:DD:EE:FF",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Translate this text", decoded["text"])
	assert.Equal(t, "id", decoded["language"])
}

func TestRAGControlRequestDTOContract(t *testing.T) {
	dto := dtos.RAGControlRequestDTO{
		Prompt:     "Turn on the lights",
		TerminalID: "term-123",
	}

	data, err := json.Marshal(dto)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "Turn on the lights", decoded["prompt"])
	assert.Equal(t, "term-123", decoded["terminal_id"])
}
