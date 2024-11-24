package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GoogleAIService implementerar AIService för Google Gemini
type GoogleAIService struct {
	apiKey string
}

func NewGoogleAIService() *GoogleAIService {
	return &GoogleAIService{
		apiKey: getAPIKey(GeminiProvider), // Använder den gemensamma getAPIKey-funktionen från ai_service.go
	}
}

func (s *GoogleAIService) GenerateContent(prompt CVPrompt) (map[string]interface{}, error) {
	log.Printf("📝 Skickar förfrågan till Google Gemini API...")
	
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return nil, fmt.Errorf("kunde inte skapa Google AI klient: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash-8b")
	promptText := buildPrompt(prompt)

	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return nil, fmt.Errorf("fel vid generering av innehåll: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("inget svar genererades")
	}

	text := resp.Candidates[0].Content.Parts[0].(genai.Text)
	var result map[string]interface{}

	if err := json.Unmarshal([]byte(string(text)), &result); err != nil {
		return nil, fmt.Errorf("kunde inte parsa AI-svar: %v", err)
	}

	return addEmojisToResponse(result), nil
}
