package recordings

import (
	"github.com/gin-gonic/gin"

	"teralux_app/domain/common/infrastructure"
	"teralux_app/domain/recordings/controllers"
	"teralux_app/domain/recordings/repositories"
	"teralux_app/domain/recordings/usecases"
)

type RecordingsModule struct {
	Controller           *controllers.RecordingsController
	SaveRecordingUseCase usecases.SaveRecordingUseCase
	GetAllUseCase        usecases.GetAllRecordingsUseCase
	GetByIDUseCase       usecases.GetRecordingByIDUseCase
	DeleteUseCase        usecases.DeleteRecordingUseCase
}

func NewRecordingsModule(badger *infrastructure.BadgerService) *RecordingsModule {
	repo := repositories.NewRecordingRepository(badger)
	
	// Inject DefaultFileService
	fileService := infrastructure.DefaultFileService
	saveUseCase := usecases.NewSaveRecordingUseCase(repo, fileService)
	
	getAllUseCase := usecases.NewGetAllRecordingsUseCase(repo)
	getByIDUseCase := usecases.NewGetRecordingByIDUseCase(repo)
	deleteUseCase := usecases.NewDeleteRecordingUseCase(repo)
	controller := controllers.NewRecordingsController(getAllUseCase, getByIDUseCase, deleteUseCase, saveUseCase)

	return &RecordingsModule{
		Controller:           controller,
		SaveRecordingUseCase: saveUseCase,
		GetAllUseCase:        getAllUseCase,
		GetByIDUseCase:       getByIDUseCase,
		DeleteUseCase:        deleteUseCase,
	}
}

func (m *RecordingsModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	api := router.Group("/api")
	{
		api.GET("/recordings", m.Controller.GetAllRecordings)
		api.GET("/recordings/:id", m.Controller.GetRecordingByID)
		api.POST("/recordings", m.Controller.UploadRecording)
		api.DELETE("/recordings/:id", m.Controller.DeleteRecording)
	}
}
