// internal/handlers/cv.go
package handlers

import (
	"awesomeProject/internal/data"
	"awesomeProject/internal/utils"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
)

func GenerateCV(c *gin.Context) {
	var request struct {
		Name           string `json:"name"`
		JobTitle       string `json:"jobTitle"`
		JobDescription string `json:"jobDescription"`
		Experience     string `json:"experience"`
		Education      string `json:"education"`
		Skills         string `json:"skills"`
		Certifications string `json:"certifications"`
		Bio           string `json:"bio"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
		Location      string `json:"location"`
		TemplateId    string `json:"templateId"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Ogiltig förfrågan: " + err.Error()})
		return
	}

	// Skapa CVPrompt med alla fält inklusive jobTitle och jobDescription
	prompt := utils.CVPrompt{
		Name:           request.Name,
		JobTitle:       request.JobTitle,
		JobDescription: request.JobDescription,
		Experience:     request.Experience,
		Education:      request.Education,
			Skills:         request.Skills,
		Certifications: request.Certifications,
		Bio:           request.Bio,
			Email:         request.Email,
		Phone:         request.Phone,
		Location:      request.Location,
	}

	// Generera AI-innehåll
	aiResponse, err := utils.GenerateAIContent(prompt)
	if err != nil {
		log.Printf("Fel vid AI-generering: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunde inte generera CV-innehåll"})
		return
	}

	// Extrahera personlig information från AI-svaret
	personligInfoMap, _ := aiResponse["personlig_info"].(map[string]interface{})
	kontaktMap, _ := personligInfoMap["kontakt"].(map[string]interface{})

	templateData := data.TemplateData{
		PersonligInfo: data.PersonligInfo{
			Namn:    getStringValueFromMap(personligInfoMap, "namn", request.Name),
			Titel:   getStringValueFromMap(personligInfoMap, "titel", request.JobTitle),
			Bild:    getStringValueFromMap(personligInfoMap, "bild", "https://via.placeholder.com/150"),
			Kontakt: []data.KontaktItem{
				{
					Typ:   "email",
					Varde: getStringValueFromMap(kontaktMap, "email", request.Email),
					Ikon:  "📧",
				},
				{
					Typ:   "telefon",
					Varde: getStringValueFromMap(kontaktMap, "telefon", request.Phone),
					Ikon:  "📱",
				},
				{
					Typ:   "adress",
					Varde: getStringValueFromMap(kontaktMap, "adress", request.Location),
					Ikon:  "📍",
				},
				{
					Typ:   "linkedin",
					Varde: getStringValueFromMap(kontaktMap, "linkedin", "LinkedIn"),
					Ikon:  "🔗",
				},
				{
					Typ:   "github",
					Varde: getStringValueFromMap(kontaktMap, "github", "GitHub"),
					Ikon:  "💻",
				},
				{
					Typ:   "portfolio",
					Varde: getStringValueFromMap(kontaktMap, "portfolio", "Portfolio"),
					Ikon:  "🌐",
				},
			},
		},
		Fardigheter:         convertToStringSlice(aiResponse["fardigheter"]),
		Sprak:               parseAISprak(aiResponse["sprak"]),
		Profil:              getStringValue(aiResponse["profil"]),
		Arbetslivserfarenhet: parseAIExperience(aiResponse["arbetslivserfarenhet"]),
		Utbildning:          parseAIEducation(aiResponse["utbildning"]),
		Projekt:             convertToStringSlice(aiResponse["projekt"]),
		Certifieringar:      convertToStringSlice(aiResponse["certifieringar"]),
	}

	// Välj mall baserat på template-ID
	templateFile := "cv_template.html" // Default mall
	if request.TemplateId == "modern" {
		templateFile = "cv-test.html"
	} else if request.TemplateId == "creative" {
		templateFile = "cv-test2.html"
	} else if request.TemplateId == "cv3" {
		templateFile = "cv-test3.html"
	}

	// Rendera template med funcs map för eq och ge helper
	tmpl, err := template.New(templateFile).Funcs(template.FuncMap{
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"ge": func(a, b int) bool {
			return a >= b
		},
	}).ParseFiles("internal/templates/" + templateFile)

	if err != nil {
		log.Printf("Fel vid parsing av template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunde inte ladda CV-mall"})
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		log.Printf("Fel vid rendering av template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunde inte generera CV"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"html": buf.String(),
	})
}

