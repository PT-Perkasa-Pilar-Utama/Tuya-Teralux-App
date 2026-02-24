package controllers

import (
	"encoding/json"
	"net/http"
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/common/utils"
	"teralux_app/domain/rag/dtos"
	"teralux_app/domain/rag/usecases"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

type RAGChatController struct {
	chatUC  usecases.ChatUseCase
	mqttSvc *infrastructure.MqttService
}

func NewRAGChatController(chatUC usecases.ChatUseCase, mqttSvc *infrastructure.MqttService) *RAGChatController {
	return &RAGChatController{
		chatUC:  chatUC,
		mqttSvc: mqttSvc,
	}
}

func (c *RAGChatController) StartMqttSubscription() {
	if c.mqttSvc == nil {
		return
	}

	topic := "users/teralux/chat"
	err := c.mqttSvc.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		payload := msg.Payload()
		utils.LogInfo("RAGChat MQTT: Received message on %s, payload size: %d", topic, len(payload))
		if len(payload) == 0 {
			return
		}

		var req dtos.RAGChatRequestDTO
		// Strictly JSON unmarshalling as per user request
		err := json.Unmarshal(payload, &req)
		if err != nil {
			utils.LogError("RAGChat MQTT: Failed to unmarshal message: %v", err)
			respTopic := "users/teralux/chat/answer"
			respData, _ := json.Marshal(dtos.StandardResponse{
				Status:  false,
				Message: "Validation Error",
				Details: []utils.ValidationErrorDetail{
					{Field: "payload", Message: "Invalid JSON payload: " + err.Error()},
				},
			})
			if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
				utils.LogError("RAGChat MQTT: Failed to publish error response: %v", err)
			}
			return
		}

		if req.Prompt == "" || req.TeraluxID == "" {
			utils.LogError("RAGChat MQTT: Missing prompt or teralux_id")
			respTopic := "users/teralux/chat/answer"
			respData, _ := json.Marshal(dtos.StandardResponse{
				Status:  false,
				Message: "Validation Error",
				Details: []utils.ValidationErrorDetail{
					{Field: "prompt", Message: "prompt is required"},
					{Field: "teralux_id", Message: "teralux_id is required"},
				},
			})
			if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
				utils.LogError("RAGChat MQTT: Failed to publish validation error response: %v", err)
			}
			return
		}

		// Process chat
		uid := req.UID
		if uid == "" {
			uid = utils.GetConfig().TuyaUserID
		}

		utils.LogInfo("RAGChat MQTT: Starting chat process for UID: %s, Prompt: '%s'", uid, req.Prompt)
		res, err := c.chatUC.Chat(uid, req.TeraluxID, req.Prompt, req.Language)
		if err != nil {
			utils.LogError("RAGChat MQTT: Chat processing failed: %v", err)
			respTopic := "users/teralux/chat/answer"
			respData, _ := json.Marshal(dtos.StandardResponse{
				Status:  false,
				Message: "Internal Server Error",
			})
			if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
				utils.LogError("RAGChat MQTT: Failed to publish internal error response: %v", err)
			}
			return
		}

		// Publish result back
		respTopic := "users/teralux/chat/answer"
		resp := dtos.StandardResponse{
			Status:  true,
			Message: "Chat processed successfully",
			Data:    res,
		}
		respData, _ := json.Marshal(resp)
		utils.LogInfo("RAGChat MQTT: Publishing answer to %s. Response: %s", respTopic, res.Response)
		if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
			utils.LogError("RAGChat MQTT: Failed to publish chat response: %v", err)
		}
	})

	if err != nil {
		utils.LogError("RAGChat MQTT: Failed to subscribe to %s: %v", topic, err)
	}
}

// Chat handles the AI Assistant chat/command classification.
// @Summary AI Assistant Chat
// @Description Classifies user prompt into Chat or Control and returns appropriate response or redirection.
// @Tags 05. RAG
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dtos.RAGChatRequestDTO true "Chat Request"
// @Success 200 {object} dtos.StandardResponse
// @Failure 400 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse "Internal Server Error"
// @Router /api/rag/chat [post]
func (c *RAGChatController) Chat(ctx *gin.Context) {
	var req dtos.RAGChatRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
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

	res, err := c.chatUC.Chat(uidStr, req.TeraluxID, req.Prompt, req.Language)
	if err != nil {
		utils.LogError("RAGChatController.Chat: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	// For control commands, check status code
	// 400 (ambiguity) is a valid response requiring clarification, not an error
	// Only 401, 404, 500 are actual errors
	if res.IsControl && res.HTTPStatusCode != 0 && res.HTTPStatusCode != 200 && res.HTTPStatusCode != 400 {
		// Return the error status code from control execution
		ctx.JSON(res.HTTPStatusCode, dtos.StandardResponse{
			Status:  false,
			Message: "Command execution failed",
			Data:    res,
		})
		return
	}

	resp := dtos.StandardResponse{
		Status:  true,
		Message: "Chat processed successfully",
		Data:    res,
	}

	// Also publish to MQTT if service is available (for unified view on mobile apps)
	if c.mqttSvc != nil {
		respTopic := "users/teralux/chat/answer"
		respData, _ := json.Marshal(resp)
		if err := c.mqttSvc.Publish(respTopic, 0, false, respData); err != nil {
			utils.LogError("RAGChatController.Chat: Failed to publish to MQTT: %v", err)
		}
	}

	ctx.JSON(http.StatusOK, resp)
}
