package handlers

import (
	"log"
	"os"

	"awesomeProject/internal/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

func AnalyzeSearchQuery(c *gin.Context) {
	var requestBody struct {
		Query string `json:"query"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if requestBody.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query cannot be empty"})
		return
	}

	// Logga vilken AI-tjänst som används
	aiProvider := os.Getenv("AI_PROVIDER")
	if aiProvider == "" {
		aiProvider = "huggingface" // Default till Hugging Face om ingen är specificerad
	}
	log.Printf("Använder AI-tjänst: %s för analys av sökfråga: '%s'", aiProvider, requestBody.Query)

	// Analysera sökfrågan med AI-tjänsten
	analysis, err := utils.AnalyzeSearchQuery(requestBody.Query)
	if err != nil {
		log.Printf("Error analyzing search query with %s: %v", aiProvider, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to analyze search query: %v", err)})
		return
	}

	log.Printf("AI-analys slutförd med %s. Extraherad information - Jobb: '%s', Kommun: '%s'", 
		aiProvider, analysis.Job, analysis.Municipality)

	// Använd municipality som separat parameter istället för att lägga till det i söktermen
	searchTerm := analysis.Job
	log.Printf("Söker efter jobb med term: '%s' i kommun: '%s'", searchTerm, analysis.Municipality)

	// Använd standard sökfunktionen med den extraherade söktermen och kommun
	apiURL, jobDetailURL, maxRecords, defaultMaxJobs, maxRetries, retryDelay := getConfig()

	// Lägg till workExtent i sökningen om det finns
	var workExtentFilter map[string]string
	if analysis.WorkExtent != "" {
		workExtentFilter = map[string]string{
			"type":  "workExtent",
			"value": analysis.WorkExtent,
		}
	}

	// Lägg till remote i sökningen om det finns
	var remoteFilter map[string]string
	if analysis.Remote == "true" {
		remoteFilter = map[string]string{
			"type":  "remote",
			"value": "true",
		}
	}

	// Lägg till körkortskrav i sökningen om det finns
	var drivingLicenseFilter map[string]string
	if analysis.DrivingLicense == "false" {
		drivingLicenseFilter = map[string]string{
			"type":  "drivingLicenseRequired",
			"value": "false",
		}
	}

	// Skapa en slice med alla filter
	var filters []map[string]string
	if workExtentFilter != nil {
		filters = append(filters, workExtentFilter)
	}
	if remoteFilter != nil {
		filters = append(filters, remoteFilter)
	}
	if drivingLicenseFilter != nil {
		filters = append(filters, drivingLicenseFilter)
	}

	jobs, err := fetchAllJobs(c.Request.Context(), apiURL, searchTerm, analysis.Municipality, defaultMaxJobs, maxRecords, maxRetries, retryDelay, filters)
	if err != nil {
		log.Printf("Fel vid jobbsökning: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch jobs: %v", err)})
		return
	}

	log.Printf("Hittade %d jobb i initial sökning", len(jobs))

	// Hämta detaljerad information för varje jobb
	var jobDetails []map[string]interface{}
	jobDetailsChan := make(chan map[string]interface{}, len(jobs))
	
	// Starta goroutines för att hämta jobbdetaljer
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)
	
	for _, job := range jobs {
		wg.Add(1)
		go func(job map[string]interface{}) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			jobID, _ := job["id"].(string)
			details, err := fetchJobDetails(c.Request.Context(), jobDetailURL, jobID, maxRetries, retryDelay)
			if err != nil {
				log.Printf("Kunde inte hämta detaljer för jobb %s: %v", jobID, err)
				return
			}
			jobDetailsChan <- details
		}(job)
	}
	
	// Vänta på att alla goroutines är klara
	go func() {
		wg.Wait()
		close(jobDetailsChan)
	}()
	
	// Samla alla jobbdetaljer och filtrera
	filteredCount := 0
	totalCount := 0
	
	for detail := range jobDetailsChan {
		totalCount++
		requiresExp, _ := detail["requiresExperience"].(bool)
		
		if analysis.RequiresExperience != nil && *analysis.RequiresExperience == false {
			if requiresExp {
				continue
			}
		}
		
		filteredCount++
		jobDetails = append(jobDetails, detail)
	}
	
	log.Printf("\n=== SÖKRESULTAT ===")
	log.Printf("Totalt antal jobb: %d", totalCount)
	log.Printf("Antal jobb efter filtrering: %d", filteredCount)
	
	c.JSON(http.StatusOK, gin.H{
		"jobs":     jobDetails,
		"analysis": analysis,
	})
}
