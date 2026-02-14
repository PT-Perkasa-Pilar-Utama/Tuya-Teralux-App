package scene

import (
	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/scene/controllers"
	"teralux_app/domain/scene/repositories"
	"teralux_app/domain/scene/usecases"
	tuyaUsecases "teralux_app/domain/tuya/usecases"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SceneModule struct {
	AddController     *controllers.SceneAddController
	ListController    *controllers.SceneListController
	UpdateController  *controllers.SceneUpdateController
	DeleteController  *controllers.SceneDeleteController
	ControlController *controllers.SceneControlController
}

func NewSceneModule(db *gorm.DB, tuyaCmd tuyaUsecases.TuyaDeviceControlExecutor, mqttSvc *infrastructure.MqttService) *SceneModule {
	repo := repositories.NewSceneRepository(db)

	addUC := usecases.NewAddSceneUseCase(repo)
	updateUC := usecases.NewUpdateSceneUseCase(repo)
	deleteUC := usecases.NewDeleteSceneUseCase(repo)
	getAllUC := usecases.NewGetAllScenesUseCase(repo)
	controlUC := usecases.NewControlSceneUseCase(repo, tuyaCmd, mqttSvc)

	return &SceneModule{
		AddController:     controllers.NewSceneAddController(addUC),
		ListController:    controllers.NewSceneListController(getAllUC),
		UpdateController:  controllers.NewSceneUpdateController(updateUC),
		DeleteController:  controllers.NewSceneDeleteController(deleteUC),
		ControlController: controllers.NewSceneControlController(controlUC),
	}
}

func (m *SceneModule) RegisterRoutes(protected *gin.RouterGroup) {
	group := protected.Group("/api/teralux/:id/scenes")
	{
		group.POST("", m.AddController.AddScene)
		group.GET("", m.ListController.ListScenes)
		group.PUT("/:scene_id", m.UpdateController.UpdateScene)
		group.DELETE("/:scene_id", m.DeleteController.DeleteScene)
		group.GET("/:scene_id/control", m.ControlController.ControlScene)
	}
}
