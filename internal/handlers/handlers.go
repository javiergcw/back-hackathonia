package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/javierg/hackathon-bqia/internal/domain"
	"github.com/javierg/hackathon-bqia/internal/llm"
	"github.com/javierg/hackathon-bqia/internal/rag"
	"github.com/javierg/hackathon-bqia/internal/session"
	"github.com/javierg/hackathon-bqia/internal/whatsapp"
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
	SessionID string `json:"sessionId"`
	ProfileID string `json:"profileId"`
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

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	profile, _ := h.ragClient.GetProfile(req.ProfileID)
	retrieval := h.ragClient.RetrieveForClient(req.Question, profile, 4)
	answer := h.generateClientAnswer(r.Context(), req.Question, retrieval, profile, sessionID)

	response := map[string]interface{}{
		"answer":    answer,
		"citations": extractCitations(retrieval.Chunks),
		"grounded":  len(retrieval.Chunks) > 0 && rag.IsInScope(req.Question, retrieval),
		"sessionId": sessionID,
	}
	if profile != nil {
		response["cliente"] = profile.Nombre
	}

	h.ok(w, response)
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

	recomendacion, producto, accion, citations := h.clientRecommendation(profile)

	h.ok(w, domain.RecommendResponse{
		Recomendacion: recomendacion,
		Producto:      producto,
		Accion:        accion,
		Citations:     citations,
	})
}

func (h *Handler) clientRecommendation(profile *domain.Profile) (string, string, string, []domain.Citation) {
	switch {
	case rag.HasProduct(profile.Productos, "cuenta_ahorros") && !rag.HasAnyProduct(profile.Productos, "cdt"):
		chunks := h.ragClient.RetrieveForClient("beneficios cdt supercdt", profile, 3).Chunks
		return "Veo que tienes cuenta de ahorros y aún no inviertes. Con nuestro *superCDT* tu plata crece con rentabilidad fija y protección Fogafín. ¿Te gustaría que te cuente más?", "cdt", "simulate-cdt", extractCitations(chunks)
	case rag.HasProduct(profile.Productos, "tarjeta_clasica") && profile.MesesTarjeta >= 6:
		chunks := h.ragClient.RetrieveForClient("aumento cupo tarjeta", profile, 3).Chunks
		return "Llevas más de 6 meses con tu tarjeta y puedes solicitar un *aumento de cupo* desde la App. ¿Te explico cómo?", "tarjeta", "aumento-cupo", extractCitations(chunks)
	case rag.HasProduct(profile.Productos, "credito_consumo"):
		chunks := h.ragClient.RetrieveForClient("debito automatico", profile, 3).Chunks
		return "Tienes crédito de consumo con nosotros. Te recomiendo activar el *débito automático* para no perder ningún pago. ¿Quieres saber cómo hacerlo?", "credito_consumo", "info-debito-automatico", extractCitations(chunks)
	default:
		return "Cuando quieras, con gusto te oriento sobre los productos de Serfinanza.", "", "", nil
	}
}

