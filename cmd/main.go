// cmd/main.go
package main

import (
	"awesomeProject/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	router := gin.Default()

	// CORS-konfiguration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Lägg till din frontend URL här
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Lägg till stöd för templates
	router.LoadHTMLGlob("internal/templates/*")

	// Lägg till stöd för statiska filer
	router.Static("/static", "./static")

	// API endpoints
	router.POST("/generate_cv", handlers.GenerateCV)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Startar server på port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Servern misslyckades: %v", err)
	}
}
