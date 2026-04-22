package recordings

import (
	"github.com/gin-gonic/gin"

	"sensio/domain/infrastructure"
	"sensio/domain/recordings/controllers"
	"sensio/domain/recordings/repositories"
	"sensio/domain/recordings/services"
	"sensio/domain/recordings/usecases"
)

type RecordingsModule struct {
	ListController                 *controllers.RecordingsListController
	GetByIDController              *controllers.RecordingsGetByIDController
	CreateController               *controllers.RecordingsCreateController
	DeleteController               *controllers.RecordingsDeleteController
	UploadIntentController         *controllers.UploadIntentController
	AudioUploadStatusController    *controllers.AudioUploadStatusController
	SaveRecordingUseCase           usecases.SaveRecordingUseCase
	GetAllUseCase                  usecases.GetAllRecordingsUseCase
	GetByIDUseCase                 usecases.GetRecordingByIDUseCase
	DeleteUseCase                  usecases.DeleteRecordingUseCase
	UploadIntentUseCase            usecases.UploadIntentUseCase
	UpdateAudioUploadStatusUseCase usecases.UpdateAudioUploadStatusUseCase
}

func NewRecordingsModule(badger *infrastructure.BadgerService, storageProvider infrastructure.StorageProvider) *RecordingsModule {
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

	// Upload Intent
	uploadIntentUseCase := usecases.NewUploadIntentUseCase(storageProvider, 900)
	uploadIntentController := controllers.NewUploadIntentController(uploadIntentUseCase)

	// Audio Upload Status
	audioUploadStatusRepo := repositories.NewAudioUploadStatusRepository()
	updateAudioUploadStatusUseCase := usecases.NewUpdateAudioUploadStatusUseCase(audioUploadStatusRepo)
	audioUploadStatusController := controllers.NewAudioUploadStatusController(updateAudioUploadStatusUseCase)

	return &RecordingsModule{
		ListController:                 listController,
		GetByIDController:              getByIDController,
		CreateController:               createController,
		DeleteController:               deleteController,
		UploadIntentController:         uploadIntentController,
		AudioUploadStatusController:    audioUploadStatusController,
		SaveRecordingUseCase:           saveUseCase,
		GetAllUseCase:                  getAllUseCase,
		GetByIDUseCase:                 getByIDUseCase,
		DeleteUseCase:                  deleteUseCase,
		UploadIntentUseCase:            uploadIntentUseCase,
		UpdateAudioUploadStatusUseCase: updateAudioUploadStatusUseCase,
	}
}

func (m *RecordingsModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	api := protected.Group("/api")
	{
		api.GET("/recordings", m.ListController.ListRecordings)
		api.GET("/recordings/:id", m.GetByIDController.GetRecordingByID)
		api.POST("/recordings", m.CreateController.CreateRecording)
		api.DELETE("/recordings/:id", m.DeleteController.DeleteRecording)
		api.POST("/recordings/upload/intent", m.UploadIntentController.CreateUploadIntent)
		api.POST("/recordings/upload/status", m.AudioUploadStatusController.UpdateStatus)
	}
}
