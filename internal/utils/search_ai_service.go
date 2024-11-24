package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"bytes"
	"net/http"
	"unicode"
	"time"
	"io"
	"bufio"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type SearchAnalysis struct {
	Job                string `json:"job"`
	Municipality       string `json:"municipality"`
	RequiresExperience *bool  `json:"requiresExperience"`
}

// normalizeString normaliserar en sträng för jämförelse
func normalizeString(s string) string {
	// Konvertera till lowercase
	s = strings.ToLower(s)
	
	// Ersätt svenska tecken
	replacements := map[string]string{
		"å": "a",
		"ä": "a",
		"ö": "o",
		"é": "e",
		"è": "e",
		"ë": "e",
		"ü": "u",
	}
	
	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}
	
	// Behåll bara bokstäver och siffror
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

func AnalyzeSearchQuery(query string) (*SearchAnalysis, error) {
	log.Printf("\n=== Analyserar sökfråga: %s ===\n", query)

	// Försök med Hugging Face först
	result, err := tryHuggingFace(query)
	if err != nil {
		log.Printf("Hugging Face misslyckades: %v, försöker med Gemini istället", err)
		return tryGemini(query)
	}

	// Om HF returnerade ett resultat med alla fält som null, använd Gemini
	if result != nil && result.Job == "" && result.Municipality == "" && result.RequiresExperience == nil {
		log.Printf("Hugging Face returnerade null-värden, försöker med Gemini istället")
		return tryGemini(query)
	}

	// Om job är tomt men vi har andra värden, sätt det till "jobb"
	if result != nil && result.Job == "" && (result.Municipality != "" || result.RequiresExperience != nil) {
		result.Job = "jobb"
	}

	return result, nil
}

