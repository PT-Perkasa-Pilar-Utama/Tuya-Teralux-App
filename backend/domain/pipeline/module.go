package pipeline

import (
	"sensio/domain/common/infrastructure"
	"sensio/domain/common/tasks"
	"sensio/domain/common/utils"
	pipelineControllers "sensio/domain/pipeline/controllers"
	pipelinedtos "sensio/domain/pipeline/dtos"
	pipelineRoutes "sensio/domain/pipeline/routes"
	pipelineUsecases "sensio/domain/pipeline/usecases"
	ragUsecases "sensio/domain/rag/usecases"
	recordingUsecases "sensio/domain/recordings/usecases"
	speechUsecases "sensio/domain/speech/usecases"

	"github.com/gin-gonic/gin"
)

func InitModule(
	protected *gin.RouterGroup,
	cfg *utils.Config,
	badger *infrastructure.BadgerService,
	transcribeUC speechUsecases.TranscribeUseCase,
	translateUC ragUsecases.TranslateUseCase,
	summaryUC ragUsecases.SummaryUseCase,
	saveRecordingUC recordingUsecases.SaveRecordingUseCase,
	uploadSessionUC speechUsecases.UploadSessionUseCase,
) {
	store := tasks.NewStatusStore[pipelinedtos.PipelineStatusDTO]()
	cache := tasks.NewBadgerTaskCacheFromService(badger, "cache:pipeline:task:")

	pipelineUC := pipelineUsecases.NewPipelineUseCase(transcribeUC, translateUC, summaryUC, cache, store)
	statusUC := tasks.NewGenericStatusUseCase(cache, store)
	pipelineCtrl := pipelineControllers.NewPipelineController(pipelineUC, statusUC, saveRecordingUC, uploadSessionUC, cfg)

	pipelineRoutes.SetupPipelineRoutes(protected, pipelineCtrl)
}
