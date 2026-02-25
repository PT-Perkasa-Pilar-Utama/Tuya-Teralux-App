package controllers

import (
	"net/http"
	"sensio/domain/common/dtos"
	"sensio/domain/common/utils"
	"sensio/domain/scene/usecases"

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

// ControlScene handles GET /api/terminal/:id/scenes/:scene_id/control
// @Summary Apply/Trigger a scene
// @Description Trigger all actions defined in a specific scene
// @Tags 03. Scenes
// @Produce json
// @Param id path string true "Terminal UUID"
// @Param scene_id path string true "Scene UUID"
// @Success 200 {object} dtos.StandardResponse "Scene applied"
// @Security BearerAuth
// @Router /api/terminal/{id}/scenes/{scene_id}/control [get]
func (c *SceneControlController) ControlScene(ctx *gin.Context) {
	terminalID := ctx.Param("id")
	id := ctx.Param("scene_id")
	accessToken := ctx.GetString("access_token")

	if err := c.useCase.ControlScene(terminalID, id, accessToken); err != nil {
		utils.LogError("SceneControlController.ControlScene: %v", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "record not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, dtos.StandardResponse{
			Status:  false,
			Message: http.StatusText(statusCode),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene applied successfully",
	})
}
