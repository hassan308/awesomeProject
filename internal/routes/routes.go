package routes

import (
	"github.com/gin-gonic/gin"
	"awesomeProject/internal/handlers"
)

func SetupRoutes(router *gin.Engine) {
	// CV routes
	router.POST("/api/generate-cv", handlers.GenerateCV)
	
	// Cover letter routes
	router.POST("/api/generate-cover-letter", handlers.GenerateCoverLetter)
	router.POST("/api/generate-ai-cover-letter", handlers.GenerateAICoverLetter)

	// Job routes
	router.POST("/api/search", handlers.SearchJobs)
	router.POST("/api/analyze-search", handlers.AnalyzeSearchQuery)
	router.POST("/api/recommended-jobs", handlers.GetRecommendedJobs)
} 