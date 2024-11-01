package logger

import (
	"log"
	"os"
)

// Logger är en enkel logger-struktur
type Logger struct {
	logger *log.Logger
}

// New skapar en ny logger-instans
func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "[APP] ", log.LstdFlags),
	}
}

// Info loggar informationsmeddelanden
func (l *Logger) Info(msg string) {
	l.logger.Printf("INFO: %s", msg)
} 