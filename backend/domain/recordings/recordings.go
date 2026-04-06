package recordings

import (
	"github.com/gin-gonic/gin"

	"sensio/domain/common/infrastructure"
	"sensio/domain/recordings/controllers"
	"sensio/domain/recordings/repositories"
	"sensio/domain/recordings/services"
	"sensio/domain/recordings/usecases"
)

type RecordingsModule struct {
	ListController       *controllers.RecordingsListController
	GetByIDController    *controllers.RecordingsGetByIDController
	CreateController     *controllers.RecordingsCreateController
	DeleteController     *controllers.RecordingsDeleteController
	SaveRecordingUseCase usecases.SaveRecordingUseCase
	GetAllUseCase        usecases.GetAllRecordingsUseCase
	GetByIDUseCase       usecases.GetRecordingByIDUseCase
	DeleteUseCase        usecases.DeleteRecordingUseCase
}

func NewRecordingsModule(badger *infrastructure.BadgerService) *RecordingsModule {
	repo := repositories.NewRecordingRepository(badger)

	// Inject DefaultFileService
	fileService := infrastructure.DefaultFileService
	bigAudioService := services.NewBIGRoomAudioUpdateService()
	saveUseCase := usecases.NewSaveRecordingUseCase(repo, fileService, bigAudioService)

	getAllUseCase := usecases.NewGetAllRecordingsUseCase(repo)
	getByIDUseCase := usecases.NewGetRecordingByIDUseCase(repo)
	deleteUseCase := usecases.NewDeleteRecordingUseCase(repo)
	listController := controllers.NewRecordingsListController(getAllUseCase)
	getByIDController := controllers.NewRecordingsGetByIDController(getByIDUseCase)
	createController := controllers.NewRecordingsCreateController(saveUseCase)
	deleteController := controllers.NewRecordingsDeleteController(deleteUseCase)

	return &RecordingsModule{
		ListController:       listController,
		GetByIDController:    getByIDController,
		CreateController:     createController,
		DeleteController:     deleteController,
		SaveRecordingUseCase: saveUseCase,
		GetAllUseCase:        getAllUseCase,
		GetByIDUseCase:       getByIDUseCase,
		DeleteUseCase:        deleteUseCase,
	}
}

func (m *RecordingsModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	api := protected.Group("/api")
	{
		api.GET("/recordings", m.ListController.ListRecordings)
		api.GET("/recordings/:id", m.GetByIDController.GetRecordingByID)
		api.POST("/recordings", m.CreateController.CreateRecording)
		api.DELETE("/recordings/:id", m.DeleteController.DeleteRecording)
	}
}
