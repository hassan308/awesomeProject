// cmd/main.go
package main

import (
	"awesomeProject/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"os"
	"time"
)

// Definiera metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Totalt antal HTTP-förfrågningar",
		},
		[]string{"method", "endpoint", "status"},
	)
)

func init() {
	// Registrera metrics
	prometheus.MustRegister(httpRequestsTotal)
}

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			string(rune(c.Writer.Status())),
		).Inc()
	}
}

func main() {
	router := gin.Default()

	// Lägg till Prometheus middleware
	router.Use(prometheusMiddleware())

	// CORS-konfiguration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://smidra.com", "https://www.smidra.com", "http://smidra.com", "http://www.smidra.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge: 12 * time.Hour,
	}))

	// Health och metrics endpoints
	router.GET("/health", handlers.HealthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Lägg till stöd för templates
	router.LoadHTMLGlob("internal/templates/*")

	// API endpoints
	router.POST("/generate_cv", handlers.GenerateCV)
	router.POST("/search", handlers.SearchJobs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Startar server på port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Servern misslyckades: %v", err)
	}
}
