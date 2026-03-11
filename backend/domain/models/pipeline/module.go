package pipeline

import (
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineControllers "sensio/domain/models/pipeline/controllers"
	pipelinedtos "sensio/domain/models/pipeline/dtos"
	pipelineRoutes "sensio/domain/models/pipeline/routes"
	pipelineUsecases "sensio/domain/models/pipeline/usecases"
	ragUsecases "sensio/domain/models/rag/usecases"
	recordingUsecases "sensio/domain/recordings/usecases"
	whisperUsecases "sensio/domain/models/whisper/usecases"

	"github.com/gin-gonic/gin"
)

func InitModule(
	protected *gin.RouterGroup,
	cfg *utils.Config,
	badger *infrastructure.BadgerService,
	transcribeUC whisperUsecases.TranscribeUseCase,
	translateUC ragUsecases.TranslateUseCase,
	summaryUC ragUsecases.SummaryUseCase,
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
	uploadSessionUC whisperUsecases.UploadSessionUseCase,
	mqttSvc *infrastructure.MqttService,
) {
	store := tasks.NewStatusStore[pipelinedtos.PipelineStatusDTO]()
	cache := tasks.NewBadgerTaskCacheFromService(badger, "cache:pipeline:task:")

	pipelineUC := pipelineUsecases.NewPipelineUseCase(transcribeUC, translateUC, summaryUC, cache, store, mqttSvc)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)
	pipelineCtrl := pipelineControllers.NewPipelineController(pipelineUC, statusUC, saveRecordingUC, uploadSessionUC, cfg)

	pipelineRoutes.SetupPipelineRoutes(protected, pipelineCtrl)
}
