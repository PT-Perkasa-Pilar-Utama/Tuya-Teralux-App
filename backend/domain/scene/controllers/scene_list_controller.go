package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	scene_dtos "teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/usecases"

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

// ListScenes handles GET /api/teralux/:id/scenes
// @Summary List all scenes for a Teralux device
// @Description Retrieve a list of all configured scenes for a specific Teralux device
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Teralux UUID"
// @Success 200 {object} dtos.StandardResponse{data=[]scene_dtos.SceneListResponseDTO} "List of scenes"
// @Security BearerAuth
// @Router /api/teralux/{id}/scenes [get]
func (c *SceneListController) ListScenes(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	scenes, err := c.useCase.ListScenes(teraluxID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	response := make([]scene_dtos.SceneListResponseDTO, len(scenes))
	for i, s := range scenes {
		response[i] = scene_dtos.SceneListResponseDTO{
			ID:        s.ID,
			TeraluxID: s.TeraluxID,
			Name:      s.Name,
		}
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scenes retrieved successfully",
		Data:    response,
	})
}
