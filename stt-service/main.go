package main

import (
	_ "stt-service/docs"
	"stt-service/domain/common/config"
	"stt-service/domain/rag"
	"stt-service/domain/speech"

	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Teralux STT Service
// @version         1.0
// @description     Speech-to-Text service for Teralux App.
// @host            localhost:8081
// @BasePath        /

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	r := gin.Default()

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize Speech Domain
	speech.InitModule(r, cfg)

	// Initialize RAG Domain
	rag.InitModule(r, cfg)

	r.Run(":" + cfg.Port)
}
