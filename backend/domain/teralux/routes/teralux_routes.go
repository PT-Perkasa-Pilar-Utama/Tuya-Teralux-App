package routes

import (
	teralux_controllers "teralux_app/domain/teralux/controllers/teralux"

	"github.com/gin-gonic/gin"
)

// SetupTeraluxRoutes registers endpoints for teralux CRUD operations
func SetupTeraluxRoutes(
	router gin.IRouter,
	createController *teralux_controllers.CreateTeraluxController,
	getAllController *teralux_controllers.GetAllTeraluxController,
	getByIDController *teralux_controllers.GetTeraluxByIDController,
	updateController *teralux_controllers.UpdateTeraluxController,
	deleteController *teralux_controllers.DeleteTeraluxController,
) {
	api := router.Group("/api/teralux")
	{
		// POST /api/teralux - Create a new teralux
		api.POST("", createController.CreateTeralux)

		// GET /api/teralux - Get all teralux
		api.GET("", getAllController.GetAllTeralux)

		// GET /api/teralux/:id - Get teralux by ID
		api.GET("/:id", getByIDController.GetTeraluxByID)

		// PUT /api/teralux/:id - Update teralux
		api.PUT("/:id", updateController.UpdateTeralux)

		// DELETE /api/teralux/:id - Delete teralux (soft delete)
		api.DELETE("/:id", deleteController.DeleteTeralux)
	}
}
