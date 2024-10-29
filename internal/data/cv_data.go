// internal/data/cv_data.go
package data

type Kontakt struct {
	Email     string `json:"email"`
	Telefon   string `json:"telefon"`
	Adress    string `json:"adress"`
	LinkedIn  string `json:"linkedin"`
	GitHub    string `json:"github"`
	Portfolio string `json:"portfolio"`
}

type Sprak struct {
	Sprak string `json:"sprak"`
	Niva  string `json:"niva"`
}

type Arbetslivserfarenhet struct {
	Titel       string   `json:"titel"`
	Foretag     string   `json:"foretag"`
	Period      string   `json:"period"`
	Beskrivning []string `json:"beskrivning"`
}

type Utbildning struct {
	Examen      string `json:"examen"`
	Skola       string `json:"skola"`
	Period      string `json:"period"`
	Beskrivning string `json:"beskrivning"`
}

type CVData struct {
	PersonligInfo struct {
		Namn    string            `json:"namn"`
		Titel   string            `json:"titel"`
		Bild    string            `json:"bild"`
		Kontakt map[string]string `json:"kontakt"`
	} `json:"personlig_info"`
	Fardigheter          []string               `json:"fardigheter"`
	Sprak                []Sprak                `json:"sprak"`
	Profil               string                 `json:"profil"`
	Arbetslivserfarenhet []Arbetslivserfarenhet `json:"arbetslivserfarenhet"`
	Utbildning           []Utbildning           `json:"utbildning"`
	Projekt              []string               `json:"projekt"`
	Certifieringar       []string               `json:"certifieringar"`
}
