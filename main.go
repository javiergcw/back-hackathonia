package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/javierg/hackathon-bqia/internal/auth"
	"github.com/javierg/hackathon-bqia/internal/domain"
	"github.com/javierg/hackathon-bqia/internal/handlers"
	"github.com/javierg/hackathon-bqia/internal/llm"
	"github.com/javierg/hackathon-bqia/internal/rag"
	"github.com/javierg/hackathon-bqia/internal/server"
	"github.com/javierg/hackathon-bqia/internal/session"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	knowledgeFolder := os.Getenv("KNOWLEDGE_FOLDER")
	if knowledgeFolder == "" {
		knowledgeFolder = "./data/knowledge"
	}
	if err := os.MkdirAll(knowledgeFolder, 0755); err != nil {
		log.Printf("warning: could not create knowledge folder: %v", err)
	}

	store := session.NewStore()
	ragClient := rag.NewRetrieve("data/knowledge.json")
	if err := ragClient.LoadProfiles("data/profiles.json"); err != nil {
		log.Printf("profiles: %v", err)
	}
	llmClient := llm.NewClient()

	users, err := auth.LoadUsers("data/users.json")
	if err != nil {
		log.Printf("users: %v (continuing without users)", err)
		users = []domain.User{}
	}
	log.Printf("loaded %d users", len(users))

	h := handlers.NewHandler(llmClient, ragClient, store, users)

	r := server.NewRouter(h)
	srv := server.NewServer(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv.Addr = ":" + port

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}