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

type SceneAddController struct {
	useCase *usecases.AddSceneUseCase
}

func NewSceneAddController(useCase *usecases.AddSceneUseCase) *SceneAddController {
	return &SceneAddController{
		useCase: useCase,
	}
}

// AddScene handles POST /api/terminal/:id/scenes
// @Summary Create a new scene
// @Description Create a scene with a name and a list of actions (Tuya/MQTT)
// @Tags 03. Scenes
// @Accept json
// @Produce json
// @Param id path string true "Terminal UUID"
// @Param scene body scene_dtos.CreateSceneRequestDTO true "Scene configuration"
// @Success 201 {object} dtos.StandardResponse{data=scene_dtos.SceneIDResponseDTO}
// @Failure      400  {object}  dtos.ValidationErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Security BearerAuth
// @Router /api/terminal/{id}/scenes [post]
func (c *SceneAddController) AddScene(ctx *gin.Context) {
	terminalID := ctx.Param("id")
	var req scene_dtos.CreateSceneRequestDTO
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

	actions := make(entities.Actions, len(req.Actions))
	for i, a := range req.Actions {
		actions[i] = entities.Action(a)
	}

	sceneID, err := c.useCase.AddScene(terminalID, req.Name, actions)
	if err != nil {
		utils.LogError("SceneAddController.AddScene: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.StandardResponse{
		Status:  true,
		Message: "Scene created successfully",
		Data:    scene_dtos.SceneIDResponseDTO{SceneID: sceneID},
	})
}
