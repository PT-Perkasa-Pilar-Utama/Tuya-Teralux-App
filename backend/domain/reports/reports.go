package reports

import (
	"github.com/gin-gonic/gin"

	"sensio/domain/common/infrastructure"
	"sensio/domain/reports/controllers"
	"sensio/domain/reports/repositories"
	"sensio/domain/reports/usecases"
)

type ReportsModule struct {
	GetByIDController     *controllers.ReportsGetByIDController
	DeleteController      *controllers.ReportsDeleteController
	DownloadController    *controllers.ReportDownloadController
	SaveReportUseCase     usecases.SaveReportUseCase
	GetByIDUseCase        usecases.GetReportByIDUseCase
	DeleteUseCase         usecases.DeleteReportUseCase
}

func NewReportsModule(badger *infrastructure.BadgerService, s3Service *infrastructure.S3Service) *ReportsModule {
	repo := repositories.NewReportRepository(badger)
	fileService := infrastructure.DefaultFileService

	saveUseCase := usecases.NewSaveReportUseCase(repo, fileService, s3Service)
	getByIDUseCase := usecases.NewGetReportByIDUseCase(repo, s3Service)
	deleteUseCase := usecases.NewDeleteReportUseCase(repo, s3Service)

	getByIDController := controllers.NewReportsGetByIDController(getByIDUseCase)
	deleteController := controllers.NewReportsDeleteController(deleteUseCase)
	downloadController := controllers.NewReportDownloadController(repo, s3Service)

	return &ReportsModule{
		GetByIDController:    getByIDController,
		DeleteController:     deleteController,
		DownloadController:   downloadController,
		SaveReportUseCase:    saveUseCase,
		GetByIDUseCase:       getByIDUseCase,
		DeleteUseCase:        deleteUseCase,
	}
}

func (m *ReportsModule) RegisterRoutes(router *gin.Engine, protected *gin.RouterGroup) {
	router.GET("/api/reports/:id/download", m.DownloadController.GetDownload)

	api := protected.Group("/api")
	{
		api.GET("/reports/:id", m.GetByIDController.GetReportByID)
		api.DELETE("/reports/:id", m.DeleteController.DeleteReport)
	}
}
