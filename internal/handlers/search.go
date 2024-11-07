package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
	"os"
	"strconv"
	"github.com/gin-gonic/gin"
	"sync"
)

// Använd miljövariabler istället för konstanter
func getConfig() (string, string, int, int) {
	apiURL := os.Getenv("PLATSBANKEN_API_URL")
	if apiURL == "" {
		apiURL = "https://platsbanken-api.arbetsformedlingen.se/jobs/v1/search"
	}

	jobDetailURL := os.Getenv("PLATSBANKEN_JOB_DETAIL_URL")
	if jobDetailURL == "" {
		jobDetailURL = "https://platsbanken-api.arbetsformedlingen.se/jobs/v1/job/"
	}

	maxRecords := 100 // Default värde
	if val := os.Getenv("PLATSBANKEN_MAX_RECORDS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			maxRecords = n
		}
	}

	defaultMaxJobs := 500 // Ändrat från 1000 till 500
	if val := os.Getenv("PLATSBANKEN_DEFAULT_MAX_JOBS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			defaultMaxJobs = n
		}
	}

	return apiURL, jobDetailURL, maxRecords, defaultMaxJobs
}

type SearchRequest struct {
	SearchTerm string `json:"search_term"`
	MaxJobs    int    `json:"max_jobs,omitempty"`
}

func SearchJobs(c *gin.Context) {
	apiURL, jobDetailURL, maxRecords, defaultMaxJobs := getConfig()

	var request SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if request.SearchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
		return
	}

	if request.MaxJobs == 0 {
		request.MaxJobs = defaultMaxJobs
	}

	log.Printf("=== Ny sökning påbörjad ===")
	log.Printf("Sökparametrar: term='%s', maxJobs=%d", request.SearchTerm, request.MaxJobs)

	jobs, err := fetchAllJobs(apiURL, request.SearchTerm, request.MaxJobs, maxRecords)
	if err != nil {
		log.Printf("❌ FEL vid jobbsökning: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("✅ Hittade totalt %d jobb för söktermen '%s'", len(jobs), request.SearchTerm)

	// Skapa en kanal för jobbdetaljer
	jobDetailsChan := make(chan map[string]interface{}, len(jobs))
	// Skapa en WaitGroup för att vänta på alla goroutines
	var wg sync.WaitGroup

	// Begränsa antalet samtidiga requests med en semaphore
	semaphore := make(chan struct{}, 200) // Ändrat från 50 till 200 samtidiga requests

	// Starta goroutines för varje jobb
	for _, job := range jobs {
		wg.Add(1)
		go func(jobID string) {
			defer wg.Done()
			
			// Använd semaphore för att begränsa samtidiga requests
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			log.Printf("🔍 Hämtar detaljer för jobb ID: %s", jobID)
			if details, err := fetchJobDetails(jobDetailURL, jobID); err == nil && details != nil {
				log.Printf("✅ Lyckades hämta detaljer för jobb ID: %s", jobID)
				if title, ok := details["headline"].(string); ok {
					log.Printf("📋 Jobbtitel: %s", title)
				}
				jobDetailsChan <- details
			} else if err != nil {
				log.Printf("❌ Fel vid hämtning av jobbdetaljer för ID %s: %v", jobID, err)
			}
		}(job["id"].(string))
	}

	// Starta en goroutine för att stänga kanalen när alla jobb är klara
	go func() {
		wg.Wait()
		close(jobDetailsChan)
	}()

	// Samla alla jobbdetaljer från kanalen
	var jobDetails []map[string]interface{}
	for detail := range jobDetailsChan {
		jobDetails = append(jobDetails, detail)
	}

	log.Printf("Hämtade jobbdetaljer klart - antal_med_detaljer: %d, sökterm: %s", 
		len(jobDetails), request.SearchTerm)

	c.JSON(http.StatusOK, jobDetails)
}

func fetchAllJobs(apiURL, searchTerm string, maxJobs, maxRecords int) ([]map[string]interface{}, error) {
	var allAds []map[string]interface{}
	startIndex := 0
	currentTime := time.Now().UTC().Format(time.RFC3339)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	log.Printf("=== Startar jobbsökning ===")
	log.Printf("🔍 Sökterm: %s", searchTerm)
	log.Printf("📊 Max antal jobb att hämta: %d", maxJobs)
	log.Printf("📊 Max antal per request: %d", maxRecords)

	for {
		currentMaxRecords := maxRecords
		if maxJobs > 0 {
			remaining := maxJobs - len(allAds)
			if remaining < currentMaxRecords {
				currentMaxRecords = remaining
			}
		}

		payload := map[string]interface{}{
			"filters": []map[string]string{
				{"type": "freetext", "value": searchTerm},
			},
			"fromDate":   nil,
			"order":      "date",
			"source":     "pb",
			"startIndex": startIndex,
			"toDate":     currentTime,
			"maxRecords": currentMaxRecords,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		log.Printf("Gör API-anrop - startIndex: %d, maxRecords: %d", startIndex, currentMaxRecords)
		
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Fel vid API-anrop: %v", err)
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("❌ API svarade med status: %d", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			log.Printf("📝 API svar: %s", string(body))
			resp.Body.Close()
			break
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		ads, ok := result["ads"].([]interface{})
		if !ok || len(ads) == 0 {
			log.Printf("Inga fler annonser hittades, avbryter. Totalt antal: %d", len(allAds))
			break
		}

		log.Printf("✅ Hämtade %d nya jobb i denna batch", len(ads))
		log.Printf("📊 Totalt antal hämtade jobb: %d", len(allAds))
		
		for i, ad := range ads {
			if adMap, ok := ad.(map[string]interface{}); ok {
				allAds = append(allAds, adMap)
				
				// Logga detaljerad information för de första 2 jobben
				if i < 2 {
					prettyJSON, err := json.MarshalIndent(adMap, "", "    ")
					if err == nil {
						log.Printf("🔍 Detaljerad information för jobb %d:\n%s", i+1, string(prettyJSON))
					}
				}
			}
		}

		startIndex += len(ads)
		log.Printf("Totalt antal hämtade annonser: %d", len(allAds))

		if maxJobs > 0 && len(allAds) >= maxJobs {
			log.Printf("Nått målantal annonser (%d), avbryter", maxJobs)
			allAds = allAds[:maxJobs]
			break
		}

		// Om vi fick färre annonser än begärt finns inga fler att hämta
		if len(ads) < currentMaxRecords {
			log.Printf("Färre annonser än begärt returnerades (%d < %d), inga fler finns", 
				len(ads), currentMaxRecords)
			break
		}

		// Lägg till en kort paus mellan anropen för att inte överbelasta API:et
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("=== Jobbsökning avslutad ===")
	log.Printf("📊 Slutligt antal hämtade jobb: %d", len(allAds))
	return allAds, nil
}

func fetchJobDetails(jobDetailURL, jobID string) (map[string]interface{}, error) {
	log.Printf("🔍 Hämtar detaljer för jobb %s", jobID)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(jobDetailURL + jobID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Kunde inte hämta detaljer för jobb %s, status: %d", jobID, resp.StatusCode)
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var details map[string]interface{}
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, err
	}

	// Logga detaljerad information för de första 2 jobben
	if jobID == details["id"] && (len(details) > 0) {
		prettyJSON, err := json.MarshalIndent(details, "", "    ")
		if err == nil {
			log.Printf("📋 Detaljerad jobbinformation för %s:\n%s", jobID, string(prettyJSON))
		}
	}

	log.Printf("✅ Hämtade detaljer för jobb %s", jobID)
	return details, nil
} 