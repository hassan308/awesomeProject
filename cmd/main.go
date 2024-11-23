// cmd/main.go
package main

import (
	"awesomeProject/internal/handlers"
	"awesomeProject/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"os"
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
	// Sätt debug-loggning
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("🚀 Startar applikationen...")
	
	// Registrera metrics
	prometheus.MustRegister(httpRequestsTotal)

	// Visa aktuell arbetskatalog
	if wd, err := os.Getwd(); err == nil {
		log.Printf("📂 Arbetskatalog: %s", wd)
	}

	// Visa ENV_FILE_PATH
	envPath := os.Getenv("ENV_FILE_PATH")
	log.Printf("📄 ENV_FILE_PATH: %s", envPath)

	// Lista över möjliga platser för .env fil
	possiblePaths := []string{
		".env",                    // Samma katalog som binären
		"../.env",                 // En nivå upp
		"../../.env",              // Två nivåer upp
		envPath,                   // Från miljövariabel
	}

	// Försök läsa .env från alla möjliga platser
	var loaded bool
	for _, path := range possiblePaths {
		if path == "" {
			continue
		}
		log.Printf("📂 Söker efter .env fil på: %s", path)
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ Laddade .env från: %s", path)
			loaded = true
			break
		} else {
			log.Printf("❌ Kunde inte läsa från %s: %v", path, err)
		}
	}

	if !loaded {
		log.Printf("⚠️ Kunde inte hitta .env fil på någon plats")
	}

	// Visa alla relevanta miljövariabler
	log.Printf("🔧 Miljövariabler:")
	aiProvider := os.Getenv("AI_PROVIDER")
	log.Printf("AI_PROVIDER=%s", aiProvider)
	log.Printf("HUGGINGFACE_API_KEY=%s", maskAPIKey(os.Getenv("HUGGINGFACE_API_KEY")))
	log.Printf("HUGGINGFACE_MODEL_ID=%s", os.Getenv("HUGGINGFACE_MODEL_ID"))
	log.Printf("GEMINI_API_KEY=%s", maskAPIKey(os.Getenv("GEMINI_API_KEY")))

	// Initiera AI service och visa vilken som används
	service := utils.GetAIService()
	log.Printf("🤖 Använder AI service: %T", service)
}

// Hjälpfunktion för att maskera API-nycklar i loggen
func maskAPIKey(key string) string {
	if key == "" {
		return "inte satt"
	}
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
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
	// Ladda miljövariabler från .env-fil
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Skapa en ny Gin router
	router := gin.Default()

	// Enkel CORS-konfiguration - låt Nginx hantera detaljerna
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}  // Tillåt alla origins, Nginx kommer hantera begränsningarna
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	router.Use(cors.New(config))

	// Prometheus middleware
	router.Use(prometheusMiddleware())

	// Health och metrics endpoints
	router.GET("/health", handlers.HealthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API endpoints
	router.POST("/search", handlers.SearchJobs)
	router.POST("/generate_cv", handlers.GenerateCV)
	router.POST("/analyze-search", handlers.AnalyzeSearchQuery)
	router.POST("/recommended-jobs", handlers.GetRecommendedJobs)

	// Starta servern
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
