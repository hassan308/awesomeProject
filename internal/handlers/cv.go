// internal/handlers/cv.go
package handlers

import (
	"awesomeProject/internal/data"
	"awesomeProject/internal/utils"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type CVRequest struct {
	DisplayName    string `json:"displayName"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Location       string `json:"location"`
	Bio            string `json:"bio"`
	Skills         string `json:"skills"`
	Experience     string `json:"experience"`
	Education      string `json:"education"`
	Certifications string `json:"certifications"`
	Jobbtitel      string `json:"Jobbtitel"`
	JobDescription string `json:"jobDescription"`
}

func GenerateCV(c *gin.Context) {
	var request CVRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Fel vid JSON-bindning: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ogiltig förfrågan: " + err.Error()})
		return
	}

	// Skapa en prompt för AI
	aiPrompt := utils.CVPrompt{
		Name:           request.DisplayName,
		JobTitle:       request.Jobbtitel,
		JobDescription: request.JobDescription,
		Experience:     request.Experience,
		Education:      request.Education,
		Skills:         request.Skills,
		Certifications: request.Certifications,
		Bio:            request.Bio,
		Email:          request.Email,
		Phone:          request.Phone,
		Location:       request.Location,
	}

	// Generera AI-innehåll
	aiResponse, err := utils.GenerateAIContent(aiPrompt)
	if err != nil {
		log.Printf("Fel vid AI-generering: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunde inte generera CV-innehåll"})
		return
	}

	// Säkerställ att personlig info finns och har rätt struktur
	personligInfo := aiResponse["personlig_info"].(map[string]interface{})
	kontakt := personligInfo["kontakt"].(map[string]interface{})

	// Uppdatera kontaktinformation
	kontakt["email"] = request.Email
	kontakt["telefon"] = request.Phone
	kontakt["adress"] = request.Location
	kontakt["linkedin"] = "/in/" + strings.ToLower(strings.Replace(request.DisplayName, " ", "-", -1))
	kontakt["github"] = "/github"
	kontakt["portfolio"] = "www.portfolio.se"

	// Hantera arbetslivserfarenhet
	var arbetslivserfarenhet []map[string]interface{}
	if aiExp, ok := aiResponse["arbetslivserfarenhet"].([]interface{}); ok && len(aiExp) > 0 {
		for _, exp := range aiExp {
			if expMap, ok := exp.(map[string]interface{}); ok {
				arbetslivserfarenhet = append(arbetslivserfarenhet, map[string]interface{}{
					"Titel":       expMap["titel"],
					"Foretag":     expMap["foretag"],
					"Period":      expMap["period"],
					"Beskrivning": expMap["beskrivning"],
				})
			}
		}
	} else {
		// Om ingen erfarenhet finns, lägg till en default-post
		arbetslivserfarenhet = []map[string]interface{}{
			{
				"Titel":       "Legitimerad Sjuksköterska",
				"Foretag":     "Söker nya möjligheter",
				"Period":      "Tillgänglig omgående",
				"Beskrivning": []string{"Redo för nya utmaningar inom kommunal vård"},
			},
		}
	}

	// Hantera utbildning
	var utbildning []map[string]interface{}
	if aiEdu, ok := aiResponse["utbildning"].([]interface{}); ok && len(aiEdu) > 0 {
		for _, edu := range aiEdu {
			if eduMap, ok := edu.(map[string]interface{}); ok {
				utbildning = append(utbildning, map[string]interface{}{
					"Examen":      eduMap["examen"],
					"Skola":       eduMap["skola"],
					"Period":      eduMap["period"],
					"Beskrivning": eduMap["beskrivning"],
				})
			}
		}
	} else {
		// Om ingen utbildning finns, lägg till en default-post
		utbildning = []map[string]interface{}{
			{
				"Examen":      "Legitimerad Sjuksköterska",
				"Skola":       "Vårdutbildning",
				"Period":      "Fullgjord utbildning",
				"Beskrivning": "Legitimerad sjuksköterska med inriktning mot kommunal vård",
			},
		}
	}

	// Hantera språk
	var sprak []map[string]interface{}
	if aiSprak, ok := aiResponse["sprak"].([]interface{}); ok && len(aiSprak) > 0 {
		for _, s := range aiSprak {
			if sprakMap, ok := s.(map[string]interface{}); ok {
				sprak = append(sprak, map[string]interface{}{
					"Sprak": sprakMap["sprak"],
					"Niva":  sprakMap["niva"],
				})
			}
		}
	} else {
		// Default språkkunskaper
		sprak = []map[string]interface{}{
			{
				"Sprak": "Svenska",
				"Niva":  "Modersmål",
			},
			{
				"Sprak": "Engelska",
				"Niva":  "Grundläggande",
			},
		}
	}

	// Konvertera svaret till rätt format för templaten
	templateData := map[string]interface{}{
		"PersonligInfo": map[string]interface{}{
			"Namn":    personligInfo["namn"],
			"Titel":   personligInfo["titel"],
			"Bild":    "https://via.placeholder.com/150",
			"Kontakt": kontakt,
		},
		"Fardigheter":          aiResponse["fardigheter"],
		"Sprak":                sprak,
		"Profil":               aiResponse["profil"],
		"Arbetslivserfarenhet": arbetslivserfarenhet,
		"Utbildning":           utbildning,
		"Projekt": []string{
			"Kommunal vård och omsorg",
			"Patientvård och dokumentation",
		},
		"Certifieringar": []string{
			"Legitimerad Sjuksköterska",
		},
	}

	// Logga data som skickas till template
	prettyJSON, _ := json.MarshalIndent(templateData, "", "  ")
	log.Printf("Data som skickas till template:\n%s\n", string(prettyJSON))

	// Rendera template
	tmpl, err := template.ParseFiles("internal/templates/cv_template.html")
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

	result := make([]data.Arbetslivserfarenhet, len(experiences))
	for i, exp := range experiences {
		expMap := exp.(map[string]interface{})
		result[i] = data.Arbetslivserfarenhet{
			Titel:       expMap["titel"].(string),
			Foretag:     expMap["foretag"].(string),
			Period:      expMap["period"].(string),
			Beskrivning: convertToStringSlice(expMap["beskrivning"]),
		}
	}
	return result
}

func parseAIEducation(input interface{}) []data.Utbildning {
	if input == nil {
		return []data.Utbildning{}
	}

	education, ok := input.([]interface{})
	if !ok {
		return []data.Utbildning{}
	}

	result := make([]data.Utbildning, len(education))
	for i, edu := range education {
		eduMap := edu.(map[string]interface{})
		result[i] = data.Utbildning{
			Examen:      eduMap["examen"].(string),
			Skola:       eduMap["skola"].(string),
			Period:      eduMap["period"].(string),
			Beskrivning: eduMap["beskrivning"].(string),
		}
	}
	return result
}
