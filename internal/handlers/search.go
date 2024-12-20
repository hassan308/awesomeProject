package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"awesomeProject/internal/data"
)

// Konfigurationsinställningar hämtas från miljövariabler
func getConfig() (string, string, int, int, int, time.Duration) {
	apiURL := os.Getenv("PLATSBANKEN_API_URL")
	if apiURL == "" {
		apiURL = "https://platsbanken-api.arbetsformedlingen.se/jobs/v1/"
	}

	jobDetailURL := os.Getenv("PLATSBANKEN_JOB_DETAIL_URL")
	if jobDetailURL == "" {
		jobDetailURL = "https://platsbanken-api.arbetsformedlingen.se/jobs/v1/job/"
	}

	maxRecords := 100 // Standardvärde
	if val := os.Getenv("PLATSBANKEN_MAX_RECORDS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			maxRecords = n
		}
	}

	defaultMaxJobs := 500 // Ändrat från 1000 till 500
	if val := os.Getenv("PLATSBANKEN_DEFAULT_MAX_JOBS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			defaultMaxJobs = n
		}
	}

	maxRetries := 3
	if val := os.Getenv("PLATSBANKEN_MAX_RETRIES"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			maxRetries = n
		}
	}

	retryDelay := 1 * time.Second
	if val := os.Getenv("PLATSBANKEN_RETRY_DELAY"); val != "" {
		if d, err := time.ParseDuration(val); err == nil && d > 0 {
			retryDelay = d
		}
	}

	return apiURL, jobDetailURL, maxRecords, defaultMaxJobs, maxRetries, retryDelay
}

type SearchRequest struct {
	SearchTerm         string `json:"search_term"`
	MaxJobs           int    `json:"max_jobs,omitempty"`
	Municipality      string `json:"municipality,omitempty"`
	RequiresExperience *bool  `json:"requiresExperience,omitempty"`
}

