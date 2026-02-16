package usecases

import (
	"encoding/json"
	"fmt"
	"strings"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/utilities"
)

type ChatUseCase interface {
	Chat(uid, teraluxID, prompt, language string) (*dtos.RAGChatResponseDTO, error)
}

type ChatUseCaseImpl struct {
	llm         utilities.LLMClient
	config      *utils.Config
	badger      *infrastructure.BadgerService
	controlUC   ControlUseCase
	translateUC TranslateUseCase
}

func NewChatUseCase(llm utilities.LLMClient, cfg *utils.Config, badger *infrastructure.BadgerService, controlUC ControlUseCase, translateUC TranslateUseCase) ChatUseCase {
	return &ChatUseCaseImpl{
		llm:         llm,
		config:      cfg,
		badger:      badger,
		controlUC:   controlUC,
		translateUC: translateUC,
	}
}

func (u *ChatUseCaseImpl) Chat(uid, teraluxID, prompt, language string) (*dtos.RAGChatResponseDTO, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	model := u.config.LLMModel
	if model == "" {
		model = "default"
	}

	// 1. Get History
	historyKey := fmt.Sprintf("chat_history:%s", teraluxID)
	var history []string
	if u.badger != nil {
		data, _ := u.badger.Get(historyKey)
		if data != nil {
			_ = json.Unmarshal(data, &history)
		}
	}

	// 2. Classify the prompt using LLM
	historyContext := ""
	if len(history) > 0 {
		historyContext = "Previous conversation:\n" + strings.Join(history, "\n") + "\n"
	}

	classificationPrompt := fmt.Sprintf(`You are an AI Assistant for a Smart Home system (Sensio).
Classify the following user prompt as either "CONTROL" or "CHAT".

"CONTROL" means the user wants to perform an action on a device (e.g., "Turn on AC", "Switch off light") or is answering a clarification question about a previous control command (e.g., "Bedroom AC", "Living room").
"CHAT" means the user is just talking, asking a general question, or greeting.

%sUser Prompt: "%s"

Classification (Only return "CONTROL" or "CHAT"):`, historyContext, prompt)

	classification, err := u.llm.CallModel(classificationPrompt, model)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	classification = strings.ToUpper(strings.TrimSpace(classification))

	if strings.Contains(classification, "CONTROL") {
		// Call ControlUseCase to get the structured response (including potential clarification)
		controlRes, err := u.controlUC.ProcessControl(uid, teraluxID, prompt)
		if err != nil {
			return nil, fmt.Errorf("control process failed: %w", err)
		}

		cleanResponse := controlRes.Message

		// If language is "id", translate the success message
		if strings.ToLower(language) == "id" && u.translateUC != nil {
			translated, err := u.translateUC.TranslateTextSync(cleanResponse, "id")
			if err == nil {
				cleanResponse = translated
			} else {
				utils.LogError("ChatUseCase: Failed to translate control response: %v", err)
			}
		}

		// Update History for CONTROL command
		if u.badger != nil {
			history = append(history, "User: "+prompt)
			history = append(history, "Assistant: "+cleanResponse)
			if len(history) > 20 {
				history = history[len(history)-20:]
			}
			data, _ := json.Marshal(history)
			_ = u.badger.Set(historyKey, data)
		}

		// Only redirect if a specific device ID was found
		var redirect *dtos.RedirectDTO
		if controlRes.DeviceID != "" {
			redirect = &dtos.RedirectDTO{
				Endpoint: "/api/rag/control",
				Method:   "POST",
				Body: dtos.RAGControlRequestDTO{
					Prompt:    prompt,
					TeraluxID: teraluxID,
				},
			}
		}

		return &dtos.RAGChatResponseDTO{
			Response:       cleanResponse,
			IsControl:      true,
			Redirect:       redirect,
			HTTPStatusCode: controlRes.HTTPStatusCode,
		}, nil
	}

	// 3. If it's CHAT, get the response from LLM with history
	historyContext = ""
	if len(history) > 0 {
		historyContext = "Previous conversation:\n" + strings.Join(history, "\n") + "\n"
	}

	chatPrompt := fmt.Sprintf(`You are Sensio AI Assistant, a helpful smart home companion.
Your tone is friendly, professional, and concise.
The user's preferred language is %s.

%sUser: %s
Assistant:`, language, historyContext, prompt)

	response, err := u.llm.CallModel(chatPrompt, model)
	if err != nil {
		return nil, fmt.Errorf("chat response failed: %w", err)
	}

	cleanResponse := strings.TrimSpace(response)

	// 4. Update History
	if u.badger != nil {
		history = append(history, "User: "+prompt)
		history = append(history, "Assistant: "+cleanResponse)
		// Limit to last 10 exchanges (20 messages)
		if len(history) > 20 {
			history = history[len(history)-20:]
		}
		data, _ := json.Marshal(history)
		_ = u.badger.Set(historyKey, data)
	}

	return &dtos.RAGChatResponseDTO{
		Response:  cleanResponse,
		IsControl: false,
	}, nil
}