func tryHuggingFace(query string) (*SearchAnalysis, error) {
	apiKey := getAPIKey(HuggingfaceProvider)
	if apiKey == "" {
		return nil, fmt.Errorf("ingen Huggingface API-nyckel tillgänglig")
	}

	url := "https://api-inference.huggingface.co/v1/chat/completions"

	log.Printf("🔍 Skickar förfrågan till HF API med query: %s", query)

	promptText := fmt.Sprintf(`dessa är lista på städer som du ska plocka staden eller kommunen utifrån kundens fråga [
  ["Blekinge län", "Karlshamn", "Karlskrona", "Olofström", "Ronneby", "Sölvesborg", "Dalarnas län", "Avesta", "Borlänge", "Falun", "Gagnef", "Hedemora", "Leksand", "Ludvika", "Malung-Sälen", "Mora", "Orsa", "Rättvik", "Smedjebacken", "Säter"],
  ["Vansbro", "Älvdalen", "Gotlands län", "Gotland", "Gävleborgs län", "Bollnäs", "Gävle", "Hofors", "Hudiksvall", "Ljusdal", "Nordanstig", "Ockelbo", "Ovanåker", "Sandviken", "Söderhamn", "Hallands län", "Falkenberg", "Halmstad", "Hylte", "Kungsbacka"],
  ["Laholm", "Varberg", "Jämtlands län", "Berg", "Bräcke", "Härjedalen", "Krokom", "Ragunda", "Strömsund", "Åre", "Östersund", "Jönköpings län", "Aneby", "Eksjö", "Gislaved", "Gnosjö", "Habo", "Jönköping", "Mullsjö", "Nässjö"],
  ["Sävsjö", "Tranås", "Vaggeryd", "Vetlanda", "Värnamo", "Kalmar län", "Borgholm", "Emmaboda", "Hultsfred", "Högsby", "Kalmar", "Mönsterås", "Mörbylånga", "Nybro", "Oskarshamn", "Torsås", "Vimmerby", "Västervik", "Kronobergs län", "Alvesta"],
  ["Lessebo", "Ljungby", "Markaryd", "Tingsryd", "Uppvidinge", "Växjö", "Älmhult", "Norrbottens län", "Arjeplog", "Arvidsjaur", "Boden", "Gällivare", "Haparanda", "Jokkmokk", "Kalix", "Kiruna", "Luleå", "Pajala", "Piteå", "Älvsbyn"],
  ["Överkalix", "Övertorneå", "Skåne län", "Bjuv", "Bromölla", "Burlöv", "Båstad", "Eslöv", "Helsingborg", "Hässleholm", "Höganäs", "Hörby", "Höör", "Klippan", "Kristianstad", "Kävlinge", "Landskrona", "Lomma", "Lund", "Malmö"],
  ["Osby", "Perstorp", "Simrishamn", "Sjöbo", "Skurup", "Staffanstorp", "Svalöv", "Svedala", "Tomelilla", "Trelleborg", "Vellinge", "Ystad", "Ängelholm", "Åstorp", "Örkelljunga", "Östra Göinge", "Stockholms län", "Botkyrka", "Danderyd", "Ekerö"],
  ["Haninge", "Huddinge", "Järfälla", "Lidingö", "Nacka", "Norrtälje", "Nykvarn", "Nynäshamn", "Salem", "Sigtuna", "Sollentuna", "Solna", "Stockholm", "Sundbyberg", "Södertälje", "Tyresö", "Täby", "Upplands Väsby", "Upplands-Bro", "Vallentuna"],
  ["Vaxholm", "Värmdö", "Österåker", "Södermanlands län", "Eskilstuna", "Flen", "Gnesta", "Katrineholm", "Nyköping", "Oxelösund", "Strängnäs", "Trosa", "Vingåker", "Uppsala län", "Enköping", "Heby", "Håbo", "Knivsta", "Tierp", "Uppsala"],
  ["Älvkarleby", "Östhammar", "Värmlands län", "Arvika", "Eda", "Filipstad", "Forshaga", "Grums", "Hagfors", "Hammarö", "Karlstad", "Kil", "Kristinehamn", "Munkfors", "Storfors", "Sunne", "Säffle", "Torsby", "Årjäng", "Västerbottens län"],
  ["Bjurholm", "Dorotea", "Lycksele", "Malå", "Nordmaling", "Norsjö", "Robertsfors", "Skellefteå", "Sorsele", "Storuman", "Umeå", "Vilhelmina", "Vindeln", "Vännäs", "Åsele", "Västernorrlands län", "Härnösand", "Kramfors", "Sollefteå", "Sundsvall"],
  ["Timrå", "Ånge", "Örnsköldsvik", "Västmanlands län", "Arboga", "Fagersta", "Hallstahammar", "Kungsör", "Köping", "Norberg", "Sala", "Skinnskatteberg", "Surahammar", "Västerås", "Västra Götalands län", "Ale", "Alingsås", "Bengtsfors", "Bollebygd"],
  ["Borås", "Dals-Ed", "Essunga", "Falköping", "Färgelanda", "Grästorp", "Gullspång", "Göteborg", "Götene", "Herrljunga", "Hjo", "Härryda", "Karlsborg", "Kungälv", "Lerum", "Lidköping", "Lilla Edet", "Lysekil", "Mariestad", "Mark"],
  ["Mellerud", "Munkedal", "Mölndal", "Orust", "Partille", "Skara", "Skövde", "Sotenäs", "Stenungsund", "Strömstad", "Svenljunga", "Tanum", "Tibro", "Tidaholm", "Tjörn", "Tranemo", "Trollhättan", "Töreboda", "Uddevalla", "Ulricehamn"],
  ["Vara", "Vänersborg", "Vårgårda", "Åmål", "Öckerö", "Örebro län", "Askersund", "Degerfors", "Hallsberg", "Hällefors", "Karlskoga", "Kumla", "Laxå", "Lekeberg", "Lindesberg", "Ljusnarsberg", "Nora", "Örebro", "Östergötlands län", "Boxholm"],
  ["Finspång", "Kinda", "Linköping", "Mjölby", "Motala", "Norrköping", "Söderköping", "Vadstena", "Valdemarsvik", "Ydre", "Åtvidaberg", "Ödeshög"]
]

VIKTIGT: Om användaren nämner ett län (t.ex. "gävleborg", "gävleborgs län"), returnera ALLTID länets fullständiga namn (t.ex. "Gävleborgs län") i municipality-fältet, inte en stad i länet.

Analysera följande jobbsökningsfråga och extrahera information.
Om personen specifikt nämner att de söker jobb utan erfarenhetskrav eller entry-level/junior-positioner, sätt requiresExperience till false.
Om personen specifikt söker senior-positioner eller jobb som kräver erfarenhet, sätt requiresExperience till true.
Om personen inte nämner något om erfarenhet, sätt requiresExperience till null.

Returnera ENDAST ett JSON-objekt med följande struktur:
{
    "job": "extraherad jobbtitel",
    "municipality": "extraherad kommun/län (använd exakt namn från listan)",
    "requiresExperience": false/true/null (baserat på erfarenhetskrav)
}

Exempel:
- Om användaren skriver "jobb i gävleborg" -> municipality: "Gävleborgs län"
- Om användaren skriver "jobb i gävle" -> municipality: "Gävle"

Sökfråga: %s`, query)

	// Skapa request body
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

	var result *SearchAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("kunde inte unmarshalla svar: %v", err)
	}

	return result, nil
}

