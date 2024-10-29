// cmd/main.go
package main

import (
	"awesomeProject/internal/utils"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	// Ladda JSON-data
	cvData, err := utils.LoadCVData("data/cv_data.json")
	if err != nil {
		log.Fatalf("Fel vid laddning av CV-data: %v", err)
	}

	// Parallellbearbeta sektioner
	utils.ProcessSectionsConcurrently(cvData)

	// Ladda och parsea HTML-template
	tmplPath := filepath.Join("internal", "templates", "cv_template.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Fatalf("Fel vid parsing av template: %v", err)
	}

	// Skapa en Gin router
	router := gin.Default()

	// Ställ in HTML-rendering med Gin
	router.SetHTMLTemplate(tmpl)

	// Definiera en GET-endpoint för att generera CV
	router.GET("/generate-cv", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.HTML(http.StatusOK, "cv_template.html", cvData)
	})

	// Starta servern på port 8080
	port := "8080"
	log.Printf("Startar server på port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Servern misslyckades: %v", err)
	}
}
