package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	scene_dtos "teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/usecases"

	"github.com/gin-gonic/gin"
)

type SceneAddController struct {
	useCase *usecases.AddSceneUseCase
}

func NewSceneAddController(useCase *usecases.AddSceneUseCase) *SceneAddController {
	return &SceneAddController{
		useCase: useCase,
	}
}

// AddScene handles POST /api/teralux/:id/scenes
// @Summary Create a new scene
// @Description Create a scene with a name and a list of actions (Tuya/MQTT)
// @Tags 03. Scenes
// @Accept json
// @Produce json
// @Param id path string true "Teralux UUID"
// @Param scene body scene_dtos.CreateSceneRequestDTO true "Scene configuration"
// @Success 201 {object} dtos.StandardResponse "Scene created"
// @Failure 400 {object} dtos.StandardResponse "Invalid request"
// @Failure 500 {object} dtos.StandardResponse "Internal error"
// @Security BearerAuth
// @Router /api/teralux/{id}/scenes [post]
func (c *SceneAddController) AddScene(ctx *gin.Context) {
	teraluxID := ctx.Param("id")
	var req scene_dtos.CreateSceneRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	actions := make(entities.Actions, len(req.Actions))
	for i, a := range req.Actions {
		actions[i] = entities.Action(a)
	}

	sceneID, err := c.useCase.AddScene(teraluxID, req.Name, actions)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Scene created successfully",
		Data:    gin.H{"scene_id": sceneID},
	})
}
