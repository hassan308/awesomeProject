package handlers

import (
	"bytes"
	"context"
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
	SearchTerm string `json:"search_term"`
	MaxJobs    int    `json:"max_jobs,omitempty"`
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

	jobs, err := fetchAllJobs(c.Request.Context(), apiURL, request.SearchTerm, request.MaxJobs, maxRecords, maxRetries, retryDelay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Skapa en kanal för jobbdetaljer
	jobDetailsChan := make(chan map[string]interface{}, 100) // Minskat buffertstorlek
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 100) // Minskat till 100

	// Starta goroutines för varje jobb
	for _, job := range jobs {
		jobID, ok := job["id"].(string)
		if !ok {
			// Hoppa över jobb utan giltigt ID
			continue
		}

		wg.Add(1)
		go func(jobID string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			details, err := fetchJobDetails(c.Request.Context(), jobDetailURL, jobID, maxRetries, retryDelay)
			if err != nil {
				// Logga felet och fortsätt
				fmt.Printf("Fel vid hämtning av detaljer för jobbID %s: %v\n", jobID, err)
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

	// Kontrollera om vi har fått färre jobb än förväntat
	if len(jobDetails) < request.MaxJobs {
		fmt.Printf("Förväntade %d jobb, men fick %d\n", request.MaxJobs, len(jobDetails))
	}

	c.JSON(http.StatusOK, jobDetails)
}

func GetRecommendedJobs(c *gin.Context) {
	fmt.Println("\n=== Starting GetRecommendedJobs handler ===")
    
    apiURL, jobDetailURL, _, _, maxRetries, retryDelay := getConfig()
    fmt.Printf("\n1. API Configuration:\n")
    fmt.Printf("Base API URL: %s\n", apiURL)
    
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
        fmt.Printf("Error creating request body: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create request body: %v", err)})
        return
    }
    
    fmt.Printf("\n2. Request Details:\n")
    fmt.Printf("Request Body: %s\n", string(jsonBody))
    
    requestURL := fmt.Sprintf("%ssearch", apiURL)
    fmt.Printf("Full Request URL: %s\n", requestURL)
    
    req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
    if err != nil {
        fmt.Printf("Error creating request: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create request: %v", err)})
        return
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    fmt.Printf("\nHeaders:\n")
    for key, values := range req.Header {
        fmt.Printf("%s: %v\n", key, values)
    }

    client := &http.Client{
        Timeout: 30 * time.Second,
    }

    fmt.Println("\n3. Making API Request...")
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error making request: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Request failed: %v", err)})
        return
    }
    defer resp.Body.Close()

    fmt.Printf("\n4. API Response:\n")
    fmt.Printf("Status Code: %d\n", resp.Status)
    fmt.Printf("Response Headers:\n")
    for key, values := range resp.Header {
        fmt.Printf("%s: %v\n", key, values)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error reading response body: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read response: %v", err)})
        return
    }

    fmt.Printf("\n5. Response Body (first 500 chars):\n%s\n", string(body)[:min(len(string(body)), 500)])

    var result map[string]interface{}
    if err := json.Unmarshal(body, &result); err != nil {
        fmt.Printf("Error parsing JSON: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": fmt.Sprintf("Failed to parse response: %v", err),
            "body": string(body),
        })
        return
    }

    total := result["total"]
    fmt.Printf("\n6. Total jobs available: %v\n", total)

    ads, ok := result["ads"].([]interface{})
    if !ok {
        fmt.Printf("No ads found in response structure: %+v\n", result)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "No jobs found in response",
            "response": result,
        })
        return
    }

    fmt.Printf("Number of ads received: %d\n", len(ads))

    // Skapa en kanal för jobbdetaljer
    jobDetailsChan := make(chan map[string]interface{}, len(ads))
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // Begränsa till 10 samtidiga anrop

    // Starta goroutines för varje jobb för att hämta detaljer
    for _, ad := range ads {
        jobMap, ok := ad.(map[string]interface{})
        if !ok {
            fmt.Printf("Invalid job data format: %+v\n", ad)
            continue
        }

        jobID, ok := jobMap["id"].(string)
        if !ok || jobID == "" {
            fmt.Printf("Invalid or missing job ID: %+v\n", jobMap)
            continue
        }

        wg.Add(1)
        go func(jobID string) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire semaphore
            defer func() { <-semaphore }() // Release semaphore

            details, err := fetchJobDetails(c.Request.Context(), jobDetailURL, jobID, maxRetries, retryDelay)
            if err != nil {
                fmt.Printf("Error fetching details for job %s: %v\n", jobID, err)
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

    fmt.Printf("\n7. Final Results:\n")
    fmt.Printf("Successfully fetched details for %d jobs\n", len(jobDetails))
    if len(jobDetails) > 0 {
        prettyJSON, _ := json.MarshalIndent(jobDetails[0], "", "  ")
        fmt.Printf("First job example:\n%s\n", string(prettyJSON))
    }

    c.JSON(http.StatusOK, jobDetails)
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

func fetchAllJobs(ctx context.Context, apiURL, searchTerm string, maxJobs, maxRecords, maxRetries int, retryDelay time.Duration) ([]map[string]interface{}, error) {
	var allAds []map[string]interface{}
	seenJobs := make(map[string]bool)
	startIndex := 0
	currentTime := time.Now().UTC().Format(time.RFC3339)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Fortsätt
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
			return nil, fmt.Errorf("fel vid JSON-marshalling: %v", err)
		}

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

		// Implementera en dynamisk backoff-strategi
		time.Sleep(100 * time.Millisecond) // Justerad väntetid
	}

	return allAds, nil
}
