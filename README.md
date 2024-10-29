# CV-Projekt

Ett projekt för att hantera och visa CV-information.

## Projektstruktur

- `cmd/`: Innehåller huvudapplikationen
- `internal/`: Intern kod specifik för detta projekt
  - `data/`: Datastrukturer och datahantering
  - `templates/`: HTML-mallar
  - `utils/`: Hjälpfunktioner
- `pkg/`: Återanvändbar kod som kan användas av andra projekt
  - `logger/`: Loggningspaket

## Kom igång

1. Klona projektet
2. Kör `go mod tidy`
3. Starta applikationen med `go run cmd/main.go` 