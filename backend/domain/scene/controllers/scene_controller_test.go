package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"teralux_app/domain/scene/dtos"
	"teralux_app/domain/scene/entities"
	"teralux_app/domain/scene/repositories"
	"teralux_app/domain/scene/usecases"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&entities.Scene{})
	return db
}

func TestSceneController(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB()
	repo := repositories.NewSceneRepository(db)

	addUC := usecases.NewAddSceneUseCase(repo)
	updateUC := usecases.NewUpdateSceneUseCase(repo)
	deleteUC := usecases.NewDeleteSceneUseCase(repo)
	getAllUC := usecases.NewGetAllScenesUseCase(repo)
	controlUC := usecases.NewControlSceneUseCase(repo, nil, nil)

	controller := NewSceneController(addUC, updateUC, deleteUC, getAllUC, controlUC)

	teraluxID := "test-teralux-id"

	t.Run("AddScene", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, r := gin.CreateTestContext(w)

		reqBody := dtos.CreateSceneRequestDTO{
			Name: "Test Scene",
			Actions: []dtos.ActionDTO{
				{DeviceID: "dev1", Code: "on", Value: 1},
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/teralux/"+teraluxID+"/scenes", bytes.NewBuffer(jsonBody))
		ctx.Params = gin.Params{{Key: "id", Value: teraluxID}}

		r.POST("/api/teralux/:id/scenes", controller.AddScene)
		r.ServeHTTP(w, ctx.Request)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetAllScenes", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, r := gin.CreateTestContext(w)

		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/teralux/"+teraluxID+"/scenes", nil)
		ctx.Params = gin.Params{{Key: "id", Value: teraluxID}}

		r.GET("/api/teralux/:id/scenes", controller.GetAllScenes)
		r.ServeHTTP(w, ctx.Request)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var resp struct {
			Data []dtos.SceneListResponseDTO `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "Test Scene", resp.Data[0].Name)
	})
}
