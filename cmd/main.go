package main

import (
	"whotterre/doculyzer/internal/config"
	"whotterre/doculyzer/internal/handlers"
	"whotterre/doculyzer/internal/initializers"
	"whotterre/doculyzer/internal/repository"
	"whotterre/doculyzer/internal/routes"
	"whotterre/doculyzer/internal/services"

	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadConfig()
	initializers.ConnectToDB()
	initializers.ConnectToS3()
}

func main() {
	repo := repository.NewDocumentRepository(initializers.DB)
	service := services.NewDocumentService(repo, initializers.S3Client, config.AppConfig.S3Bucket)
	handler := handlers.NewDocumentHandler(service)

	r := gin.Default()
	routes.SetupRoutes(r, handler)

	r.Run()
}
