package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/javierg/hackathon-bqia/internal/domain"
	"github.com/javierg/hackathon-bqia/internal/llm"
	"github.com/javierg/hackathon-bqia/internal/rag"
	"github.com/javierg/hackathon-bqia/internal/session"
)

type Handler struct {
	llmClient *llm.Client
	ragClient *rag.Client
	store     *session.Store
}

func NewHandler(llmClient *llm.Client, ragClient *rag.Client, store *session.Store) *Handler {
	return &Handler{
		llmClient: llmClient,
		ragClient: ragClient,
		store:     store,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.ok(w, map[string]string{"status": "ok"})
}

type AskRequest struct {
	Question  string `json:"question"`
	Channel   string `json:"channel"`
	SessionID string `json:"sessionId"`
}

func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
	var req AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	if req.Question == "" {
		h.error(w, http.StatusBadRequest, "MISSING_QUESTION", "question es requerido")
		return
	}

	channel := req.Channel
	if channel == "" {
		channel = "cliente"
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	chunks := h.ragClient.Retrieve(req.Question, 3)

	var answer string
	var err error

	answer, err = h.llmClient.Generate(r.Context(), req.Question, channel, chunks)
	if err != nil {
		answer = buildFallbackAnswer(chunks, channel)
	}

	h.store.AddMessage(sessionID, "user", req.Question)
	h.store.AddMessage(sessionID, "assistant", answer)

	citations := extractCitations(chunks)

	h.ok(w, map[string]interface{}{
		"answer":     answer,
		"citations":  citations,
		"grounded":   len(citations) > 0,
		"sessionId":  sessionID,
	})
}

func (h *Handler) SimulateCDT(w http.ResponseWriter, r *http.Request) {
	var req domain.SimulateCDTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	if req.Monto < 500000 {
		h.error(w, http.StatusBadRequest, "INVALID_MONTO", "monto mínimo $500.000")
		return
	}
	if req.PlazoDias <= 0 {
		h.error(w, http.StatusBadRequest, "INVALID_PLAZO", "plazo debe ser mayor a 0")
		return
	}

	tasaEA := lookupTasa(req.PlazoDias, req.Monto)
	interestBruto := calcularInteres(req.Monto, tasaEA, req.PlazoDias)
	retencion := int(math.Round(float64(interestBruto) * 0.07))
	interesNeto := interestBruto - retencion
	total := req.Monto + interesNeto

	resp := domain.SimulateCDTResponse{
		Monto:            req.Monto,
		PlazoDias:        req.PlazoDias,
		TasaEA:           tasaEA,
		InteresBruto:     interestBruto,
		RetencionFuente:  retencion,
		InteresNeto:      interesNeto,
		TotalVencimiento: total,
		Citation: domain.Citation{
			Doc:     "03_cdt_caracteristicas_beneficios.pdf",
			Seccion: "3. Tasas / 8. Simulador",
		},
	}

	h.ok(w, resp)
}

func lookupTasa(plazoDias, monto int) float64 {
	type rateEntry struct {
		maxDays int
		rate    float64
	}

	brackets := []rateEntry{
		{30, 0},
		{60, 0},
		{90, 0},
		{180, 0},
		{360, 0},
		{730, 0},
		{1095, 0},
	}

	baseRates := []float64{0.075, 0.08, 0.085, 0.092, 0.10, 0.105, 0.11}
	midRates := []float64{0.078, 0.083, 0.088, 0.095, 0.103, 0.108, 0.113}
	highRates := []float64{0.081, 0.086, 0.091, 0.098, 0.106, 0.111, 0.116}

	idx := 0
	for i, e := range brackets {
		if plazoDias <= e.maxDays {
			idx = i
			break
		}
		if i == len(brackets)-1 {
			idx = i
		}
	}

	var rates []float64
	if monto < 10000000 {
		rates = baseRates
	} else if monto < 50000000 {
		rates = midRates
	} else {
		rates = highRates
	}

	return rates[idx]
}

func calcularInteres(monto int, tasaEA float64, plazoDias int) int {
	factor := math.Pow(1+tasaEA, float64(plazoDias)/365.0) - 1
	return int(math.Round(float64(monto) * factor))
}

type RecommendRequest struct {
	ProfileID string `json:"profileId"`
}

func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req RecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	if req.ProfileID == "" {
		h.error(w, http.StatusBadRequest, "MISSING_PROFILE_ID", "profileId es requerido")
		return
	}

	profile, err := h.ragClient.GetProfile(req.ProfileID)
	if err != nil || profile == nil {
		h.error(w, http.StatusNotFound, "PROFILE_NOT_FOUND", "perfil no encontrado")
		return
	}

	var recomendacion, producto, accion string
	var citations []domain.Citation

	switch {
	case hasProduct(profile.Productos, "cuenta_ahorros") && !hasAnyProduct(profile.Productos, "cdt"):
		recomendacion = "Tienes una cuenta de ahorros y aún no inviertes. Un CDT te da rentabilidad fija con protección Fogafín hasta $50.000.000. ¿Te lo simulo?"
		producto = "cdt"
		accion = "simulate-cdt"
		chunks := h.ragClient.Retrieve("beneficios cdt supercdt", 3)
		citations = extractCitations(chunks)
	case hasProduct(profile.Productos, "tarjeta_clasica") && profile.MesesTarjeta >= 6:
		recomendacion = "Veo que tienes tu tarjeta hace más de 6 meses. Puedes solicitar un aumento de cupo. ¿Te ayudo?"
		producto = "tarjeta"
		accion = "aumento-cupo"
		chunks := h.ragClient.Retrieve("aumento cupo tarjeta", 3)
		citations = extractCitations(chunks)
	case hasProduct(profile.Productos, "credito_consumo"):
		recomendacion = "Tienes un crédito de consumo. ¿Ya conoces el débito automático desde tu cuenta Serfinanza para no perderte ningún pago?"
		producto = "credito_consumo"
		accion = "info-debito-automatico"
		chunks := h.ragClient.Retrieve("debito automatico", 3)
		citations = extractCitations(chunks)
	default:
		recomendacion = "Visitaré tu perfil para sugerirte los mejores productos cuando tengas más productos activos."
		producto = ""
		accion = ""
	}

	h.ok(w, domain.RecommendResponse{
		Recomendacion: recomendacion,
		Producto:      producto,
		Accion:        accion,
		Citations:     citations,
	})
}

