package municipality

import (
	"strings"
)

// MunicipalityMap maps municipality names to their IDs
var MunicipalityMap = map[string]string{
	"Stockholm":    "AvNB_uwa_6n6",
	"Göteborg":     "PVZL_BQT_XtL",
	"Gothenburg":   "PVZL_BQT_XtL", // Engelsk variant
	"Malmö":        "oYPt_yRA_Smm",
	"Malmo":        "oYPt_yRA_Smm", // Utan ö
	"Uppsala":      "otaF_bQY_4ZD",
	"Västerås":     "8deT_FRF_2SP",
	"Vasteras":     "8deT_FRF_2SP", // Utan ä
	"Solna":        "zHxw_uJZ_NJ8",
	"Gävle":        "qk8Y_2b6_82D",
	"Gavle":        "qk8Y_2b6_82D", // Utan ä
	// Län mappningar
	"Stockholms län":      "zupA_8Nt_xcD",
	"Uppsala län":         "zupA_8Nt_xcD",
	"Södermanlands län":   "zupA_8Nt_xcD",
	"Östergötlands län":   "zupA_8Nt_xcD",
	"Jönköpings län":      "zupA_8Nt_xcD",
	"Kronobergs län":      "zupA_8Nt_xcD",
	"Kalmar län":          "zupA_8Nt_xcD",
	"Gotlands län":        "zupA_8Nt_xcD",
	"Blekinge län":        "zupA_8Nt_xcD",
	"Skåne län":           "zupA_8Nt_xcD",
	"Hallands län":        "zupA_8Nt_xcD",
	"Västra Götalands län": "zupA_8Nt_xcD",
	"Värmlands län":       "zupA_8Nt_xcD",
	"Örebro län":          "zupA_8Nt_xcD",
	"Västmanlands län":    "zupA_8Nt_xcD",
	"Dalarnas län":        "zupA_8Nt_xcD",
	"Gävleborgs län":      "zupA_8Nt_xcD",
	"Västernorrlands län": "zupA_8Nt_xcD",
	"Jämtlands län":       "zupA_8Nt_xcD",
	"Västerbottens län":   "zupA_8Nt_xcD",
	"Norrbottens län":     "zupA_8Nt_xcD",
	// Fyll på med resten av kommunerna här
}

// AlternativeNames maps alternative spellings to the official name
var AlternativeNames = map[string]string{
	"gothenburg": "Göteborg",
	"malmo":     "Malmö",
	"vasteras":  "Västerås",
	"gavle":     "Gävle",
	// Län alternativa namn
	"stockholms":      "Stockholms län",
	"uppsala":         "Uppsala län",
	"sodermanlands":   "Södermanlands län",
	"ostergotlands":   "Östergötlands län",
	"jonkopings":      "Jönköpings län",
	"kronobergs":      "Kronobergs län",
	"kalmar":          "Kalmar län",
	"gotlands":        "Gotlands län",
	"blekinge":        "Blekinge län",
	"skane":           "Skåne län",
	"hallands":        "Hallands län",
	"vastra gotalands": "Västra Götalands län",
	"varmlands":       "Värmlands län",
	"orebro":          "Örebro län",
	"vastmanlands":    "Västmanlands län",
	"dalarnas":        "Dalarnas län",
	"gavleborgs":      "Gävleborgs län",
	"vasternorrlands": "Västernorrlands län",
	"jamtlands":       "Jämtlands län",
	"vasterbottens":   "Västerbottens län",
	"norrbottens":     "Norrbottens län",
}

// GetMunicipalityID returns the ID for a given municipality name
func GetMunicipalityID(name string) string {
	if name == "" {
		return ""
	}

	// Konvertera första bokstaven till versal, resten till gemener
	searchName := strings.ToUpper(name[:1]) + strings.ToLower(name[1:])

	// Försök först med det exakta namnet
	if id, exists := MunicipalityMap[searchName]; exists {
		return id
	}

	// Kolla om det finns en alternativ stavning
	if officialName, exists := AlternativeNames[strings.ToLower(name)]; exists {
		return MunicipalityMap[officialName]
	}

	// Om inget hittas, returnera tom sträng
	return ""
}
