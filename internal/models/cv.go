package models

type Utbildning struct {
    Examen      string   `json:"examen"`
    Skola       string   `json:"skola"`
    Period      string   `json:"period"`
    Beskrivning []string `json:"beskrivning"`
} 