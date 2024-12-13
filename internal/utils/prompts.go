package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"log"
)

// CoverLetterPrompt innehåller information för att generera personligt brev
type CoverLetterPrompt struct {
	Template     string `json:"template"`
	JobTitle     string `json:"job_title"`
	JobDesc      string `json:"job_desc"`
	CompanyName  string `json:"company_name"`
}

// GeneratePersonalLetter genererar innehåll för personligt brev med hjälp av AI
func GeneratePersonalLetter(prompt CoverLetterPrompt) (map[string]interface{}, error) {
	log.Printf("Genererar personligt brev för tjänst: %s hos %s", prompt.JobTitle, prompt.CompanyName)

	// Skapa system prompt för AI
	systemPrompt := `Du är en expert på att skriva personliga brev. 
	Din uppgift är att generera ett JSON-objekt som innehåller information för ett personligt brev.
	Svara ENDAST med ett giltigt JSON-objekt, inget annat.
	
	JSON-objektet ska ha följande struktur och fält:
	{
		"namn": "Ett lämpligt namn",
		"titel": "En passande yrkestitel",
		"email": "En professionell e-postadress",
		"telefon": "Ett svenskt telefonnummer",
		"adress": "En svensk adress",
		"mottagare_namn": "Ett lämpligt namn på rekryteraren",
		"mottagare_foretag": "` + prompt.CompanyName + `",
		"mottagare_position": "Rekryteringsansvarig",
		"datum": "` + time.Now().Format("2006-01-02") + `",
		"inledning": "En engagerande inledning som refererar till tjänsten",
		"huvudtext": "En övertygande huvudtext som matchar jobbeskrivningen",
		"avslutning": "En professionell avslutning",
		"halsningsfras": "Med vänliga hälsningar"
	}`

	// Skapa användarmeddelande med jobbinformation
	userPrompt := fmt.Sprintf(`Generera ett personligt brev för följande tjänst:
	Titel: %s
	Företag: %s
	Beskrivning: %s`, 
	prompt.JobTitle, 
	prompt.CompanyName, 
	prompt.JobDesc)

	log.Printf("=== AI Prompt för Personligt Brev ===")
	log.Printf("System Prompt:\n%s", systemPrompt)
	log.Printf("\nUser Prompt:\n%s", userPrompt)

	// Använd den gemensamma AI-tjänsten
	aiService := GetAIService()
	log.Printf("\nAnropar AI-tjänst")
	
	aiResponse, err := aiService.GenerateContent(CVPrompt{
		JobTitle:       prompt.JobTitle,
		JobDescription: prompt.JobDesc,
		SystemPrompt:   systemPrompt,
		UserPrompt:     userPrompt,
	})

	if err != nil {
		log.Printf("❌ Fel vid AI-anrop: %v", err)
		return nil, fmt.Errorf("AI-anrop misslyckades: %v", err)
	}

	// Hämta response-strängen direkt från map[string]string
	responseStr := aiResponse["response"]
	if responseStr == "" {
		log.Printf("❌ Tomt svar från AI")
		return nil, fmt.Errorf("tomt svar från AI")
	}

	log.Printf("✅ AI-svar mottaget, längd: %d tecken", len(responseStr))

	// Försök hitta JSON i svaret
	startIndex := strings.Index(responseStr, "{")
	endIndex := strings.LastIndex(responseStr, "}")
	
	if startIndex == -1 || endIndex == -1 || startIndex > endIndex {
		log.Printf("❌ Kunde inte hitta giltigt JSON i svaret. Rått svar:\n%s", responseStr)
		return nil, fmt.Errorf("kunde inte hitta giltigt JSON i svaret")
	}

	jsonStr := responseStr[startIndex:endIndex+1]
	log.Printf("📝 Extraherat JSON:\n%s", jsonStr)

	// Konvertera svaret till map
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("❌ JSON-parsning misslyckades: %v", err)
		return nil, fmt.Errorf("kunde inte parsa JSON: %v\nSvar: %s", err, jsonStr)
	}

	log.Printf("✅ Personligt brev genererat framgångsrikt")
	return result, nil
} 