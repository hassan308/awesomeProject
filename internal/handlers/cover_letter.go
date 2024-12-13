package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"github.com/gin-gonic/gin"
	"awesomeProject/internal/utils"
	"awesomeProject/internal/prompts"
	"awesomeProject/internal/data"
	"log"
)

type CoverLetterRequest struct {
	JobTitle       string                 `json:"jobTitle"`
	JobDescription string                 `json:"jobDescription"`
	TemplateStyle  string                 `json:"templateStyle"`
	Job            map[string]interface{} `json:"job"`
	Content        map[string]string      `json:"content,omitempty"`
}

type CoverLetterAIRequest struct {
	JobTitle       string                 `json:"jobTitle"`
	JobDescription string                 `json:"jobDescription"`
	Job            map[string]interface{} `json:"job"`
	TemplateStyle  string                 `json:"templateStyle"`
}

// GenerateAICoverLetter hanterar AI-generering av personligt brev innehåll
func GenerateAICoverLetter(c *gin.Context) {
	var req CoverLetterAIRequest
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Extrahera företagsnamn från job-objektet
	companyName := ""
	if company, ok := req.Job["company"].(map[string]interface{}); ok {
		if name, ok := company["name"].(string); ok {
			companyName = name
		}
	}

	// Skapa prompten med jobinformation
	userPrompt := fmt.Sprintf(prompts.CoverLetterUserPrompt,
		req.JobTitle,
		req.JobDescription,
		companyName,
	)

	// Skapa en prompt för AI
	prompt := utils.CVPrompt{
		JobTitle:       req.JobTitle,
		JobDescription: req.JobDescription,
		SystemPrompt:   prompts.CoverLetterSystemPrompt,
		UserPrompt:     userPrompt,
	}

	// Använd den gemensamma AI-tjänsten
	aiService := utils.GetAIService()
	aiResponse, err := aiService.GenerateContent(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI error: %v", err)})
		return
	}

	// Extrahera innehåll från AI-svaret och skicka som JSON
	content := map[string]string{
		"introduction": utils.GetStringValue(aiResponse["introduction"]),
		"experience":   utils.GetStringValue(aiResponse["experience"]),
		"motivation":   utils.GetStringValue(aiResponse["motivation"]),
		"closing":      utils.GetStringValue(aiResponse["closing"]),
	}

	response, err := json.Marshal(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", response)
}

func GenerateCoverLetter(c *gin.Context) {
	var request struct {
		TemplateId      string `json:"templateId"`
		JobTitle        string `json:"jobTitle"`
		JobDescription  string `json:"jobDescription"`
		CompanyName     string `json:"companyName"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Ogiltig förfrågan: " + err.Error()})
		return
	}

	// Skapa CoverLetterPrompt med jobbinformation
	prompt := utils.CoverLetterPrompt{
		Template:    request.TemplateId,
		JobTitle:    request.JobTitle,
		JobDesc:     request.JobDescription,
		CompanyName: request.CompanyName,
	}

	// Generera AI-innehåll
	aiResponse, err := utils.GeneratePersonalLetter(prompt)
	if err != nil {
		log.Printf("Fel vid AI-generering av personligt brev: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunde inte generera personligt brev"})
		return
	}

	// Skapa template data struktur
	templateData := data.CoverLetterData{
		PersonligInfo: data.PersonligInfo{
			Namn:    utils.GetStringValue(aiResponse["namn"]),
			Titel:   utils.GetStringValue(aiResponse["titel"]),
			Kontakt: []data.KontaktItem{
				{
					Typ:   "email",
					Varde: utils.GetStringValue(aiResponse["email"]),
					Ikon:  "📧",
				},
				{
					Typ:   "telefon",
					Varde: utils.GetStringValue(aiResponse["telefon"]),
					Ikon:  "📱",
				},
				{
					Typ:   "adress",
					Varde: utils.GetStringValue(aiResponse["adress"]),
					Ikon:  "📍",
				},
			},
		},
		Mottagare: data.Mottagare{
			Namn:     utils.GetStringValue(aiResponse["mottagare_namn"]),
			Foretag:  request.CompanyName,
			Position: utils.GetStringValue(aiResponse["mottagare_position"]),
		},
		Innehall: data.Innehall{
			Inledning:     utils.GetStringValue(aiResponse["inledning"]),
			Huvudtext:     utils.GetStringValue(aiResponse["huvudtext"]),
			Avslutning:    utils.GetStringValue(aiResponse["avslutning"]),
			Halsningsfras: utils.GetStringValue(aiResponse["halsningsfras"]),
		},
		Datum: utils.GetStringValue(aiResponse["datum"]),
		Jobb: data.Jobb{
			Titel:       request.JobTitle,
			Beskrivning: request.JobDescription,
			Foretag:     request.CompanyName,
		},
	}

	// Välj mall (just nu har vi bara creative)
	templateFile := "cover-letter-creative.html"

	// Rendera template
	tmpl, err := template.New(templateFile).ParseFiles("internal/templates/" + templateFile)
	if err != nil {
		log.Printf("Fel vid parsing av template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunde inte ladda personligt brev-mall"})
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		log.Printf("Fel vid rendering av template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kunde inte generera personligt brev"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"html": buf.String(),
	})
} 