func hasProduct(productos []string, product string) bool {
	for _, p := range productos {
		if p == product {
			return true
		}
	}
	return false
}

func hasAnyProduct(productos []string, targets ...string) bool {
	for _, p := range productos {
		for _, target := range targets {
			if p == target {
				return true
			}
		}
	}
	return false
}

func (h *Handler) WhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.ok(w, map[string]string{"status": "webhook_verified"})
		return
	}

	var payload domain.WhatsAppPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	phone, incomingText, ok := extractIncomingText(payload)
	if !ok {
		h.ok(w, map[string]interface{}{"received": true, "action": "ignored"})
		return
	}

	answer := h.processAsk(incomingText, "whatsapp", phone)

	mockWhatsapp := os.Getenv("MOCK_WHATSAPP") == "true"
	if !mockWhatsapp {
		go h.sendWhatsAppMessage(phone, answer)
	}

	h.ok(w, map[string]interface{}{
		"received":      true,
		"from":          phone,
		"incoming_text": incomingText,
		"answer":        answer,
	})
}

func extractIncomingText(payload domain.WhatsAppPayload) (string, string, bool) {
	if payload.Message == "" {
		return "", "", false
	}
	return payload.Phone, payload.Message, true
}

func (h *Handler) processAsk(text, channel, sessionID string) string {
	chunks := h.ragClient.Retrieve(text, 3)

	answer, err := h.llmClient.Generate(context.Background(), text, channel, chunks)
	if err != nil {
		answer = buildFallbackAnswer(chunks, channel)
	}

	if sessionID != "" {
		h.store.AddMessage(sessionID, "user", text)
		h.store.AddMessage(sessionID, "assistant", answer)
	}

	return answer
}

func buildFallbackAnswer(chunks []domain.Chunk, channel string) string {
	if len(chunks) == 0 {
		return "No tengo esa información en mis fuentes oficiales. Te recomiendo contactar a un asesor de Banco Serfinanza."
	}

	best := chunks[0]
	if channel == "asesor" {
		return fmt.Sprintf("Información recuperada de %s (%s): %s", best.Doc, best.Seccion, best.Contenido)
	}
	return fmt.Sprintf("Según la información oficial de Serfinanza: %s", best.Contenido)
}

func (h *Handler) sendWhatsAppMessage(to, text string) error {
	evolutionURL := os.Getenv("WHATSAPP_EVOLUTION_URL")
	instance := os.Getenv("WHATSAPP_EVOLUTION_INSTANCE")
	token := os.Getenv("WHATSAPP_EVOLUTION_TOKEN")

	if evolutionURL == "" || instance == "" || token == "" {
		return fmt.Errorf("configuración de WhatsApp incompleta")
	}

	url := fmt.Sprintf("%s/message/sendText/%s", evolutionURL, instance)

	body, _ := json.Marshal(map[string]string{"number": to, "text": text})
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func (h *Handler) ok(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
}

func (h *Handler) error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(domain.ErrorResponse{
		Error: domain.ErrorDetail{Code: code, Message: message},
	})
}

func extractCitations(chunks []domain.Chunk) []domain.Citation {
	citationMap := make(map[string]domain.Citation)
	for _, c := range chunks {
		if c.Doc != "" && c.Seccion != "" {
			key := c.Doc + "|" + c.Seccion
			citationMap[key] = domain.Citation{Doc: c.Doc, Seccion: c.Seccion}
		}
	}
	result := make([]domain.Citation, 0, len(citationMap))
	for _, v := range citationMap {
		result = append(result, v)
	}
	return result
}