package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/scene/usecases"

	"github.com/gin-gonic/gin"
)

type SceneControlController struct {
	useCase *usecases.ControlSceneUseCase
}

func NewSceneControlController(useCase *usecases.ControlSceneUseCase) *SceneControlController {
	return &SceneControlController{
		useCase: useCase,
	}
}

// ControlScene handles GET /api/teralux/:id/scenes/:scene_id/control
// @Summary Apply/Trigger a scene
// @Description Trigger all actions defined in a specific scene
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Teralux UUID"
// @Param scene_id path string true "Scene UUID"
// @Success 200 {object} dtos.StandardResponse "Scene applied"
// @Security BearerAuth
// @Router /api/teralux/{id}/scenes/{scene_id}/control [get]
func (c *SceneControlController) ControlScene(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	id := ctx.Param("scene_id")
	accessToken := ctx.GetString("access_token")

	if err := c.useCase.ControlScene(teraluxID, id, accessToken); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "record not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene applied successfully",
	})
}
