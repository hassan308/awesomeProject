package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

type AIResponse struct {
	GeneratedContent string `json:"generated_content"`
}

func GenerateCV(prompt string, apiKey string) (string, error) {
	// Bygg upp begäran enligt API-dokumentationen
	requestBody, err := json.Marshal(map[string]interface{}{
		"prompt":      prompt,
		"max_tokens":  1500, // Anpassa efter behov
		"temperature": 0.1,  // Anpassa efter behov
	})
	if err != nil {
		return "", err
	}

	// Skapa en HTTP-klient med timeout
	client := &http.Client{
		Timeout: time.Second * 60,
	}

	// Bygg API URL baserat på Googles dokumentation
	apiURL := "https://generativeai.googleapis.com/v1/models/gemini-1.5-flash:generate"

	// Skapa en ny HTTP POST-begäran
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Sätt nödvändiga headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey) // Anpassa om nödvändigt

	// Utför begäran
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Läs responsen
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("API request failed with status " + resp.Status)
	}

	// Parse JSON-responsen
	var aiResp AIResponse
	err = json.Unmarshal(body, &aiResp)
	if err != nil {
		return "", err
	}

	return aiResp.GeneratedContent, nil
}
