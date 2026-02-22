package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"
	scene_dtos "teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/usecases"

	"github.com/gin-gonic/gin"
)

type SceneListAllController struct {
	useCase *usecases.GetAllGroupedScenesUseCase
}

func NewSceneListAllController(useCase *usecases.GetAllGroupedScenesUseCase) *SceneListAllController {
	return &SceneListAllController{useCase: useCase}
}

// ListAllScenes handles GET /api/scenes
// @Summary List all scenes across all Teralux devices
// @Description Retrieve all configured scenes grouped under each Teralux device
// @Tags 03. Scenes
// @Produce json
// @Success 200 {object} dtos.StandardResponse{data=[]scene_dtos.TeraluxScenesWrapperDTO} "All scenes grouped by Teralux"
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

	response := make([]scene_dtos.TeraluxScenesWrapperDTO, 0, len(grouped))
	for teraluxID, scenes := range grouped {
		response = append(response, scene_dtos.TeraluxScenesWrapperDTO{
			Teralux: scene_dtos.TeraluxScenesDTO{
				TeraluxID: teraluxID,
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

func toSceneResponseDTOs(scenes []entities.Scene) []scene_dtos.SceneResponseDTO {
	result := make([]scene_dtos.SceneResponseDTO, len(scenes))
	for i, s := range scenes {
		result[i] = scene_dtos.SceneResponseDTO{
			ID:        s.ID,
			TeraluxID: s.TeraluxID,
			Name:      s.Name,
			Actions:   toActionDTOs(s.Actions),
		}
	}
	return result
}