func SearchJobs(c *gin.Context) {
	apiURL, jobDetailURL, maxRecords, defaultMaxJobs, maxRetries, retryDelay := getConfig()

	var request SearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ogiltigt förfrågningsformat"})
		return
	}

	if request.SearchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sökterm krävs"})
		return
	}

	if request.MaxJobs == 0 {
		request.MaxJobs = defaultMaxJobs
	}

	log.Printf("Söker efter jobb med term: '%s' i kommun: '%s'", request.SearchTerm, request.Municipality)

	jobs, err := fetchAllJobs(c.Request.Context(), apiURL, request.SearchTerm, request.Municipality, request.MaxJobs, maxRecords, maxRetries, retryDelay, nil)
	if err != nil {
		log.Printf("Fel vid jobbsökning: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Hittade %d jobb i initial sökning", len(jobs))

	// Skapa en map för att spara originaldata
	originalJobData := make(map[string]map[string]interface{})
	for _, job := range jobs {
		if jobID, ok := job["id"].(string); ok {
			originalJobData[jobID] = job
		}
	}

	// Skapa en kanal för jobbdetaljer
	jobDetailsChan := make(chan map[string]interface{}, len(jobs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Begränsa till 10 samtidiga anrop

	// Starta goroutines för varje jobb
	for _, job := range jobs {
		jobID, ok := job["id"].(string)
		if !ok {
			continue
		}

		wg.Add(1)
		go func(jobID string) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			details, err := fetchJobDetails(c.Request.Context(), jobDetailURL, jobID, maxRetries, retryDelay)
			if err != nil {
				log.Printf("Kunde inte hämta detaljer för jobb %s: %v", jobID, err)
				return
			}
			if details != nil {
				jobDetailsChan <- details
			}
		}(jobID)
	}

	go func() {
		wg.Wait()
		close(jobDetailsChan)
	}()

	// Samla alla jobbdetaljer
	var jobDetails []map[string]interface{}
	totalCount := 0
	
	for detail := range jobDetailsChan {
		totalCount++
		jobDetails = append(jobDetails, detail)
	}

	log.Printf("\n=== SÖKRESULTAT ===")
	log.Printf("Totalt antal jobb: %d", totalCount)
	log.Printf("Antal jobb efter filtrering: %d", len(jobDetails))

	response := gin.H{
		"jobs": jobDetails,
		"debug": gin.H{
			"totalJobsBeforeFilter": len(jobs),
			"totalJobsAfterFilter":  len(jobDetails),
			"searchQuery":           request.SearchTerm,
			"municipality":          request.Municipality,
		},
	}

	c.JSON(http.StatusOK, response)
}

func GetRecommendedJobs(c *gin.Context) {
    apiURL, jobDetailURL, _, _, maxRetries, retryDelay := getConfig()
    
    // Skapa request body för Platsbanken API
    requestBody := map[string]interface{}{
        "filters":    []interface{}{},
        "fromDate":   nil,
        "order":      "relevance",
        "maxRecords": 25,
        "startIndex": 0,
        "toDate":     time.Now().Format(time.RFC3339),
        "source":     "pb",
    }
    
    jsonBody, err := json.Marshal(requestBody)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create request body: %v", err)})
        return
    }
    
    requestURL := fmt.Sprintf("%ssearch", apiURL)
    
    req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create request: %v", err)})
        return
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    client := &http.Client{
        Timeout: 30 * time.Second,
    }

    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Request failed: %v", err)})
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read response: %v", err)})
        return
    }

    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("Failed to parse response: %v", err),
            "body": string(body),
        })
        return
    }

    ads, ok := result["ads"].([]interface{})
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "No jobs found in response",
            "response": result,
        })
        return
    }

    // Skapa en kanal för jobbdetaljer
    jobDetailsChan := make(chan map[string]interface{}, len(ads))
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // Begränsa till 10 samtidiga anrop

    // Starta goroutines för varje jobb för att hämta detaljer
    for _, ad := range ads {
        jobMap, ok := ad.(map[string]interface{})
        if !ok {
            continue
        }

        jobID, ok := jobMap["id"].(string)
        if !ok || jobID == "" {
            continue
        }

        wg.Add(1)
        go func(jobID string) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire semaphore
            defer func() { <-semaphore }() // Release semaphore

            details, err := fetchJobDetails(c.Request.Context(), jobDetailURL, jobID, maxRetries, retryDelay)
            if err != nil {
                return
            }
            if details != nil {
                jobDetailsChan <- details
            }
        }(jobID)
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

    // Kontrollera att jobDetails är en array innan vi skickar
    if jobDetails == nil {
        jobDetails = []map[string]interface{}{} // Tom array istället för nil
    }

    response := gin.H{
        "jobs": jobDetails,
        "debug": gin.H{
            "totalJobs": len(jobDetails),
            "timestamp": time.Now().Format(time.RFC3339),
        },
    }

    c.JSON(http.StatusOK, response)
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// Helper function to safely get nested string values
func getNestedString(m map[string]interface{}, keys ...string) string {
    current := m
    for _, key := range keys[:len(keys)-1] {
        if next, ok := current[key].(map[string]interface{}); ok {
            current = next
        } else {
            return ""
        }
    }
    
    lastKey := keys[len(keys)-1]
    if val, ok := current[lastKey].(string); ok {
        return val
    }
    return ""
}

func fetchJobDetails(ctx context.Context, jobDetailURL, jobID string, maxRetries int, retryDelay time.Duration) (map[string]interface{}, error) {
	url := jobDetailURL + jobID

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("kunde inte skapa request: %v", err)
		}

		req = req.WithContext(ctx)
		req.Header.Set("Accept", "application/json")
		
		// Add API key if available
		if apiKey := os.Getenv("PLATSBANKEN_API_KEY"); apiKey != "" {
			req.Header.Set("api-key", apiKey)
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("kunde inte hämta jobbdetaljer efter %d försök: %v", maxRetries, err)
			}
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("fick status %d vid hämtning av jobbdetaljer", resp.StatusCode)
			}
			time.Sleep(retryDelay)
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("kunde inte avkoda jobbdetaljer: %v", err)
		}

		return result, nil
	}

	return nil, fmt.Errorf("kunde inte hämta jobbdetaljer efter %d försök", maxRetries)
}