// Hjälpfunktioner för att konvertera AI-svar
func convertToStringSlice(input interface{}) []string {
	if input == nil {
		return []string{}
	}
	
	slice, ok := input.([]interface{})
	if !ok {
		return []string{}
	}

	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

func parseAIExperience(input interface{}) []data.Arbetslivserfarenhet {
	if input == nil {
		return []data.Arbetslivserfarenhet{}
	}

	experiences, ok := input.([]interface{})
	if !ok {
		return []data.Arbetslivserfarenhet{}
	}

	var result []data.Arbetslivserfarenhet
	for _, exp := range experiences {
		expMap, ok := exp.(map[string]interface{})
		if !ok {
			continue
		}

		erfarenhet := data.Arbetslivserfarenhet{
			Titel:   getStringValueFromMap(expMap, "titel", ""),
			Foretag: getStringValueFromMap(expMap, "foretag", ""),
			Period:  getStringValueFromMap(expMap, "period", ""),
		}

		// Hantera beskrivning som kan vara antingen sträng eller array
		beskrivning := expMap["beskrivning"]
		switch v := beskrivning.(type) {
		case string:
			erfarenhet.Beskrivning = []string{v}
		case []interface{}:
			erfarenhet.Beskrivning = convertToStringSlice(v)
		}

		if erfarenhet.Titel != "" {
			result = append(result, erfarenhet)
		}
	}

	return result
}

func parseAIEducation(input interface{}) []data.Utbildning {
	if input == nil {
		return []data.Utbildning{}
	}

	eduSlice, ok := input.([]interface{})
	if !ok {
		return []data.Utbildning{}
	}

	var utbildningar []data.Utbildning
	for _, edu := range eduSlice {
		eduMap, ok := edu.(map[string]interface{})
		if !ok {
			continue
		}

		// Sätt default-värden om data saknas
		utbildning := data.Utbildning{
			Examen:      getStringValueFromMap(eduMap, "examen", "Examen saknas"),
			Skola:       getStringValueFromMap(eduMap, "universitet", "Universitet saknas"),
			Period:      getStringValueFromMap(eduMap, "period", "Period saknas"),
			Beskrivning: []string{getStringValueFromMap(eduMap, "beskrivning", "")},
		}

		// Lägg endast till om vi har meningsfull data
		if utbildning.Examen != "Examen saknas" || utbildning.Skola != "Universitet saknas" {
			utbildningar = append(utbildningar, utbildning)
		}
	}

	return utbildningar
}

func parseAISprak(input interface{}) []data.Sprak {
	if input == nil {
		return []data.Sprak{}
	}

	sprakList, ok := input.([]interface{})
	if !ok {
		return []data.Sprak{}
	}

	var result []data.Sprak
	for _, sprak := range sprakList {
		sprakMap, ok := sprak.(map[string]interface{})
		if !ok {
			continue
		}
		
		sprakItem := data.Sprak{
			Sprak: getStringValueFromMap(sprakMap, "sprak", ""),
			Niva:  getStringValueFromMap(sprakMap, "niva", ""),
		}
		
		if sprakItem.Sprak != "" && sprakItem.Niva != "" {
			result = append(result, sprakItem)
		}
	}
	return result
}

// Ny hjälpfunktion för att hantera null-värden
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

// Lägg till en ny funktion för att hämta strängvärden från map med default-värde
func getStringValueFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, exists := m[key]; exists && val != nil {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}
