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
	r.Use(middleware.Timeout(50 * time.Second))

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
	r.Head("/health", h.Health)
	r.Post("/ask", h.Ask)
	r.Post("/simulate-cdt", h.SimulateCDT)
	r.Post("/recommend", h.Recommend)
	r.Get("/whatsapp/webhook", h.WhatsAppWebhook)
	r.Post("/whatsapp/webhook", h.WhatsAppWebhook)
	r.Get("/whatsapp/webhook/{number}", h.WhatsAppWebhook)
	r.Post("/whatsapp/webhook/{number}", h.WhatsAppWebhook)

	return r
}

func ListenAndServe(r *chi.Mux, port string) error {
	return http.ListenAndServe(":"+port, r)
}