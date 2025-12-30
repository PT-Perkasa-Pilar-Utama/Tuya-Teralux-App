package teralux

import (
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/teralux/controllers/teralux"
	"teralux_app/domain/teralux/repositories"
	"teralux_app/domain/teralux/routes"
	usecases "teralux_app/domain/teralux/usecases/teralux"

	"github.com/gin-gonic/gin"
)

// TeraluxModule encapsulates Teralux domain components
type TeraluxModule struct {
	CreateController    *controllers.CreateTeraluxController
	GetAllController    *controllers.GetAllTeraluxController
	GetByIDController   *controllers.GetTeraluxByIDController
	UpdateController    *controllers.UpdateTeraluxController
	DeleteController    *controllers.DeleteTeraluxController
}

// NewTeraluxModule initializes the Teralux module
func NewTeraluxModule(badger *persistence.BadgerService) *TeraluxModule {
	// Repository
	repository := repositories.NewTeraluxRepository(badger)

	// Use Cases
	createUseCase := usecases.NewCreateTeraluxUseCase(repository)
	getAllUseCase := usecases.NewGetAllTeraluxUseCase(repository)
	getByIDUseCase := usecases.NewGetTeraluxByIDUseCase(repository)
	updateUseCase := usecases.NewUpdateTeraluxUseCase(repository)
	deleteUseCase := usecases.NewDeleteTeraluxUseCase(repository)

	// Controllers
	return &TeraluxModule{
		CreateController:    controllers.NewCreateTeraluxController(createUseCase),
		GetAllController:    controllers.NewGetAllTeraluxController(getAllUseCase),
		GetByIDController:   controllers.NewGetTeraluxByIDController(getByIDUseCase),
		UpdateController:    controllers.NewUpdateTeraluxController(updateUseCase),
		DeleteController:    controllers.NewDeleteTeraluxController(deleteUseCase),
	}
}

// RegisterRoutes registers Teralux routes
func (m *TeraluxModule) RegisterRoutes(protected *gin.RouterGroup) {
	routes.SetupTeraluxRoutes(
		protected,
		m.CreateController,
		m.GetAllController,
		m.GetByIDController,
		m.UpdateController,
		m.DeleteController,
	)
}
