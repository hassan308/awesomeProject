package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type HuggingFaceRequest struct {
	Inputs     string  `json:"inputs"`
	Parameters struct {
		Temperature float64 `json:"temperature"`
		MaxTokens   int     `json:"max_tokens"`
	} `json:"parameters"`
}

type HuggingFaceResponse struct {
	GeneratedText string `json:"generated_text"`
}

// CallAI gör ett anrop till Hugging Face API med systemPrompt och användarmeddelande
func CallAI(systemPrompt string, userMessage string) (string, error) {
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	modelID := os.Getenv("HUGGINGFACE_MODEL_ID")
	
	if apiKey == "" || modelID == "" {
		return "", fmt.Errorf("HUGGINGFACE_API_KEY eller HUGGINGFACE_MODEL_ID är inte satta")
	}

	// Kombinera system prompt och användarmeddelande
	prompt := fmt.Sprintf("%s\n\nUser: %s\nAssistant:", systemPrompt, userMessage)

	requestBody := HuggingFaceRequest{
		Inputs: prompt,
	}
	requestBody.Parameters.Temperature = 0.7
	requestBody.Parameters.MaxTokens = 1000

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("kunde inte serialisera request: %v", err)
	}

	url := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", modelID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("kunde inte skapa request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("kunde inte göra API-anrop: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returnerade felstatus: %d", resp.StatusCode)
	}

	var response []HuggingFaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("kunde inte avkoda svar: %v", err)
	}

	if len(response) == 0 {
		return "", fmt.Errorf("inget svar från AI")
	}

	return response[0].GeneratedText, nil
} 