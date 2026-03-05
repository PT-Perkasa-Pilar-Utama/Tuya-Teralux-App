package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	commonDtos "sensio/domain/common/dtos"
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/utils"
	"sensio/domain/rag/dtos"
	"sensio/domain/rag/usecases"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RAGChatController struct {
	chatUC         usecases.ChatUseCase
	mqttSvc        *infrastructure.MqttService
	lastPrompt     map[string]string    // terminalID -> lastPrompt (deduplication)
	lastPromptTime map[string]time.Time // terminalID -> lastTime
	instanceID     string               // server start time identifier
	mu             sync.Mutex
}

func NewRAGChatController(chatUC usecases.ChatUseCase, mqttSvc *infrastructure.MqttService) *RAGChatController {
	return &RAGChatController{
		chatUC:         chatUC,
		mqttSvc:        mqttSvc,
		lastPrompt:     make(map[string]string),
		lastPromptTime: make(map[string]time.Time),
		instanceID:     time.Now().Format("2006-01-02 15:04:05"),
	}
}

func (c *RAGChatController) StartMqttSubscription() error {
	if c.mqttSvc == nil {
		return nil
	}

	config := utils.GetConfig()
	topic := fmt.Sprintf("users/+/%s/chat", config.ApplicationEnvironment)
	err := c.mqttSvc.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		correlationID := uuid.New().String()
		utils.LogInfo("[%s] RAGChat MQTT: Received message on %s, payload size: %d", correlationID, msg.Topic(), len(payload))
		if len(payload) == 0 {
			return
		}

		var req dtos.RAGChatRequestDTO
		// Strictly JSON unmarshalling as per user request
		err := json.Unmarshal(payload, &req)
		if err != nil {
			utils.LogError("[%s] RAGChat MQTT: Failed to unmarshal message: %v", correlationID, err)
			return
		}

		// Extract MAC from topic: (optionally $share/group/)users/MAC/env/chat
		topicParts := strings.Split(msg.Topic(), "/")
		mac := ""
		for i, part := range topicParts {
			if part == "users" && i+1 < len(topicParts) {
				mac = topicParts[i+1]
				break
			}
		}

		if req.Prompt == "" || req.TerminalID == "" {
			utils.LogError("[%s] RAGChat MQTT: Missing prompt or terminal_id", correlationID)
			if mac != "" {
				respTopic := fmt.Sprintf("users/%s/%s/chat/answer", mac, utils.GetConfig().ApplicationEnvironment)
				respData, _ := json.Marshal(commonDtos.StandardResponse{
					Status:  false,
					Message: "Validation Error",
					Details: []utils.ValidationErrorDetail{
						{Field: "prompt", Message: "prompt is required"},
						{Field: "terminal_id", Message: "terminal_id is required"},
					},
				})
				if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
					utils.LogError("RAGChat MQTT: Failed to publish validation error response: %v", err)
				}
			}
			return
		}

		// Debounce: if the Tuya terminal sent BOTH an audio file (whisper) AND text (chat) simultaneously,
		// we run the chat processing in a goroutine with a start delay. This prevents blocking the MQTT
		// client thread and gives the whisper handler enough time to download audio and set the active flag.
		go func(mac string, req dtos.RAGChatRequestDTO) {
			if mac != "" || req.TerminalID != "" {
				time.Sleep(500 * time.Millisecond)

				// 1. Check if Whisper is active (using both MAC from topic and TerminalID from payload)
				isWhisperActive := false
				if mac != "" {
					if _, active := utils.ActiveTranscriptions.Load(mac); active {
						isWhisperActive = true
					}
				}
				if !isWhisperActive && req.TerminalID != "" {
					if _, active := utils.ActiveTranscriptions.Load(req.TerminalID); active {
						isWhisperActive = true
					}
				}

				if isWhisperActive {
					utils.LogInfo("[%s] RAGChat MQTT: Dropping text query because a Whisper task is active for Terminal %s/%s", correlationID, mac, req.TerminalID)
					if mac != "" {
						respTopic := fmt.Sprintf("users/%s/%s/chat/answer", mac, utils.GetConfig().ApplicationEnvironment)
						respData, _ := json.Marshal(commonDtos.StandardResponse{
							Status:  true,
							Message: "Chat request received (sync with active whisper)",
						})
						_ = c.mqttSvc.Publish(respTopic, 0, false, respData)
					}
					return
				}

				// 2. Exact Deduplication (prevent processing the SAME prompt 3x)
				c.mu.Lock()
				terminalKey := req.TerminalID
				if terminalKey == "" {
					terminalKey = mac
				}
				last := c.lastPrompt[terminalKey]
				lastTime := c.lastPromptTime[terminalKey]
				now := time.Now()

				// If prompt is exactly same as last one and happened < 3 seconds ago, drop it.
				if last == req.Prompt && now.Sub(lastTime) < 3*time.Second {
					c.mu.Unlock()
					utils.LogInfo("[%s] RAGChat MQTT: Dropping duplicate prompt for %s: '%s'", correlationID, terminalKey, req.Prompt)
					if mac != "" {
						respTopic := fmt.Sprintf("users/%s/%s/chat/answer", mac, utils.GetConfig().ApplicationEnvironment)
						respData, _ := json.Marshal(commonDtos.StandardResponse{
							Status:  true,
							Message: "Chat request received (duplicate dropped)",
						})
						_ = c.mqttSvc.Publish(respTopic, 0, false, respData)
					}
					return
				}

				// Update cache
				c.lastPrompt[terminalKey] = req.Prompt
				c.lastPromptTime[terminalKey] = now
				c.mu.Unlock()
			}

			// Process chat
			requestID := correlationID // Reuse the early ingress correlation ID
			uid := req.UID
			if uid == "" {
				uid = utils.GetConfig().TuyaUserID
			}

			utils.LogInfo("[%s] RAGChat MQTT [Handler: StartMqttSubscription]: Starting chat process for UID: %s, Prompt: '%s'", requestID, uid, req.Prompt)
			res, err := c.chatUC.Chat(uid, req.TerminalID, req.Prompt, req.Language)
			if err != nil {
				utils.LogError("[%s] RAGChat MQTT: Chat processing failed: %v", requestID, err)
				if mac != "" {
					respTopic := fmt.Sprintf("users/%s/%s/chat/answer", mac, utils.GetConfig().ApplicationEnvironment)
					respData, _ := json.Marshal(commonDtos.StandardResponse{
						Status:  false,
						Message: "Internal Server Error",
					})
					if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
						utils.LogError("[%s] RAGChat MQTT: Failed to publish internal error response: %v", requestID, err)
					}
				}
				return
			}

			// Add tracking metadata to response
			res.RequestID = requestID
			res.Source = "MQTT_SUBSCRIBER"
			res.InstanceID = c.instanceID

			// Publish result back
			if mac != "" {
				respTopic := fmt.Sprintf("users/%s/%s/chat/answer", mac, utils.GetConfig().ApplicationEnvironment)
				// Use req.TerminalID if mac from topic is generic (security/consistency check)
				if mac != req.TerminalID && req.TerminalID != "" {
					utils.LogDebug("[%s] [Instance: %s] RAGChat MQTT: Response topic override check: TopicMAC=%s, PayloadID=%s", requestID, c.instanceID, mac, req.TerminalID)
				}

				resp := commonDtos.StandardResponse{
					Status:  true,
					Message: "Chat processed successfully",
					Data:    res,
				}
				respData, _ := json.Marshal(resp)
				utils.LogInfo("[%s] [Instance: %s] RAGChat MQTT [Handler: StartMqttSubscription]: Publishing answer to %s. Response: %s", requestID, c.instanceID, respTopic, res.Response)
				if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
					utils.LogError("[%s] RAGChat MQTT: Failed to publish chat response: %v", requestID, err)
				}
			}
		}(mac, req)
	})

	// Subscribe to general task signaling
	taskTopic := fmt.Sprintf("users/+/%s/task", config.ApplicationEnvironment)
	_ = c.mqttSvc.Subscribe(taskTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		utils.LogInfo("RAG Task Signaling MQTT: Received message on %s: %s", msg.Topic(), string(payload))
	})

	if err != nil {
		utils.LogError("RAGChat MQTT: Failed to subscribe to %s: %v", topic, err)
		return err
	}
	utils.LogInfo("RAGChat MQTT: Successfully subscribed to %s", topic)
	return nil
}