func (h *Handler) WhatsAppWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.ok(w, map[string]string{"status": "webhook_verified"})
		return
	}

	if keyNumber := chi.URLParam(r, "number"); keyNumber != "" {
		if whatsapp.NormalizePhone(keyNumber) != whatsapp.AllowedNumber() {
			h.error(w, http.StatusForbidden, "UNAUTHORIZED_WEBHOOK", "número de webhook no autorizado")
			return
		}
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_BODY", "no se pudo leer el cuerpo")
		return
	}

	incoming, ok := whatsapp.ParseWebhook(raw)
	if !ok {
		log.Printf("whatsapp webhook: payload no reconocido (%d bytes)", len(raw))
		h.ok(w, map[string]interface{}{"received": true, "action": "ignored"})
		return
	}

	if !whatsapp.IsAllowedPhone(incoming.Phone) {
		log.Printf("whatsapp webhook: número no autorizado %s", incoming.Phone)
		h.ok(w, map[string]interface{}{
			"received": true,
			"action":   "ignored",
			"reason":   "unauthorized_number",
		})
		return
	}

	log.Printf("whatsapp webhook: mensaje de %s: %q", incoming.Phone, incoming.Text)

	profileID := whatsapp.UserProfileID()
	var payload domain.WhatsAppPayload
	_ = json.Unmarshal(raw, &payload)
	if payload.ProfileID != "" {
		profileID = payload.ProfileID
	}

	profile, _ := h.ragClient.GetProfile(profileID)
	answer := h.processClientAsk(incoming.Text, incoming.Phone, profile)

	mockWhatsapp := os.Getenv("MOCK_WHATSAPP") == "true"
	if !mockWhatsapp {
		go func(phone, reply string) {
			if err := h.sendWhatsAppMessage(phone, reply); err != nil {
				log.Printf("whatsapp send error to %s: %v", phone, err)
			}
		}(incoming.Phone, answer)
	}

	response := map[string]interface{}{
		"received":      true,
		"from":          incoming.Phone,
		"incoming_text": incoming.Text,
		"answer":        answer,
		"profileId":     profileID,
	}
	if profile != nil {
		response["cliente"] = profile.Nombre
	}

	h.ok(w, response)
}

func (h *Handler) processClientAsk(text, sessionID string, profile *domain.Profile) string {
	retrieval := h.ragClient.RetrieveForClient(text, profile, 4)
	return h.generateClientAnswer(context.Background(), text, retrieval, profile, sessionID)
}

func (h *Handler) generateClientAnswer(ctx context.Context, question string, retrieval rag.RetrieveResult, profile *domain.Profile, sessionID string) string {
	if rag.IsGreeting(question) {
		answer := llm.FormatForWhatsApp(rag.GreetingReply)
		h.saveSession(sessionID, question, answer)
		return answer
	}
	if !rag.IsInScope(question, retrieval) {
		answer := llm.FormatForWhatsApp(rag.OutOfScopeReply)
		h.saveSession(sessionID, question, answer)
		return answer
	}

	history := h.store.GetMessages(sessionID)

	answer, err := h.llmClient.GenerateForClient(ctx, llm.ClientRequest{
		Question:       question,
		Chunks:         retrieval.Chunks,
		ProfileContext: rag.ProfileContext(profile),
		ProactiveHint:  rag.ProactiveHint(profile, question),
		History:        history,
	})
	if err != nil {
		answer = buildClientFallback(retrieval.Chunks)
	}
	answer = llm.FormatForWhatsApp(answer)
	h.saveSession(sessionID, question, answer)
	return answer
}

func (h *Handler) saveSession(sessionID, question, answer string) {
	if sessionID != "" {
		h.store.AddMessage(sessionID, "user", question)
		h.store.AddMessage(sessionID, "assistant", answer)
	}
}

func buildClientFallback(chunks []domain.Chunk) string {
	if len(chunks) == 0 {
		return "No tengo ese dato a la mano, pero un asesor de Banco Serfinanza puede ayudarte con gusto por WhatsApp, la App o en sucursal."
	}
	return fmt.Sprintf("Te cuento: %s", chunks[0].Contenido)
}

func (h *Handler) sendWhatsAppMessage(to, text string) error {
	evolutionURL := os.Getenv("WHATSAPP_EVOLUTION_URL")
	if evolutionURL == "" {
		evolutionURL = os.Getenv("EVOLUTION_API_URL")
	}
	instance := os.Getenv("WHATSAPP_EVOLUTION_INSTANCE")
	if instance == "" {
		instance = os.Getenv("EVOLUTION_INSTANCE")
	}
	token := os.Getenv("WHATSAPP_EVOLUTION_TOKEN")
	if token == "" {
		token = os.Getenv("EVOLUTION_API_KEY")
	}

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
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("evolution respondió %d: %s", resp.StatusCode, string(respBody))
	}
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
