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
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/javierg/hackathon-bqia/internal/auth"
	"github.com/javierg/hackathon-bqia/internal/domain"
	"github.com/javierg/hackathon-bqia/internal/llm"
	"github.com/javierg/hackathon-bqia/internal/rag"
	"github.com/javierg/hackathon-bqia/internal/session"
)

type Handler struct {
	llmClient *llm.Client
	ragClient *rag.Client
	store     *session.Store
	Users     []domain.User
}

func NewHandler(llmClient *llm.Client, ragClient *rag.Client, store *session.Store, users []domain.User) *Handler {
	return &Handler{
		llmClient: llmClient,
		ragClient: ragClient,
		store:     store,
		Users:     users,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.ok(w, map[string]string{
		"status":  "ok",
		"version": "1.0.1",
	})
}

func (h *Handler) Identify(w http.ResponseWriter, r *http.Request) {
	var req domain.IdentifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	user := auth.IdentifyUser(h.Users, req.Phone, req.ProfileID)
	if user == nil {
		user = auth.DefaultUser()
	}

	h.ok(w, domain.IdentifyResponse{
		UserID:      user.ID,
		Nombre:      user.Nombre,
		Role:        user.Role,
		ProfileID:   user.ProfileID,
		AllowedTags: user.AllowedTags,
	})
}

type AskRequest struct {
	Question  string `json:"question"`
	Channel   string `json:"channel"`
	SessionID string `json:"sessionId"`
	UserID    string `json:"userId,omitempty"`
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

	var allowedTags []string
	var profileContext string

	if req.UserID != "" {
		for _, u := range h.Users {
			if u.ID == req.UserID {
				allowedTags = auth.GetAllowedTags(&u)
				if u.ProfileID != "" {
					profile, _ := h.ragClient.GetProfile(u.ProfileID)
					if profile != nil {
						profileContext = buildProfileContext(u.Nombre, profile)
					}
				}
				break
			}
		}
	}
	if allowedTags == nil {
		allowedTags = []string{"publico", "general"}
	}

	answer, chunks := h.generateChatAnswer(r.Context(), req.Question, sessionID, allowedTags, profileContext)

	h.store.AddMessage(sessionID, "user", req.Question)
	h.store.AddMessage(sessionID, "assistant", answer)

	citations := extractCitations(chunks)

	h.ok(w, map[string]interface{}{
		"answer":    answer,
		"citations": citations,
		"grounded":  len(citations) > 0,
		"sessionId": sessionID,
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

	user := auth.IdentifyUser(h.Users, phone, "")
	profileContext := ""
	identified := false

	if user != nil && user.ProfileID != "" {
		profile, _ := h.ragClient.GetProfile(user.ProfileID)
		if profile != nil {
			profileContext = buildProfileContext(user.Nombre, profile)
			identified = true
		}
		h.store.SetSessionUser(phone, user.ID)
		if user.ProfileID != "" {
			h.store.SetSessionProfile(phone, user.ProfileID)
		}
	}

	if !identified && (strings.Contains(strings.ToLower(incomingText), "identificarme") ||
		strings.Contains(strings.ToLower(incomingText), "mi número") ||
		strings.Contains(strings.ToLower(incomingText), "soy ")) {

		answer := h.handleIdentificationRequest(phone, incomingText)
		mockWhatsapp := os.Getenv("MOCK_WHATSAPP") == "true"
		if !mockWhatsapp {
			go h.sendWhatsAppMessage(phone, answer)
		}
		h.ok(w, map[string]interface{}{
			"received":      true,
			"from":          phone,
			"incoming_text": incomingText,
			"answer":        answer,
			"identified":    false,
		})
		return
	}

	allowedTags := auth.GetAllowedTags(user)
	answer := h.processAskWithProfile(incomingText, "whatsapp", phone, allowedTags, profileContext)

	mockWhatsapp := os.Getenv("MOCK_WHATSAPP") == "true"
	if !mockWhatsapp {
		go h.sendWhatsAppMessage(phone, answer)
	}

	h.ok(w, map[string]interface{}{
		"received":      true,
		"from":          phone,
		"incoming_text": incomingText,
		"answer":        answer,
		"userId":        user.ID,
		"role":          user.Role,
		"identified":    identified,
	})
}

func (h *Handler) handleIdentificationRequest(sessionID, message string) string {
	phoneRe := regexp.MustCompile(`(?:mi número es|número|teléfono|celular)[\s:]*(\d+)`)
	profileRe := regexp.MustCompile(`(?i)(?:identificarme como|profile|id)[\s:]*([A-Za-z0-9]+)`)
	nameRe := regexp.MustCompile(`(?i)soy\s+([A-Za-zÀÉÍÓÚÑáéíóúñ\s]+?)(?:\s|,|\.|!|$)`)

	var phoneMatch, profileMatch, nameMatch []string

	if phoneMatch = phoneRe.FindStringSubmatch(message); len(phoneMatch) > 1 {
		cleanPhone := "+" + strings.TrimPrefix(phoneMatch[1], "0")
		if !strings.HasPrefix(cleanPhone, "+57") {
			cleanPhone = "+57" + phoneMatch[1]
		}
		user := auth.IdentifyUser(h.Users, cleanPhone, "")
		if user != nil {
			return h.doIdentifyUser(sessionID, user)
		}
	}

	if profileMatch = profileRe.FindStringSubmatch(message); len(profileMatch) > 1 {
		user := auth.IdentifyUser(h.Users, "", profileMatch[1])
		if user != nil {
			return h.doIdentifyUser(sessionID, user)
		}
	}

	if nameMatch = nameRe.FindStringSubmatch(message); len(nameMatch) > 1 {
		name := strings.TrimSpace(nameMatch[1])
		for _, u := range h.Users {
			if strings.Contains(strings.ToLower(u.Nombre), strings.ToLower(name)) {
				return h.doIdentifyUser(sessionID, &u)
			}
		}
	}

	return "No pude identificarte. Por favor verifica tu número de teléfono o ingresa tu ID de cliente (ej: C001)."
}

func (h *Handler) doIdentifyUser(sessionID string, user *domain.User) string {
	h.store.SetSessionUser(sessionID, user.ID)
	if user.ProfileID != "" {
		h.store.SetSessionProfile(sessionID, user.ProfileID)
	}

	greeting := fmt.Sprintf("¡Hola %s! Te he identificado correctamente.", user.Nombre)

	profile, _ := h.ragClient.GetProfile(user.ProfileID)
	if profile != nil && len(profile.Productos) > 0 {
		productos := strings.Join(profile.Productos, ", ")
		return fmt.Sprintf("%s Veo que tienes los siguientes productos: %s. ¿En qué puedo ayudarte?", greeting, productos)
	}

	return fmt.Sprintf("%s ¿En qué puedo ayudarte?", greeting)
}

func buildProfileContext(nombre string, profile *domain.Profile) string {
	if profile == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- Nombre: %s\n", nombre))

	if len(profile.Productos) > 0 {
		sb.WriteString(fmt.Sprintf("- Productos actuales: %s\n", strings.Join(profile.Productos, ", ")))
	}

	if profile.Tarjeta != nil && *profile.Tarjeta != "" {
		sb.WriteString(fmt.Sprintf("- Tarjeta: %s (hace %d meses)\n", *profile.Tarjeta, profile.MesesTarjeta))
	}

	return sb.String()
}

func extractIncomingText(payload domain.WhatsAppPayload) (string, string, bool) {
	if payload.Message == "" {
		return "", "", false
	}
	return payload.Phone, payload.Message, true
}

func (h *Handler) processAsk(text, channel, sessionID string) string {
	return h.processAskWithTags(text, channel, sessionID, nil)
}

func (h *Handler) processAskWithTags(text, channel, sessionID string, allowedTags []string) string {
	return h.processAskWithProfile(text, channel, sessionID, allowedTags, "")
}

func (h *Handler) processAskWithProfile(text, channel, sessionID string, allowedTags []string, profileContext string) string {
	answer, _ := h.generateChatAnswer(context.Background(), text, sessionID, allowedTags, profileContext)

	if sessionID != "" {
		h.store.AddMessage(sessionID, "user", text)
		h.store.AddMessage(sessionID, "assistant", answer)
	}

	return answer
}

func (h *Handler) generateChatAnswer(ctx context.Context, question, sessionID string, allowedTags []string, profileContext string) (string, []domain.Chunk) {
	_, intent := rag.ResolveQuery(question)
	chunks := h.ragClient.RetrieveWithTags(question, 8, allowedTags)
	if len(chunks) == 0 && rag.HasBankIntent(question) {
		chunks = h.ragClient.RetrieveWithTags(rag.RetrieveQuery(question), 8, allowedTags)
	}

	chunks = rag.FilterChunksForIntent(intent, chunks)

	history := h.store.GetMessages(sessionID)

	req := llm.ClientRequest{
		Question:       question,
		Chunks:         chunks,
		ProfileContext: profileContext,
		History:        history,
	}

	answer, err := h.llmClient.GenerateForClient(ctx, req)
	if err != nil {
		answer = buildFallbackAnswer(question, chunks, intent)
	}

	return answer, chunks
}

func buildFallbackAnswer(question string, chunks []domain.Chunk, intent rag.QueryIntent) string {
	switch {
	case rag.IsUrgentCard(question):
		if text := rag.BestChunkForKeywords(chunks, "bloque", "robo", "BLOQUEAR"); text != "" {
			return formatUrgentCardAnswer(text)
		}
		return rag.UrgentCardReply
	case rag.IsGreeting(question):
		return rag.GreetingReply
	case rag.IsThanks(question):
		return rag.ThanksReply
	case rag.IsFarewell(question):
		return rag.FarewellReply
	case rag.IsAcknowledgment(question):
		return rag.AckReply
	case rag.IsConfused(question):
		return rag.ConfusedReply
	}

	switch intent.ID {
	case "portal_access":
		if chunk := rag.SelectBestChunk(intent, chunks); chunk != "" && rag.IsPortalLoginChunk(chunk) {
			if text := formatHelpfulAnswer("Para ingresar a Serfinanza Virtual / Banca en Línea:", chunk); text != "" {
				return text
			}
		}
		return rag.PortalAccessReply
	case "plan_ahorro":
		return rag.PlanAhorroReply
	case "extracto":
		return rag.ExtractoGuideReply
	case "actualizacion_datos":
		return rag.ActualizacionDatosReply
	case "cdt_beneficios":
		if text := formatHelpfulAnswer("Estos son los beneficios del CDT Serfinanza (superCDT):", rag.SelectBestChunk(intent, chunks)); text != "" {
			return text
		}
	}

	if rag.HasBankIntent(question) {
		if text := rag.SelectBestChunk(intent, chunks); text != "" {
			return formatHelpfulAnswer("", text)
		}
	}

	return rag.CasualFallbackReply
}

func formatHelpfulAnswer(prefix, body string) string {
	body = cleanChunkForDisplay(body)
	lower := strings.ToLower(body)
	if rag.IsBoilerplate(body) || strings.Contains(lower, "sarlaft") || strings.Contains(lower, "extractos y documentos") {
		return ""
	}
	body = truncateFallback(body, 520)
	if body == "" {
		return ""
	}
	if prefix == "" {
		return body
	}
	return prefix + " " + body
}

func cleanChunkForDisplay(text string) string {
	lower := strings.ToLower(text)
	startMarkers := []string{
		"ingresa a la app",
		"accede a www.serfinanza",
		"escribe al +57",
		"la app es el canal",
		"para generar",
		"desde el portal",
	}
	best := len(text)
	for _, m := range startMarkers {
		if i := strings.Index(lower, m); i >= 0 && i < best {
			best = i
		}
	}
	if best > 0 && best < len(text) {
		text = text[best:]
	}

	text = strings.ReplaceAll(text, " n ", ". ")
	text = strings.ReplaceAll(text, "® ", "\n• ")
	text = strings.ReplaceAll(text, " ", "")
	// Fragmentos cortados al inicio (ej. "atrás).")
	if idx := strings.Index(text, ")."); idx >= 0 && idx < 40 {
		if rest := strings.TrimSpace(text[idx+2:]); rest != "" {
			text = rest
		}
	}
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	return strings.TrimSpace(text)
}

func formatUrgentCardAnswer(raw string) string {
	const prefix = "Entiendo, es urgente. "
	if idx := strings.Index(strings.ToLower(raw), "bloque"); idx >= 0 {
		raw = raw[idx:]
	}
	return prefix + truncateFallback(raw, 460)
}

func truncateFallback(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if len(text) <= maxLen {
		return text
	}
	cut := text[:maxLen]
	if idx := strings.LastIndex(cut, " "); idx > maxLen/2 {
		cut = cut[:idx]
	}
	return cut + "…"
}

func hasUsefulContent(content string) bool {
	if rag.IsBoilerplate(content) {
		return false
	}
	if len(content) < 30 {
		return false
	}

	letterCount := 0
	for _, r := range content {
		if unicode.IsLetter(r) {
			letterCount++
		}
	}

	if float64(letterCount)/float64(len(content)) < 0.3 {
		return false
	}

	return true
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

func (h *Handler) ListKnowledge(w http.ResponseWriter, r *http.Request) {
	items := h.ragClient.ListKnowledge()
	h.ok(w, domain.KnowledgeListResponse{Items: items})
}

func (h *Handler) AddKnowledge(w http.ResponseWriter, r *http.Request) {
	var req domain.AddKnowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	if req.Doc == "" || req.Seccion == "" || req.Contenido == "" {
		h.error(w, http.StatusBadRequest, "MISSING_FIELDS", "doc, seccion y contenido son requeridos")
		return
	}

	id, err := h.ragClient.AddKnowledge(req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "ADD_FAILED", err.Error())
		return
	}

	h.ok(w, domain.AddKnowledgeResponse{ID: id})
}

func (h *Handler) UpdateKnowledge(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "MISSING_ID", "id es requerido")
		return
	}

	var req domain.UpdateKnowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	id, before, err := h.ragClient.UpdateKnowledge(id, req)
	if err != nil {
		h.error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	h.ok(w, domain.UpdateKnowledgeResponse{
		ID:     id,
		Before: before,
		After:  "",
	})
}

func (h *Handler) DeleteKnowledge(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "MISSING_ID", "id es requerido")
		return
	}

	err := h.ragClient.DeleteKnowledge(id)
	if err != nil {
		h.error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	h.ok(w, domain.DeleteKnowledgeResponse{Deleted: true})
}

func (h *Handler) ReloadKnowledge(w http.ResponseWriter, r *http.Request) {
	count, err := h.ragClient.ReloadKnowledge()
	if err != nil {
		h.error(w, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	h.ok(w, domain.ReloadKnowledgeResponse{Count: count})
}

func (h *Handler) GetScope(w http.ResponseWriter, r *http.Request) {
	scope := h.ragClient.GetScope()
	h.ok(w, scope)
}

func (h *Handler) SetScope(w http.ResponseWriter, r *http.Request) {
	var req domain.SetScopeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	scope := h.ragClient.SetScope(req)
	h.ok(w, scope)
}
