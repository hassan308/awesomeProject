package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// AIService interface definierar kontraktet för alla AI-tjänster
type AIService interface {
	GenerateContent(prompt CVPrompt) (map[string]string, error)
}

// AIServiceFactory skapar rätt AI-tjänst baserat på konfiguration
func GetAIService() AIService {
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	modelID := os.Getenv("HUGGINGFACE_MODEL_ID")
	service := &HuggingFaceAIService{
		apiKey:  apiKey,
		modelID: modelID,
	}
	log.Printf("🤖 Använder Hugging Face AI med modell: %s", modelID)
	return service
}

// HuggingFaceAIService implementerar AIService för Hugging Face
type HuggingFaceAIService struct {
	apiKey  string
	modelID string
}

func (s *HuggingFaceAIService) GenerateContent(prompt CVPrompt) (map[string]string, error) {
	requestBody := map[string]interface{}{
		"inputs": fmt.Sprintf("System: %s\n\nUser: %s\nAssistant:", prompt.SystemPrompt, prompt.UserPrompt),
		"parameters": map[string]interface{}{
			"max_new_tokens": 2000,
			"temperature":    0.7,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("kunde inte serialisera request: %v", err)
	}

	url := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", s.modelID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("kunde inte skapa request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", "application/json")

	log.Printf("📝 Skickar förfrågan till Hugging Face API...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kunde inte göra API-anrop: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returnerade felstatus %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response []struct {
		GeneratedText string `json:"generated_text"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kunde inte läsa svar: %v", err)
	}

	log.Printf("📥 Mottaget råsvar från API: %s", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("kunde inte avkoda svar: %v", err)
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("tomt svar från API")
	}

	// Extrahera Assistant-delen av svaret
	text := response[0].GeneratedText
	parts := strings.Split(text, "Assistant:")
	if len(parts) < 2 {
		return nil, fmt.Errorf("kunde inte hitta Assistant-svar i texten")
	}

	assistantResponse := strings.TrimSpace(parts[1])
	log.Printf("🔍 Extraherat Assistant-svar: %s", assistantResponse)

	return map[string]string{
		"response": assistantResponse,
	}, nil
}
