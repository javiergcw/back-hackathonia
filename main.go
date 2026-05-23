package main

import (
	"log"
	"os"

	"github.com/javierg/hackathon-bqia/internal/handlers"
	"github.com/javierg/hackathon-bqia/internal/llm"
	"github.com/javierg/hackathon-bqia/internal/rag"
	"github.com/javierg/hackathon-bqia/internal/session"
	"github.com/javierg/hackathon-bqia/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	store := session.NewStore()
	ragClient := rag.NewRetrieve("data/knowledge.json")
	llmClient := llm.NewClient()

	h := handlers.NewHandler(llmClient, ragClient, store)

	r := server.NewRouter(h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := server.ListenAndServe(r, port); err != nil {
		log.Fatalf("server: %v", err)
	}
}