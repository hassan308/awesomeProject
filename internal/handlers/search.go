package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
	"os"
	"strconv"
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

	defaultMaxJobs := 300 // Default värde
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

	jobDetails := make([]map[string]interface{}, 0)
	for _, job := range jobs {
		if details, err := fetchJobDetails(jobDetailURL, job["id"].(string)); err == nil && details != nil {
			jobDetails = append(jobDetails, details)
		}
	}

	c.JSON(http.StatusOK, jobDetails)
}

func fetchAllJobs(apiURL, searchTerm string, maxJobs, maxRecords int) ([]map[string]interface{}, error) {
	var allAds []map[string]interface{}
	startIndex := 0
	currentTime := time.Now().UTC().Format(time.RFC3339)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

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

		for _, ad := range ads {
			if adMap, ok := ad.(map[string]interface{}); ok {
				allAds = append(allAds, adMap)
			}
		}

		startIndex += len(ads)

		if maxJobs > 0 && len(allAds) >= maxJobs {
			allAds = allAds[:maxJobs]
			break
		}

		if len(ads) < currentMaxRecords {
			break
		}

		time.Sleep(80 * time.Millisecond)
	}

	return allAds, nil
}

func fetchJobDetails(jobDetailURL, jobID string) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(jobDetailURL + jobID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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

	return details, nil
} 