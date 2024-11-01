package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func getAPIKey() string {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		key = "AIzaSyAuEqi2qBJ-_4QjtdDDgU0Vwd0G05zWZVs"
		log.Println("Varning: Använder default API-nyckel. Sätt GEMINI_API_KEY miljövariabel i produktion.")
	}
	return key
}

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

func addEmojisToResponse(result map[string]interface{}) map[string]interface{} {
	if personligInfo, ok := result["personlig_info"].(map[string]interface{}); ok {
		if kontakt, ok := personligInfo["kontakt"].(map[string]interface{}); ok {
			personligInfo["kontakt"] = kontakt
		}
		result["personlig_info"] = personligInfo
	}
	return result
}

func cleanHTML(text string) string {
	// Ta bort HTML-taggar och konvertera till ren text
	re := regexp.MustCompile("<[^>]*>")
	text = re.ReplaceAllString(text, "")
	// Ersätt flera mellanslag/radbrytningar med ett mellanslag
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func GenerateAIContent(prompt CVPrompt) (map[string]interface{}, error) {
	// Rensa HTML från jobbeskrivningen
	cleanedJobDescription := cleanHTML(prompt.JobDescription)

	log.Printf("Data som skickas till AI:\n"+
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
		cleanedJobDescription)

	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(getAPIKey()))
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

	return result, nil
}

// Uppdatera prompten för att få ett mer strukturerat svar
func buildPrompt(prompt CVPrompt) string {
	return fmt.Sprintf(`Skapa ett detaljerat och personligt CV. Fyll på informationen på kreativ sätt och hitta på så att den låter realikstisk på alla fält använd dig av  på följande information:
Mitt Namn: %s
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
        "namn": "%s",
        "titel": "%s",
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
        // Första set av parametrar för informationen
        prompt.Name,
        prompt.JobTitle,
        prompt.JobDescription,
        prompt.Experience,
        prompt.Education,
        prompt.Skills,
        prompt.Certifications,
        prompt.Bio,
        // Andra set av parametrar för JSON-strukturen
        prompt.Name,
        prompt.JobTitle,
        prompt.Email,
        prompt.Phone,
        prompt.Location)
}
