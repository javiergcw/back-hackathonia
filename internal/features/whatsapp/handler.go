package whatsapp

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/javierg/hackathon-bqia/internal/infrastructure/config"
	"github.com/javierg/hackathon-bqia/internal/infrastructure/evolution"
	"github.com/javierg/hackathon-bqia/internal/shared/response"
	"github.com/javierg/hackathon-bqia/internal/shared/validation"
)

type Handler struct {
	cfg    *config.Config
	client *evolution.Client
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg:    cfg,
		client: evolution.NewClient(cfg),
	}
}

type sendMessageRequest struct {
	Number string `json:"number"` // destinatario: código país + número, ej. 573024158002
	Text   string `json:"text"`
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validation.Required(req.Number, "number"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validation.Required(req.Text, "text"); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	number := evolution.NormalizeNumber(req.Number)
	result, err := h.client.SendText(r.Context(), number, strings.TrimSpace(req.Text))
	if err != nil {
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}

	var parsed any
	if err := json.Unmarshal(result, &parsed); err != nil {
		parsed = string(result)
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"sent":   true,
		"to":     number,
		"result": parsed,
	})
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	log.Printf("whatsapp webhook event=%s instance=%s", payload.Event, payload.Instance)

	number, incomingText, ok := ExtractIncomingText(payload)
	if !ok {
		response.JSON(w, http.StatusOK, map[string]any{
			"received": true,
			"action":   "ignored",
		})
		return
	}

	target := h.cfg.WhatsAppTargetNumber
	if target == "" {
		target = number
	}

	reply := "Hola! Recibimos tu mensaje en Hackathon BQIA: \"" + incomingText + "\""
	result, err := h.client.SendText(r.Context(), target, reply)
	if err != nil {
		log.Printf("whatsapp webhook send error: %v", err)
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}

	var parsed any
	if err := json.Unmarshal(result, &parsed); err != nil {
		parsed = string(result)
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"received":      true,
		"from":          number,
		"sent_to":       evolution.NormalizeNumber(target),
		"incoming_text": incomingText,
		"result":        parsed,
	})
}

func (h *Handler) RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	webhookURL := h.cfg.AppPublicURL + "/api/v1/webhooks/evolution"
	if h.cfg.AppPublicURL == "" {
		response.Error(w, http.StatusBadRequest, "APP_PUBLIC_URL no configurada")
		return
	}

	result, err := h.client.SetWebhook(r.Context(), webhookURL)
	if err != nil {
		response.Error(w, http.StatusBadGateway, err.Error())
		return
	}

	var parsed any
	if err := json.Unmarshal(result, &parsed); err != nil {
		parsed = string(result)
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"registered": true,
		"url":        webhookURL,
		"result":     parsed,
	})
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"evolution_url":        h.cfg.EvolutionAPIURL,
		"instance":             h.cfg.EvolutionInstance,
		"instance_id":          h.cfg.EvolutionInstanceID,
		"sender_number":        h.cfg.WhatsAppSenderNumber,
		"api_key_configured":   h.cfg.EvolutionAPIKey != "",
		"target_number":        h.cfg.WhatsAppTargetNumber,
		"webhook_url":          h.cfg.AppPublicURL + "/api/v1/webhooks/evolution",
		"app_public_url_set":   h.cfg.AppPublicURL != "",
	})
}
