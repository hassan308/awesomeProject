// internal/data/cv_data.go
package data

type CVRequest struct {
    DisplayName    string `json:"display_name"`
    Jobbtitel      string `json:"jobbtitel"`
    JobDescription string `json:"job_description"`
    Experience     string `json:"experience"`
    Education      string `json:"education"`
    Skills         string `json:"skills"`
    Certifications string `json:"certifications"`
    Bio            string `json:"bio"`
    Email          string `json:"email"`
    Phone          string `json:"phone"`
    Location       string `json:"location"`
}

type KontaktItem struct {
    Typ   string `json:"typ"`
    Varde string `json:"varde"`
    Ikon  string `json:"ikon"`
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
    Examen      string   `json:"examen"`
    Skola       string   `json:"skola"`
    Period      string   `json:"period"`
    Beskrivning []string `json:"beskrivning"`
}

type PersonligInfo struct {
    Namn    string       `json:"namn"`
    Titel   string       `json:"titel"`
    Bild    string       `json:"bild"`
    Kontakt []KontaktItem `json:"kontakt"`
}

type CVData struct {
    PersonligInfo        PersonligInfo         `json:"personlig_info"`
    Fardigheter         []string              `json:"fardigheter"`
    Sprak               []Sprak               `json:"sprak"`
    Profil              string                `json:"profil"`
    Arbetslivserfarenhet []Arbetslivserfarenhet `json:"arbetslivserfarenhet"`
    Utbildning          []Utbildning           `json:"utbildning"`
    Projekt             []string               `json:"projekt"`
    Certifieringar      []string               `json:"certifieringar"`
}

type TemplateData struct {
    PersonligInfo         PersonligInfo           `json:"personlig_info"`
    Fardigheter          []string                `json:"fardigheter"`
    Sprak                []Sprak                 `json:"sprak"`
    Profil               string                  `json:"profil"`
    Arbetslivserfarenhet []Arbetslivserfarenhet `json:"arbetslivserfarenhet"`
    Utbildning           []Utbildning           `json:"utbildning"`
    Projekt              []string                `json:"projekt"`
    Certifieringar       []string                `json:"certifieringar"`
}
