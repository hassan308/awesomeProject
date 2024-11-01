// cmd/main.go
package main

import (
	"awesomeProject/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
	"bytes"
)

// Definiera metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

// Prometheus middleware
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		
		// Fortsätt med request
		c.Next()

		// Efter request, registrera metriker
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// Uppdatera logging middleware
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Logga request detaljer före hantering
		log.Printf("[DEBUG] Inkommande request: %s %s", c.Request.Method, c.Request.URL.Path)
		log.Printf("[DEBUG] Headers: %v", c.Request.Header)
		
		// Om det är en POST, logga body
		if c.Request.Method == "POST" {
			bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
			// Återställ body för senare användning
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			log.Printf("[DEBUG] Request Body: %s", string(bodyBytes))
		}

		c.Next()

		// Efter request
		duration := time.Since(startTime)
		log.Printf("[DEBUG] Response Status: %d, Duration: %v, Error: %v",
			c.Writer.Status(),
			duration,
			c.Errors.String(),
		)
	}
}

func init() {
	// Registrera metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

func main() {
	// Ladda .env fil
	if err := godotenv.Load(); err != nil {
		log.Printf("Varning: Kunde inte ladda .env fil: %v", err)
	}

	// Sätt Gin i debug mode
	gin.SetMode(gin.DebugMode)
	
	router := gin.Default()

	// Använd middleware
	router.Use(loggingMiddleware())
	router.Use(prometheusMiddleware())

	// CORS-konfiguration från .env
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	// Lägg till stöd för templates från .env
	router.LoadHTMLGlob(os.Getenv("TEMPLATE_PATH"))

	// Lägg till stöd för statiska filer från .env
	router.Static("/static", os.Getenv("STATIC_PATH"))

	// API endpoints
	router.POST("/generate_cv", handlers.GenerateCV)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	router.POST("/search", handlers.SearchJobs)

	// Metrics endpoint om aktiverad
	if os.Getenv("METRICS_ENABLED") == "true" {
		router.GET(os.Getenv("METRICS_PATH"), gin.WrapH(promhttp.Handler()))
	}

	// Använd port från .env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default om ingen port är satt
	}
	
	log.Printf("Startar server på port %s i %s läge...\n", 
		port, 
		os.Getenv("ENVIRONMENT"),
	)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Servern misslyckades: %v", err)
	}
}
