package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	EvolutionAPIURL       string
	EvolutionAPIKey       string
	EvolutionInstance     string
	EvolutionInstanceID   string
	WhatsAppSenderNumber  string
	WhatsAppTargetNumber  string
	AppPublicURL          string
	LicenseKey            string
}

func Load(envFile string) (*Config, error) {
	if err := godotenv.Load(envFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load env file: %w", err)
	}

	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "hackathon_bqia"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		EvolutionAPIURL:      getEnv("EVOLUTION_API_URL", "http://144.91.79.105:8083"),
		EvolutionAPIKey:      getEnv("EVOLUTION_API_KEY", ""),
		EvolutionInstance:    getEnv("EVOLUTION_INSTANCE", "javierg"),
		EvolutionInstanceID:  getEnv("EVOLUTION_INSTANCE_ID", ""),
		WhatsAppSenderNumber: getEnv("WHATSAPP_SENDER_NUMBER", ""),
		WhatsAppTargetNumber: getEnv("WHATSAPP_TARGET_NUMBER", ""),
		AppPublicURL:         getEnv("APP_PUBLIC_URL", ""),
		LicenseKey:           getEnv("LICENSE_KEY", ""),
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
