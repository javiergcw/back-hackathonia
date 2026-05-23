package whatsapp

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/javierg/hackathon-bqia/internal/infrastructure/config"
	"github.com/javierg/hackathon-bqia/internal/infrastructure/http/middleware"
)

func RegisterRoutes(r *mux.Router, cfg *config.Config) {
	h := NewHandler(cfg)

	r.HandleFunc("/whatsapp/status", h.Status).Methods("GET")
	r.Handle("/whatsapp/send", middleware.LicenseKey(cfg)(http.HandlerFunc(h.Send))).Methods("POST")
	r.HandleFunc("/whatsapp/webhook/register", h.RegisterWebhook).Methods("POST")
	r.HandleFunc("/webhooks/evolution", h.Webhook).Methods("POST")
}
