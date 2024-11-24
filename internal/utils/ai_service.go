package utils

import (
	"bytes"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIProvider string

const (
	GeminiProvider     AIProvider = "gemini"
	HuggingfaceProvider AIProvider = "huggingface"
)

type CVPrompt struct {
	Name           string
	JobTitle       string
	JobDescription string
	Experience     string
	Education      string
	Skills         string
	Certifications string
	Bio            string
	Email          string
	Phone          string
	Location       string
}

// HFStreamResponse representerar svarsstrukturen från Hugging Face streaming API
type HFStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func getAPIKey(provider AIProvider) string {
	var key string
	switch provider {
	case GeminiProvider:
		key = os.Getenv("GEMINI_API_KEY")
		if key == "" {
			key = "AIzaSyA2UxB4f5SfEI-lf35044uZEiVxCNcDigE"
			log.Println("Varning: Använder default Gemini API-nyckel. Sätt GEMINI_API_KEY miljövariabel i produktion.")
		}
	case HuggingfaceProvider:
		key = os.Getenv("HUGGINGFACE_API_KEY")
		if key == "" {
			log.Println("Varning: Ingen Huggingface API-nyckel hittades. Sätt HUGGINGFACE_API_KEY miljövariabel.")
		}
	}
	return key
}

func GenerateAIContent(prompt CVPrompt) (map[string]interface{}, error) {
	result, err := generateWithHuggingface(prompt)
	if err != nil {
		log.Printf("Hugging Face misslyckades: %v", err)
		return nil, err
	}
	return result, nil
}

func generateWithHuggingface(prompt CVPrompt) (map[string]interface{}, error) {
	apiKey := getAPIKey(HuggingfaceProvider)
	if apiKey == "" {
		return nil, fmt.Errorf("ingen Huggingface API-nyckel tillgänglig")
	}

	url := "https://api-inference.huggingface.co/v1/chat/completions"

	log.Printf("🔍 Skickar förfrågan till HF API för CV-generering")

	promptText := buildPrompt(prompt)

	requestBody := map[string]interface{}{
		"model": "meta-llama/Llama-3.2-3B-Instruct",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": promptText,
			},
		},
		"temperature": 0.3,
		"max_tokens": 2048,
		"top_p": 0.7,
		"stream": true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("kunde inte skapa request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HF API anrop misslyckades: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ HF API fel status %d\nHeaders: %v\nBody: %s", resp.StatusCode, resp.Header, string(body))
		return nil, fmt.Errorf("HF API returnerade status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ HF API status: %d", resp.StatusCode)

	// Hantera streaming response
	var fullResponse bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}
			var streamResponse HFStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResponse); err != nil {
				continue
			}

			if len(streamResponse.Choices) > 0 {
				content := streamResponse.Choices[0].Delta.Content
				fullResponse.WriteString(content)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Fel vid läsning av stream: %v", err)
		return nil, fmt.Errorf("fel vid läsning av stream: %v", err)
	}

	// Extrahera JSON från svaret
	responseStr := fullResponse.String()
	startIdx := strings.Index(responseStr, "{")
	endIdx := strings.LastIndex(responseStr, "}")
	
	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("kunde inte hitta JSON i svaret: %s", responseStr)
	}

	jsonStr := responseStr[startIdx : endIdx+1]
	log.Printf("📥 Extraherat JSON-svar: %s", jsonStr)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("kunde inte unmarshalla svar: %v", err)
	}

	return result, nil
}

func generateWithGemini(prompt CVPrompt) (map[string]interface{}, error) {
	cleanedJobDescription := cleanHTML(prompt.JobDescription)

	log.Printf("Data som skickas till Gemini:\n"+
		"Namn: %s\n"+
		"Telefon: %s\n"+
		"Plats: %s\n"+
		"Bio: %s\n"+
		"Färdigheter: %s\n"+
		"Erfarenhet: %s\n"+
		"Utbildning: %s\n"+
		"Certifieringar: %s\n"+
		"Jobbtitel: %s\n"+
		"Jobbeskrivning: %s\n",
		prompt.Name,
		prompt.Phone,
		prompt.Location,
		prompt.Bio,
		prompt.Skills,
		prompt.Experience,
		prompt.Education,
		prompt.Certifications,
		prompt.JobTitle,
		cleanedJobDescription)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(getAPIKey(GeminiProvider)))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash-8b")
	model.SetTemperature(0.1)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)

	session := model.StartChat()
	promptText := buildPrompt(prompt)
	
	log.Printf("\nPrompt som skickas till Gemini:\n%s\n", promptText)

	resp, err := session.SendMessage(ctx, genai.Text(promptText))
	if err != nil {
		log.Printf("Fel vid generering av innehåll med Gemini: %v", err)
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("inget svar från Gemini")
	}

	var aiResponse string
	for _, part := range resp.Candidates[0].Content.Parts {
		aiResponse += fmt.Sprintf("%v", part)
	}

	return parseAIResponse(aiResponse)
}

