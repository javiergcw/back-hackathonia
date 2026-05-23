package main

import (
	"log"
	"os"

	"github.com/javierg/hackathon-bqia/internal/infrastructure/config"
	"github.com/javierg/hackathon-bqia/internal/infrastructure/database"
	httpserver "github.com/javierg/hackathon-bqia/internal/infrastructure/http"
)

func main() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env.dev"
	}

	cfg, err := config.Load(envFile)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	server := httpserver.NewServer(cfg, db)
	log.Printf("server listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
