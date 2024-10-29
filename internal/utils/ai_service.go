package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const GEMINI_API_KEY = "AIzaSyAFl2ldHhMnRRhSqud71S8lLe9QKes1KtA"

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

func GenerateAIContent(prompt CVPrompt) (map[string]interface{}, error) {
	// Logga inkommande data
	log.Printf("\nData som skickas till AI:\n"+
		"Namn: %s\n"+
		"Email: %s\n"+
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
		prompt.Email,
		prompt.Phone,
		prompt.Location,
		prompt.Bio,
		prompt.Skills,
		prompt.Experience,
		prompt.Education,
		prompt.Certifications,
		prompt.JobTitle,
		prompt.JobDescription)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(GEMINI_API_KEY))
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

	promptText := fmt.Sprintf(`Skapa ett detaljerat och professionellt CV  fyll på informationen på kreativ sätt på följande information:
Mitt Namn: %s
Jobbtitel som jag söker till: %s
Beskrivning av önskad position: %s
mina erfarenhet: %s
mina utbildningar: %s
mina skills: %s
mina certifactions: %s
övriga informationen om mig: %s

skicka tillbaka endast med JSON-format med följande struktur och förklara inte koden eller med text. Skicka tillbaka endast med json format. börja inte med ordent med json heller
gå rakt på saken
{
    "personlig_info": {
        "namn": "Fullständigt Namn",
        "titel": "Professionell Titel",
        "bild": "URL till profilbild",
        "kontakt": {
            "email": "%s",
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
		prompt.Name, prompt.JobTitle, prompt.JobDescription, prompt.Experience,
		prompt.Education, prompt.Skills, prompt.Certifications, prompt.Bio,
		prompt.Email, prompt.Phone, prompt.Location)

	// Logga prompten som skickas till AI
	log.Printf("\nPrompt som skickas till AI:\n%s\n", promptText)

	resp, err := session.SendMessage(ctx, genai.Text(promptText))
	if err != nil {
		log.Printf("Fel vid generering av innehåll: %v", err)
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("inget svar från AI")
	}

	var aiResponse string
	for _, part := range resp.Candidates[0].Content.Parts {
		aiResponse += fmt.Sprintf("%v", part)
	}

	cleanedResponse := cleanJSONResponse(aiResponse)
	log.Printf("\nRensat AI-svar:\n%s\n", cleanedResponse)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedResponse), &result); err != nil {
		log.Printf("JSON parsing error. Original svar: %s\nRensat svar: %s", aiResponse, cleanedResponse)
		return nil, fmt.Errorf("fel vid parsing av AI-svar: %v", err)
	}

	return result, nil
}

func cleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)
	response = regexp.MustCompile("```json\n|```\n|```").ReplaceAllString(response, "")

	startBrace := strings.Index(response, "{")
	endBrace := strings.LastIndex(response, "}")

	if startBrace >= 0 && endBrace >= 0 && endBrace > startBrace {
		response = response[startBrace : endBrace+1]
	}

	return strings.TrimSpace(response)
}
