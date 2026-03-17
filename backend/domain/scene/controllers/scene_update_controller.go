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

type SceneUpdateController struct {
	useCase *usecases.UpdateSceneUseCase
}

func NewSceneUpdateController(useCase *usecases.UpdateSceneUseCase) *SceneUpdateController {
	return &SceneUpdateController{
		useCase: useCase,
	}
}

// UpdateScene handles PUT /api/terminal/:id/scenes/:scene_id
// @Summary Update an existing scene
// @Description Update the configuration of a specific scene
// @Tags 03. Scenes
// @Accept json
// @Produce json
// @Param id path string true "Terminal UUID"
// @Param scene_id path string true "Scene UUID"
// @Param scene body scene_dtos.UpdateSceneRequestDTO true "Updated scene configuration"
// @Success 200 {object} dtos.StandardResponse{data=scene_dtos.SceneIDResponseDTO}
// @Failure      400  {object}  dtos.ValidationErrorResponse
// @Failure      404  {object}  dtos.ErrorResponse
// @Failure      500  {object}  dtos.ErrorResponse
// @Security BearerAuth
// @Router /api/terminal/{id}/scenes/{scene_id} [put]
func (c *SceneUpdateController) UpdateScene(ctx *gin.Context) {
	terminalID := ctx.Param("id")
	id := ctx.Param("scene_id")
	var req scene_dtos.UpdateSceneRequestDTO
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

	var actions entities.Actions
	if len(req.Actions) > 0 {
		actions = make(entities.Actions, len(req.Actions))
		for i, a := range req.Actions {
			actions[i] = entities.Action(a)
		}
	}

	if err := c.useCase.UpdateScene(terminalID, id, req.Name, actions); err != nil {
		utils.LogError("SceneUpdateController.UpdateScene: %v", err)
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
		Message: "Scene updated successfully",
		Data:    scene_dtos.SceneIDResponseDTO{SceneID: id},
	})
}
