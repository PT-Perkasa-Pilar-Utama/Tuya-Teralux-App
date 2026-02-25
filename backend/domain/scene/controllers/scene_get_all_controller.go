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

type SceneListAllController struct {
	useCase *usecases.GetAllGroupedScenesUseCase
}

func NewSceneListAllController(useCase *usecases.GetAllGroupedScenesUseCase) *SceneListAllController {
	return &SceneListAllController{useCase: useCase}
}

// ListAllScenes handles GET /api/scenes
// @Summary List all scenes across all Terminal devices
// @Description Retrieve all configured scenes grouped under each Terminal device
// @Tags 03. Scenes
// @Produce json
// @Success 200 {object} dtos.StandardResponse{data=[]scene_dtos.TerminalScenesWrapperDTO} "All scenes grouped by Terminal"
// @Security BearerAuth
// @Router /api/scenes [get]
func (c *SceneListAllController) ListAllScenes(ctx *gin.Context) {
	grouped, err := c.useCase.ListAllGrouped()
	if err != nil {
		utils.LogError("SceneListAllController.ListAllScenes: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	response := make([]scene_dtos.TerminalScenesWrapperDTO, 0, len(grouped))
	for terminalID, scenes := range grouped {
		response = append(response, scene_dtos.TerminalScenesWrapperDTO{
			Terminal: scene_dtos.TerminalScenesDTO{
				TerminalID: terminalID,
				Scenes:    toSceneItemDTOs(scenes),
			},
		})
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scenes retrieved successfully",
		Data:    response,
	})
}

func toSceneItemDTOs(scenes []entities.Scene) []scene_dtos.SceneItemDTO {
	result := make([]scene_dtos.SceneItemDTO, len(scenes))
	for i, s := range scenes {
		result[i] = scene_dtos.SceneItemDTO{
			ID:      s.ID,
			Name:    s.Name,
			Actions: toActionDTOs(s.Actions),
		}
	}
	return result
}
