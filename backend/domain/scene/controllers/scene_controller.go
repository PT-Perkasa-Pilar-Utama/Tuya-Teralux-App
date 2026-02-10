package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	scene_dtos "teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/usecases"

	"github.com/gin-gonic/gin"
)

type SceneController struct {
	addUC     *usecases.AddSceneUseCase
	updateUC  *usecases.UpdateSceneUseCase
	deleteUC  *usecases.DeleteSceneUseCase
	getAllUC  *usecases.GetAllScenesUseCase
	controlUC *usecases.ControlSceneUseCase
}

func NewSceneController(
	addUC *usecases.AddSceneUseCase,
	updateUC *usecases.UpdateSceneUseCase,
	deleteUC *usecases.DeleteSceneUseCase,
	getAllUC *usecases.GetAllScenesUseCase,
	controlUC *usecases.ControlSceneUseCase,
) *SceneController {
	return &SceneController{
		addUC:     addUC,
		updateUC:  updateUC,
		deleteUC:  deleteUC,
		getAllUC:  getAllUC,
		controlUC: controlUC,
	}
}

// AddScene handles POST /api/scenes
// @Summary Create a new scene
// @Description Create a scene with a name and a list of actions (Tuya/MQTT)
// @Tags scenes
// @Accept json
// @Produce json
// @Param scene body scene_dtos.CreateSceneRequestDTO true "Scene configuration"
// @Success 201 {object} dtos.StandardResponse "Scene created"
// @Failure 400 {object} dtos.StandardResponse "Invalid request"
// @Failure 500 {object} dtos.StandardResponse "Internal error"
// @Router /api/scenes [post]
func (c *SceneController) AddScene(ctx *gin.Context) {
	var req scene_dtos.CreateSceneRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	id, err := c.addUC.Execute(&req)
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
		Data:    gin.H{"scene_id": id},
	})
}

// UpdateScene handles PUT /api/scenes/:id
// @Summary Update an existing scene
// @Description Update the configuration of a specific scene
// @Tags scenes
// @Accept json
// @Produce json
// @Param id path string true "Scene UUID"
// @Param scene body scene_dtos.UpdateSceneRequestDTO true "Updated scene configuration"
// @Success 200 {object} dtos.StandardResponse "Scene updated"
// @Failure 404 {object} dtos.StandardResponse "Scene not found"
// @Router /api/scenes/{id} [put]
func (c *SceneController) UpdateScene(ctx *gin.Context) {
	id := ctx.Param("id")
	var req scene_dtos.UpdateSceneRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	if err := c.updateUC.Execute(id, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene updated successfully",
		Data:    gin.H{"scene_id": id},
	})
}

// DeleteScene handles DELETE /api/scenes/:id
// @Summary Delete a scene
// @Description Remove a specific scene configuration
// @Tags scenes
// @Produce json
// @Param id path string true "Scene UUID"
// @Success 200 {object} dtos.StandardResponse "Scene deleted"
// @Router /api/scenes/{id} [delete]
func (c *SceneController) DeleteScene(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.deleteUC.Execute(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scene deleted successfully",
	})
}

// GetAllScenes handles GET /api/scenes
// @Summary List all scenes
// @Description Retrieve a list of all configured scenes
// @Tags scenes
// @Produce json
// @Success 200 {object} dtos.StandardResponse{data=[]scene_dtos.SceneListResponseDTO} "List of scenes"
// @Router /api/scenes [get]
func (c *SceneController) GetAllScenes(ctx *gin.Context) {
	scenes, err := c.getAllUC.Execute()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Scenes retrieved successfully",
		Data:    scenes,
	})
}

// ControlScene handles GET /api/scenes/:id/control
// @Summary Apply/Trigger a scene
// @Description Trigger all actions defined in a specific scene
// @Tags scenes
// @Produce json
// @Param id path string true "Scene UUID"
// @Success 200 {object} dtos.StandardResponse "Scene applied"
// @Router /api/scenes/{id}/control [get]
func (c *SceneController) ControlScene(ctx *gin.Context) {
	id := ctx.Param("id")
	accessToken := ctx.GetString("access_token") // Typically set by AuthMiddleware

	if err := c.controlUC.Execute(id, accessToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
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
