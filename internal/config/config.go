package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Config struct {
	Debug       bool   `env:"DEBUG" envDefault:"false"`
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
	// Lägg till andra konfigurationsvariabler här
}

func LoadConfig() *Config {
	// Ladda .env fil om den finns
	if err := godotenv.Load(); err != nil {
		log.Printf("Ingen .env fil hittades: %v", err)
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Kunde inte ladda konfiguration: %v", err)
	}

	return &cfg
} 