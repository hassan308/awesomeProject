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
	"strings"
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

	// Hämta tillåtna ursprung från miljövariabel eller använd default
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	origins := []string{"https://smidra.com", "https://www.smidra.com"}
	
	if gin.Mode() == gin.DebugMode {
		origins = append(origins, "http://localhost:3000")
	}
	
	if allowedOrigins != "" {
		origins = append(origins, strings.Split(allowedOrigins, ",")...)
	}

	config := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// Använd CORS-middleware
	router.Use(cors.New(config))

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
