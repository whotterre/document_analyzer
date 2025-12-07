package routes

import (
	"whotterre/doculyzer/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *handlers.DocumentHandler) {
	r.POST("/documents/upload", h.UploadDocument)
	r.POST("/documents/:id/analyze", h.AnalyzeDocument)
	r.GET("/documents/:id", h.GetDocument)
}