func fetchAllJobs(ctx context.Context, apiURL, searchTerm, municipalityName string, maxJobs, maxRecords, maxRetries int, retryDelay time.Duration, additionalFilters []map[string]string) ([]map[string]interface{}, error) {
	var allAds []map[string]interface{}
	seenJobs := make(map[string]bool)
	startIndex := 0
	
	currentTime := time.Now().Format(time.RFC3339)

	var locationFilter map[string]string
	if municipalityName != "" {
		log.Printf("\n=== KONVERTERAR PLATS ===")
		if strings.Contains(municipalityName, "län") {
			municipalityID := data.GetMunicipalityID(municipalityName)
			log.Printf("🔍 Län: '%s'", municipalityName)
			log.Printf("🎯 ID: '%s'", municipalityID)
			locationFilter = map[string]string{
				"type":  "region",
				"value": municipalityID,
			}
		} else {
			municipalityID := data.GetMunicipalityID(municipalityName)
			log.Printf("🔍 Kommun: '%s'", municipalityName)
			log.Printf("🎯 ID: '%s'", municipalityID)
			if municipalityID != "" {
				locationFilter = map[string]string{
					"type":  "municipality",
					"value": municipalityID,
				}
			} else {
				log.Printf("⚠️ Kunde inte hitta ID för kommun: %s", municipalityName)
			}
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		currentMaxRecords := maxRecords
		if maxJobs > 0 {
			remaining := maxJobs - len(allAds)
			if remaining <= 0 {
				break
			}
			if remaining < currentMaxRecords {
				currentMaxRecords = remaining
			}
		}

		filters := []map[string]string{
			{"type": "freetext", "value": searchTerm},
		}
		
		if locationFilter != nil {
			filters = append(filters, locationFilter)
		}

		// Lägg till alla ytterligare filter
		if len(additionalFilters) > 0 {
			filters = append(filters, additionalFilters...)
		}

		payload := map[string]interface{}{
			"filters":    filters,
			"fromDate":   nil,
			"order":      "relevance",
			"maxRecords": currentMaxRecords,
			"startIndex": startIndex,
			"toDate":     currentTime,
			"source":     "pb",
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("fel vid JSON-marshalling: %v", err)
		}

		// Logga payload som skickas till Arbetsförmedlingen
		log.Printf("\n=== PAYLOAD TILL ARBETSFÖRMEDLINGEN ===")
		log.Printf("URL: %s", apiURL+"search")
		prettyJSON, _ := json.MarshalIndent(payload, "", "  ")
		log.Printf("Payload:\n%s", string(prettyJSON))

		req, err := http.NewRequestWithContext(ctx, "POST", apiURL+"search", bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("fel vid skapande av HTTP-förfrågan: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fel vid HTTP-förfrågan: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("API returnerade status %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("fel vid JSON-dekodning: %v", err)
		}
		resp.Body.Close()

		ads, ok := result["ads"].([]interface{})
		if !ok || len(ads) == 0 {
			break
		}

		newAdsCount := 0
		for _, ad := range ads {
			if adMap, ok := ad.(map[string]interface{}); ok {
				jobID, ok := adMap["id"].(string)
				if ok && !seenJobs[jobID] {
					seenJobs[jobID] = true
					allAds = append(allAds, adMap)
					newAdsCount++
					if maxJobs > 0 && len(allAds) >= maxJobs {
						break
					}
				}
			}
		}

		if newAdsCount == 0 {
			break
		}

		startIndex += len(ads)

		if maxJobs > 0 && len(allAds) >= maxJobs {
			allAds = allAds[:maxJobs]
			break
		}

		if len(ads) < currentMaxRecords {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	return allAds, nil
}
