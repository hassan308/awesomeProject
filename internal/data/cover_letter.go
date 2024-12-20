package data

// CoverLetterData innehåller all data som behövs för personligt brev
type CoverLetterData struct {
	PersonligInfo PersonligInfo `json:"personlig_info"`
	Mottagare    Mottagare     `json:"mottagare"`
	Innehall     Innehall      `json:"innehall"`
	Datum        string        `json:"datum"`
	Jobb         Jobb          `json:"jobb"`
}

// Mottagare innehåller information om mottagaren av brevet
type Mottagare struct {
	Namn     string `json:"namn"`
	Foretag  string `json:"foretag"`
	Position string `json:"position"`
	Adress   string `json:"adress"`
	PostOrt  string `json:"postort"`
}

// Innehall innehåller brevets olika delar
type Innehall struct {
	Inledning     string `json:"inledning"`
	Huvudtext     string `json:"huvudtext"`
	Avslutning    string `json:"avslutning"`
	Halsningsfras string `json:"halsningsfras"`
}

// Jobb innehåller information om jobbet som söks
type Jobb struct {
	Titel       string `json:"titel"`
	Beskrivning string `json:"beskrivning"`
	Foretag     string `json:"foretag"`
} 