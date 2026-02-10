package scene

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/scene/controllers"
	"teralux_app/domain/scene/repositories"
	"teralux_app/domain/scene/usecases"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SceneModule struct {
	Controller *controllers.SceneController
}

func NewSceneModule(db *gorm.DB, tuyaCmd controllers.TuyaDeviceControlExecutor, mqttSvc *infrastructure.MqttService) *SceneModule {
	repo := repositories.NewSceneRepository(db)

	addUC := usecases.NewAddSceneUseCase(repo)
	updateUC := usecases.NewUpdateSceneUseCase(repo)
	deleteUC := usecases.NewDeleteSceneUseCase(repo)
	getAllUC := usecases.NewGetAllScenesUseCase(repo)
	controlUC := usecases.NewControlSceneUseCase(repo, tuyaCmd, mqttSvc)

	controller := controllers.NewSceneController(addUC, updateUC, deleteUC, getAllUC, controlUC)

	return &SceneModule{
		Controller: controller,
	}
}

func (m *SceneModule) RegisterRoutes(protected *gin.RouterGroup) {
	group := protected.Group("/api/teralux/:id/scenes")
	{
		group.POST("", m.Controller.AddScene)
		group.GET("", m.Controller.GetAllScenes)
		group.PUT("/:scene_id", m.Controller.UpdateScene)
		group.DELETE("/:scene_id", m.Controller.DeleteScene)
		group.GET("/:scene_id/control", m.Controller.ControlScene)
	}
}
