package utils

import (
	"os"
	"github.com/sirupsen/logrus"
)

// Logger är en global logger-instans som kan användas i hela applikationen
var Logger = logrus.New()

func init() {
	// Konfigurera loggern
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Sätt output till stdout
	Logger.SetOutput(os.Stdout)

	// Sätt default log level
	Logger.SetLevel(logrus.InfoLevel)
}