// Chat handles the AI Assistant chat/command classification.
// @Summary AI Assistant Chat
// @Description Classifies user prompt into Chat or Control and returns appropriate response or redirection.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGChatRequestDTO true "Chat Request"
// @Success 200 {object} commonDtos.StandardResponse
// @Failure 400 {object} commonDtos.StandardResponse
// @Failure 500 {object} commonDtos.StandardResponse "Internal Server Error"
// @Router /api/rag/chat [post]
func (c *RAGChatController) Chat(ctx *gin.Context) {
	var req dtos.RAGChatRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, commonDtos.StandardResponse{
			Status:  false,
			Message: "Validation Error",
			Details: []utils.ValidationErrorDetail{
				{Field: "payload", Message: "Invalid request body: " + err.Error()},
			},
		})
		return
	}

	uid, _ := ctx.Get("uid")
	uidStr := ""
	if uid != nil {
		uidStr = uid.(string)
	}

	// Apply deduplication to HTTP path as well
	c.mu.Lock()
	terminalKey := req.TerminalID
	last := c.lastPrompt[terminalKey]
	lastTime := c.lastPromptTime[terminalKey]
	now := time.Now()

	if last == req.Prompt && now.Sub(lastTime) < 3*time.Second {
		c.mu.Unlock()
		utils.LogInfo("RAGChat HTTP: Dropping duplicate prompt for %s (from HTTP): '%s'", terminalKey, req.Prompt)
		// Return previous success but don't re-process
		ctx.JSON(http.StatusOK, commonDtos.StandardResponse{
			Status:  true,
			Message: "Chat request received (duplicate dropped)",
		})
		return
	}

	// Update cache
	c.lastPrompt[terminalKey] = req.Prompt
	c.lastPromptTime[terminalKey] = now
	c.mu.Unlock()

	requestID := uuid.New().String()
	utils.LogInfo("[%s] RAGChat HTTP [Handler: Chat]: Starting chat process for UID: %s, Terminal: %s, Prompt: '%s'", requestID, uidStr, req.TerminalID, req.Prompt)

	res, err := c.chatUC.Chat(uidStr, req.TerminalID, req.Prompt, req.Language)
	if err != nil {
		utils.LogError("[%s] RAGChatController.Chat: %v", requestID, err)
		ctx.JSON(http.StatusInternalServerError, commonDtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	// Add tracking metadata
	res.Source = "HTTP_HANDLER"
	res.RequestID = requestID
	res.InstanceID = c.instanceID

	// For control commands, check status code
	// 400 (ambiguity) is a valid response requiring clarification, not an error
	// Only 401, 404, 500 are actual errors
	if res.IsControl && res.HTTPStatusCode != 0 && res.HTTPStatusCode != 200 && res.HTTPStatusCode != 400 {
		// Return the error status code from control execution
		ctx.JSON(res.HTTPStatusCode, commonDtos.StandardResponse{
			Status:  false,
			Message: "Command execution failed",
			Data:    res,
		})
		return
	}

	resp := commonDtos.StandardResponse{
		Status:  true,
		Message: "Chat processed successfully",
		Data:    res,
	}

	// Also publish to MQTT if service is available (for unified view on mobile apps)
	if c.mqttSvc != nil {
		// TerminalID in request for REST usually contains MAC or actual terminal ID
		// Using req.TerminalID as the identifier for MQTT response topic
		respTopic := fmt.Sprintf("users/%s/%s/chat/answer", req.TerminalID, utils.GetConfig().ApplicationEnvironment)
		respData, _ := json.Marshal(resp)
		utils.LogInfo("[%s] [Instance: %s] RAGChat HTTP [Handler: Chat]: Publishing answer to %s. Response: %s", requestID, c.instanceID, respTopic, res.Response)
		if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
			utils.LogError("[%s] RAGChatController.Chat: Failed to publish to MQTT: %v", requestID, err)
		}
	}

	ctx.JSON(http.StatusOK, resp)
}
