package server

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/javierg/hackathon-bqia/internal/handlers"
)

func NewRouter(h *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(30 * time.Second))

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/health", h.Health)
	r.Post("/ask", h.Ask)
	r.Post("/simulate-cdt", h.SimulateCDT)
	r.Post("/recommend", h.Recommend)
	r.Post("/whatsapp/webhook", h.WhatsAppWebhook)

	r.Get("/knowledge", h.ListKnowledge)
	r.Post("/knowledge", h.AddKnowledge)
	r.Put("/knowledge/{id}", h.UpdateKnowledge)
	r.Delete("/knowledge/{id}", h.DeleteKnowledge)
	r.Post("/knowledge/reload", h.ReloadKnowledge)
	r.Post("/knowledge/upload", h.UploadKnowledge)
	r.Post("/knowledge/scan-folder", h.ScanFolderKnowledge)
	r.Get("/knowledge/status", h.GetKnowledgeStatus)
	r.Delete("/knowledge/doc/{docName}", h.DeleteKnowledgeByDoc)
	r.Post("/knowledge/clear", h.ClearAllKnowledge)

	r.Get("/scope", h.GetScope)
	r.Put("/scope", h.SetScope)

	r.Post("/auth/identify", h.Identify)
	r.Get("/analytics/morosidad", h.GetMorosidad)
	r.Get("/analytics/proyeccion", h.GetProyeccion)
	r.Get("/analytics/top-preguntas", h.GetTopPreguntas)

	return r
}

func ListenAndServe(r *chi.Mux, port string) error {
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.ListenAndServe()
	}()

	<-errChan
	return nil
}

func GracefulListenAndServe(srv *http.Server, port string) error {
	srv.Addr = ":" + port

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.ListenAndServe()
	}()

	<-errChan
	return nil
}

func NewServer(r *chi.Mux) *http.Server {
	return &http.Server{
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}