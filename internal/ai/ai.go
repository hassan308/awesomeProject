package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type AIResponse struct {
	GeneratedContent string `json:"generated_content"`
}

func GenerateCV(prompt string, apiKey string) (string, error) {
	// Läs konfiguration från .env
	maxTokens, _ := strconv.Atoi(os.Getenv("AI_MAX_TOKENS"))
	if maxTokens == 0 {
		maxTokens = 1500 // Default värde
	}

	temperature, _ := strconv.ParseFloat(os.Getenv("AI_TEMPERATURE"), 64)
	if temperature == 0 {
		temperature = 0.1 // Default värde
	}

	timeout, _ := strconv.Atoi(os.Getenv("AI_REQUEST_TIMEOUT"))
	if timeout == 0 {
		timeout = 60 // Default värde i sekunder
	}

	// Bygg upp begäran
	requestBody, err := json.Marshal(map[string]interface{}{
		"prompt":      prompt,
		"max_tokens":  maxTokens,
		"temperature": temperature,
	})
	if err != nil {
		return "", err
	}

	// Skapa HTTP-klient med timeout från .env
	client := &http.Client{
		Timeout: time.Second * time.Duration(timeout),
	}

	// Använd API URL från .env
	apiURL := os.Getenv("AI_API_URL")
	if apiURL == "" {
		apiURL = "https://generativeai.googleapis.com/v1/models/gemini-1.5-flash:generate" // Default värde
	}

	// Skapa HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// Sätt headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Resten av koden är oförändrad...
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("API request failed with status " + resp.Status)
	}

	var aiResp AIResponse
	err = json.Unmarshal(body, &aiResp)
	if err != nil {
		return "", err
	}

	return aiResp.GeneratedContent, nil
}
