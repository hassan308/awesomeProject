package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	
	"github.com/gin-gonic/gin"
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

	jobs, err := fetchAllJobs(apiURL, request.SearchTerm, request.MaxJobs, maxRecords)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Skapa en kanal för jobbdetaljer
	jobDetailsChan := make(chan map[string]interface{}, len(jobs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 200)

	// Starta goroutines för varje jobb
	for _, job := range jobs {
		wg.Add(1)
		go func(jobID string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if details, err := fetchJobDetails(jobDetailURL, jobID); err == nil && details != nil {
				jobDetailsChan <- details
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

	c.JSON(http.StatusOK, jobDetails)
}

func fetchAllJobs(apiURL, searchTerm string, maxJobs, maxRecords int) ([]map[string]interface{}, error) {
	var allAds []map[string]interface{}
	seenJobs := make(map[string]bool)
	startIndex := 0
	currentTime := time.Now().UTC().Format(time.RFC3339)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for {
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
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
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
			break
		}

		newAdsCount := 0
		for _, ad := range ads {
			if adMap, ok := ad.(map[string]interface{}); ok {
				if jobID, ok := adMap["id"].(string); ok {
					if !seenJobs[jobID] {
						seenJobs[jobID] = true
						allAds = append(allAds, adMap)
						newAdsCount++
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

		time.Sleep(50 * time.Millisecond) // Minskad väntetid
	}

	return allAds, nil
}

const (
	maxRetries = 3
	retryDelay = 1 * time.Second
)

func fetchJobDetails(jobDetailURL, jobID string) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			time.Sleep(retryDelay * time.Duration(attempt-1))
		}

		resp, err := client.Get(jobDetailURL + jobID)
		if err != nil {
			lastErr = err
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		var details map[string]interface{}
		if err := json.Unmarshal(body, &details); err != nil {
				lastErr = err
				continue
		}

		if details["id"] == nil {
			lastErr = fmt.Errorf("invalid job details response")
			continue
		}

		return details, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
} 