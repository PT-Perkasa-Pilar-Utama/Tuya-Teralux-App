package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	scene_dtos "sensio/domain/scene/dtos"
	"sensio/domain/scene/entities"
	"sensio/domain/scene/usecases"

	"github.com/gin-gonic/gin"
)

type SceneListController struct {
	useCase *usecases.GetAllScenesUseCase
}

func NewSceneListController(useCase *usecases.GetAllScenesUseCase) *SceneListController {
	return &SceneListController{
		useCase: useCase,
	}
}

// ListScenes handles GET /api/terminal/:id/scenes
// @Summary List all scenes for a Terminal device
// @Description Retrieve a list of all configured scenes for a specific Terminal device, including all actions for each scene.
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Terminal UUID"
// @Success 200 {object} dtos.StandardResponse{data=[]scene_dtos.SceneResponseDTO} "List of scenes with actions"
// @Security BearerAuth
// @Router /api/terminal/{id}/scenes [get]
func (c *SceneListController) ListScenes(ctx *gin.Context) {
	terminalID := ctx.Param("id")
	scenes, err := c.useCase.ListScenes(terminalID)
	if err != nil {
		utils.LogError("SceneListController.ListScenes: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	response := make([]scene_dtos.SceneResponseDTO, len(scenes))
	for i, s := range scenes {
		response[i] = scene_dtos.SceneResponseDTO{
			ID:         s.ID,
			TerminalID: s.TerminalID,
			Name:       s.Name,
			Actions:    toActionDTOs(s.Actions),
		}
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scenes retrieved successfully",
		Data:    response,
	})
}

func toActionDTOs(actions entities.Actions) []scene_dtos.ActionDTO {
	result := make([]scene_dtos.ActionDTO, len(actions))
	for i, a := range actions {
		result[i] = scene_dtos.ActionDTO{
			DeviceID: a.DeviceID,
			Code:     a.Code,
			RemoteID: a.RemoteID,
			Topic:    a.Topic,
			Value:    a.Value,
		}
	}
	return result
}