func tryGemini(query string) (*SearchAnalysis, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(getAPIKey(GeminiProvider)))
	if err != nil {
		return nil, fmt.Errorf("kunde inte skapa Gemini-klient: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash-8b")
	model.SetTemperature(0.2)
	model.SetTopK(40)
	model.SetTopP(0.95)

	promptText := fmt.Sprintf(`dessa är lista på städer som du ska plocka staden eller kommunen utifrån kundens fråga [
  ["Blekinge län", "Karlshamn", "Karlskrona", "Olofström", "Ronneby", "Sölvesborg", "Dalarnas län", "Avesta", "Borlänge", "Falun", "Gagnef", "Hedemora", "Leksand", "Ludvika", "Malung-Sälen", "Mora", "Orsa", "Rättvik", "Smedjebacken", "Säter"],
  ["Vansbro", "Älvdalen", "Gotlands län", "Gotland", "Gävleborgs län", "Bollnäs", "Gävle", "Hofors", "Hudiksvall", "Ljusdal", "Nordanstig", "Ockelbo", "Ovanåker", "Sandviken", "Söderhamn", "Hallands län", "Falkenberg", "Halmstad", "Hylte", "Kungsbacka"],
  ["Laholm", "Varberg", "Jämtlands län", "Berg", "Bräcke", "Härjedalen", "Krokom", "Ragunda", "Strömsund", "Åre", "Östersund", "Jönköpings län", "Aneby", "Eksjö", "Gislaved", "Gnosjö", "Habo", "Jönköping", "Mullsjö", "Nässjö"],
  ["Sävsjö", "Tranås", "Vaggeryd", "Vetlanda", "Värnamo", "Kalmar län", "Borgholm", "Emmaboda", "Hultsfred", "Högsby", "Kalmar", "Mönsterås", "Mörbylånga", "Nybro", "Oskarshamn", "Torsås", "Vimmerby", "Västervik", "Kronobergs län", "Alvesta"],
  ["Lessebo", "Ljungby", "Markaryd", "Tingsryd", "Uppvidinge", "Växjö", "Älmhult", "Norrbottens län", "Arjeplog", "Arvidsjaur", "Boden", "Gällivare", "Haparanda", "Jokkmokk", "Kalix", "Kiruna", "Luleå", "Pajala", "Piteå", "Älvsbyn"],
  ["Överkalix", "Övertorneå", "Skåne län", "Bjuv", "Bromölla", "Burlöv", "Båstad", "Eslöv", "Helsingborg", "Hässleholm", "Höganäs", "Hörby", "Höör", "Klippan", "Kristianstad", "Kävlinge", "Landskrona", "Lomma", "Lund", "Malmö"],
  ["Osby", "Perstorp", "Simrishamn", "Sjöbo", "Skurup", "Staffanstorp", "Svalöv", "Svedala", "Tomelilla", "Trelleborg", "Vellinge", "Ystad", "Ängelholm", "Åstorp", "Örkelljunga", "Östra Göinge", "Stockholms län", "Botkyrka", "Danderyd", "Ekerö"],
  ["Haninge", "Huddinge", "Järfälla", "Lidingö", "Nacka", "Norrtälje", "Nykvarn", "Nynäshamn", "Salem", "Sigtuna", "Sollentuna", "Solna", "Stockholm", "Sundbyberg", "Södertälje", "Tyresö", "Täby", "Upplands Väsby", "Upplands-Bro", "Vallentuna"],
  ["Vaxholm", "Värmdö", "Österåker", "Södermanlands län", "Eskilstuna", "Flen", "Gnesta", "Katrineholm", "Nyköping", "Oxelösund", "Strängnäs", "Trosa", "Vingåker", "Uppsala län", "Enköping", "Heby", "Håbo", "Knivsta", "Tierp", "Uppsala"],
  ["Älvkarleby", "Östhammar", "Värmlands län", "Arvika", "Eda", "Filipstad", "Forshaga", "Grums", "Hagfors", "Hammarö", "Karlstad", "Kil", "Kristinehamn", "Munkfors", "Storfors", "Sunne", "Säffle", "Torsby", "Årjäng", "Västerbottens län"],
  ["Bjurholm", "Dorotea", "Lycksele", "Malå", "Nordmaling", "Norsjö", "Robertsfors", "Skellefteå", "Sorsele", "Storuman", "Umeå", "Vilhelmina", "Vindeln", "Vännäs", "Åsele", "Västernorrlands län", "Härnösand", "Kramfors", "Sollefteå", "Sundsvall"],
  ["Timrå", "Ånge", "Örnsköldsvik", "Västmanlands län", "Arboga", "Fagersta", "Hallstahammar", "Kungsör", "Köping", "Norberg", "Sala", "Skinnskatteberg", "Surahammar", "Västerås", "Västra Götalands län", "Ale", "Alingsås", "Bengtsfors", "Bollebygd"],
  ["Borås", "Dals-Ed", "Essunga", "Falköping", "Färgelanda", "Grästorp", "Gullspång", "Göteborg", "Götene", "Herrljunga", "Hjo", "Härryda", "Karlsborg", "Kungälv", "Lerum", "Lidköping", "Lilla Edet", "Lysekil", "Mariestad", "Mark"],
  ["Mellerud", "Munkedal", "Mölndal", "Orust", "Partille", "Skara", "Skövde", "Sotenäs", "Stenungsund", "Strömstad", "Svenljunga", "Tanum", "Tibro", "Tidaholm", "Tjörn", "Tranemo", "Trollhättan", "Töreboda", "Uddevalla", "Ulricehamn"],
  ["Vara", "Vänersborg", "Vårgårda", "Åmål", "Öckerö", "Örebro län", "Askersund", "Degerfors", "Hallsberg", "Hällefors", "Karlskoga", "Kumla", "Laxå", "Lekeberg", "Lindesberg", "Ljusnarsberg", "Nora", "Örebro", "Östergötlands län", "Boxholm"],
  ["Finspång", "Kinda", "Linköping", "Mjölby", "Motala", "Norrköping", "Söderköping", "Vadstena", "Valdemarsvik", "Ydre", "Åtvidaberg", "Ödeshög"]
]

VIKTIGT: Om användaren nämner ett län (t.ex. "gävleborg", "gävleborgs län"), returnera ALLTID länets fullständiga namn (t.ex. "Gävleborgs län") i municipality-fältet, inte en stad i länet.
Försök att översätta till svenska språk från kundens fråga från stad till yrke. Alltid på svenska.
Analysera följande jobbsökningsfråga och extrahera information.
Om personen specifikt nämner att de söker jobb utan erfarenhetskrav eller entry-level/junior-positioner, sätt requiresExperience till false.
Om personen specifikt söker senior-positioner eller jobb som kräver erfarenhet, sätt requiresExperience till true.
Om personen inte nämner något om erfarenhet, sätt requiresExperience till null.

Returnera ENDAST ett JSON-objekt med följande struktur:
{
    "job": "extraherad jobbtitel",
    "municipality": "extraherad kommun/län (använd exakt namn från listan)",
    "requiresExperience": false/true/null (baserat på erfarenhetskrav)
}

Exempel:
- Om användaren skriver "jobb i gävleborg" -> municipality: "Gävleborgs län"
- Om användaren skriver "jobb i gävle" -> municipality: "Gävle"

Sökfråga: %s`, query)

	log.Printf("\nSkickar prompt till Gemini:\n%s\n", promptText)

	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		return nil, fmt.Errorf("Gemini generering misslyckades: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("inget svar från Gemini")
	}

	var aiResponse string
	for _, part := range resp.Candidates[0].Content.Parts {
		aiResponse += fmt.Sprintf("%v", part)
	}

	log.Printf("\nGemini-svar:\n%s\n", aiResponse)

	// Extrahera JSON från svaret
	var cleanedJSON string
	if jsonStart := strings.Index(aiResponse, "{"); jsonStart != -1 {
		if jsonEnd := strings.LastIndex(aiResponse, "}"); jsonEnd != -1 {
			cleanedJSON = aiResponse[jsonStart : jsonEnd+1]
		}
	}

	if cleanedJSON == "" {
		return nil, fmt.Errorf("kunde inte extrahera JSON från Gemini-svar")
	}

	var result SearchAnalysis
	if err := json.Unmarshal([]byte(cleanedJSON), &result); err != nil {
		return nil, fmt.Errorf("kunde inte parsa Gemini-svar: %v", err)
	}

	if result.Job == "" {
		return nil, fmt.Errorf("kunde inte identifiera jobb från Gemini-svar")
	}

	log.Printf("\nGemini extraherad information:\nJobb: %s\nKommun: %s\nErfarenhetskrav: %v\n", 
		result.Job, result.Municipality, result.RequiresExperience)

	return &result, nil
}