func cleanHTML(text string) string {
	// Ta bort HTML-taggar och konvertera till ren text
	re := regexp.MustCompile("<[^>]*>")
	text = re.ReplaceAllString(text, "")
	// Ersätt flera mellanslag/radbrytningar med ett mellanslag
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func parseAIResponse(aiResponse string) (map[string]interface{}, error) {
	// Förbättrad JSON-extrahering
	var cleanedJSON string
	if jsonStart := strings.Index(aiResponse, "{"); jsonStart != -1 {
		if jsonEnd := strings.Index(aiResponse[jsonStart:], "\n```"); jsonEnd != -1 {
			cleanedJSON = aiResponse[jsonStart:jsonStart+jsonEnd]
		} else if jsonEnd := strings.LastIndex(aiResponse[jsonStart:], "}"); jsonEnd != -1 {
			cleanedJSON = aiResponse[jsonStart : jsonStart+jsonEnd+1]
		} else {
			return nil, fmt.Errorf("kunde inte hitta slutet på JSON-objektet")
		}
	} else {
		return nil, fmt.Errorf("kunde inte hitta början på JSON-objektet")
	}

	// Rensa bort eventuella markdown-markeringar och extra text
	cleanedJSON = strings.TrimSpace(cleanedJSON)
	cleanedJSON = strings.TrimPrefix(cleanedJSON, "```json")
	cleanedJSON = strings.TrimSuffix(cleanedJSON, "```")

	// Logga den rensade JSON:en för felsökning
	log.Printf("Rensat JSON-svar:\n%s", cleanedJSON)

	var result map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(cleanedJSON))
	decoder.UseNumber() // För bättre hantering av numeriska värden
	
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("fel vid parsing av AI-svar: %v\nJSON: %s", err, cleanedJSON)
	}

	// Validera nyckelstrukturen
	if _, ok := result["personlig_info"].(map[string]interface{}); !ok {
		return nil, fmt.Errorf("ogiltig JSON-struktur: saknar personlig_info")
	}

	return addEmojisToResponse(result), nil
}

func addEmojisToResponse(response map[string]interface{}) map[string]interface{} {
	if personInfo, ok := response["personlig_info"].(map[string]interface{}); ok {
		if kontakt, ok := personInfo["kontakt"].(map[string]interface{}); ok {
			if _, ok := kontakt["email"]; ok {
				kontakt["email"] = "📧 " + kontakt["email"].(string)
			}
			if _, ok := kontakt["telefon"]; ok {
				kontakt["telefon"] = "📱 " + kontakt["telefon"].(string)
			}
			if _, ok := kontakt["adress"]; ok {
				kontakt["adress"] = "📍 " + kontakt["adress"].(string)
			}
			if _, ok := kontakt["linkedin"]; ok {
				kontakt["linkedin"] = "💼 " + kontakt["linkedin"].(string)
			}
			if _, ok := kontakt["github"]; ok {
				kontakt["github"] = "💻 " + kontakt["github"].(string)
			}
			if _, ok := kontakt["portfolio"]; ok {
				kontakt["portfolio"] = "🌐 " + kontakt["portfolio"].(string)
			}
		}
	}
	return response
}

func buildPrompt(prompt CVPrompt) string {
	return fmt.Sprintf(`Skapa ett detaljerat och personligt CV. Fyll på informationen på kreativ sätt och hitta på så att den låter realikstisk på alla fält använd dig av  på följande information:
Mitt Namn: "exemple"
Jobbtitel som jag söker till: %s
Beskrivning av önskad position: %s
mina erfarenhet: %s
mina utbildningar: %s
mina skills: %s
mina certifactions: %s
övriga informationen om mig: %s
skicka tillbaka endast med JSON-format med följande struktur och förklara inte koden eller med text.
Skicka tillbaka endast med json format. börja inte med ordent med json heller
gå rakt på saken
{
    "personlig_info": {
        "namn": "exemple",
        "titel": "%s",
        "bild": "URL till profilbild",
        "kontakt": {
            "email": "exempel@email.se",
            "telefon": "%s",
            "adress": "%s",
            "linkedin": "/in/dinlinkedin",
            "github": "/dinhub",
            "portfolio": "www.dinportfolio.se"
        }
    },
    "fardigheter": [
        "Färdighet1", "Färdighet2", "Färdighet3"
    ],
    "sprak": [
        {"sprak": "Språk1", "niva": "Nivå"},
        {"sprak": "Språk2", "niva": "Nivå"}
    ],
    "profil": "En kort professionell profiltext.",
    "arbetslivserfarenhet": [
        {
            "titel": "Jobbtitel",
            "foretag": "Företagets Namn",
            "period": "Startdatum - Slutdatum",
            "beskrivning": [
                "Ansvar eller prestation 1",
                "Ansvar eller prestation 2"
            ]
        }
    ],
    "utbildning": [
        {
            "examen": "Examenstyp",
            "skola": "Skolans Namn",
            "period": "Startår - Slutår",
            "beskrivning": "Beskrivning av utbildningen."
        }
    ],
    "projekt": [
        "Projekt1",
        "Projekt2"
    ],
    "certifieringar": [
        "Certifiering1",
        "Certifiering2"
    ]
}`,
        // Första set av parametrar för informationen
        prompt.JobTitle,
        prompt.JobDescription,
        prompt.Experience,
        prompt.Education,
        prompt.Skills,
        prompt.Certifications,
        prompt.Bio,
        // Andra set av parametrar för JSON-strukturen
        prompt.JobTitle,
        prompt.Phone,
        prompt.Location)
}
