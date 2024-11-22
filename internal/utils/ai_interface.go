package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// AIService interface definierar kontraktet för alla AI-tjänster
type AIService interface {
	GenerateContent(prompt CVPrompt) (map[string]interface{}, error)
}

// AIServiceFactory skapar rätt AI-tjänst baserat på konfiguration
func GetAIService() AIService {
	provider := os.Getenv("AI_PROVIDER")
	if provider == "huggingface" {
		log.Printf("🤖 Använder Hugging Face AI med modell: %s", os.Getenv("HUGGINGFACE_MODEL_ID"))
		return NewHuggingFaceAIService()
	}
	// Default till Google
	log.Printf("🤖 Använder Google Gemini AI")
	return NewGoogleAIService()
}

// HuggingFaceAIService implementerar AIService för Hugging Face
type HuggingFaceAIService struct {
	apiKey   string
	baseURL  string
	modelID  string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type HFRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	TopP        float64   `json:"top_p"`
	Stream      bool      `json:"stream"`
}

type HFResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func NewHuggingFaceAIService() *HuggingFaceAIService {
	return &HuggingFaceAIService{
		apiKey:  os.Getenv("HUGGINGFACE_API_KEY"),
		baseURL: "https://api-inference.huggingface.co/v1/",
		modelID: getEnvOrDefault("HUGGINGFACE_MODEL_ID", "meta-llama/Llama-3.2-3B-Instruct"),
	}
}

func (s *HuggingFaceAIService) GenerateContent(prompt CVPrompt) (map[string]interface{}, error) {
	log.Printf("📝 Skickar förfrågan till Hugging Face API...")
	promptText := buildPrompt(prompt) // Använder samma prompt-builder som Google-implementationen

	req := HFRequest{
		Model: s.modelID,
		Messages: []Message{
			{
				Role:    "user",
				Content: promptText,
			},
		},
		Temperature: 0.5,
		MaxTokens:   2048,
		TopP:        0.7,
		Stream:      true,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("fel vid marshalling av request: %v", err)
	}

	request, err := http.NewRequest("POST", s.baseURL+"chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("fel vid skapande av request: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+s.apiKey)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fel vid API-anrop: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("API fel (status %d): %s", response.StatusCode, string(body))
	}

	// Hantera streaming response
	var fullResponse string
	decoder := json.NewDecoder(response.Body)
	for {
		var chunk HFResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("fel vid läsning av stream: %v", err)
		}

		if len(chunk.Choices) > 0 {
			fullResponse += chunk.Choices[0].Delta.Content
		}
	}

	// Parsa den fullständiga responsen till map[string]interface{}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(fullResponse), &result); err != nil {
		return nil, fmt.Errorf("fel vid parsing av response: %v", err)
	}

	return addEmojisToResponse(result), nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
