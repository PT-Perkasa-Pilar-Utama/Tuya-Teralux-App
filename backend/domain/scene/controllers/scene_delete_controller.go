package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/scene/usecases"

	"github.com/gin-gonic/gin"
)

type SceneDeleteController struct {
	useCase *usecases.DeleteSceneUseCase
}

func NewSceneDeleteController(useCase *usecases.DeleteSceneUseCase) *SceneDeleteController {
	return &SceneDeleteController{
		useCase: useCase,
	}
}

// DeleteScene handles DELETE /api/terminal/:id/scenes/:scene_id
// @Summary Delete a scene
// @Description Remove a specific scene configuration
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Terminal UUID"
// @Param scene_id path string true "Scene UUID"
// @Success 200 {object} dtos.StandardResponse "Scene deleted"
// @Security BearerAuth
// @Router /api/terminal/{id}/scenes/{scene_id} [delete]
func (c *SceneDeleteController) DeleteScene(ctx *gin.Context) {
	terminalID := ctx.Param("id")
	id := ctx.Param("scene_id")
	if err := c.useCase.DeleteScene(terminalID, id); err != nil {
		utils.LogError("SceneDeleteController.DeleteScene: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene deleted successfully",
	})
}
