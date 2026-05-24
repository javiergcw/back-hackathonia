package rag

import (
	"regexp"
	"strings"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

// QueryIntent agrupa la consulta RAG optimizada y pistas de relevancia.
type QueryIntent struct {
	ID            string
	RetrieveQuery string
	Keywords      []string
	DocHints      []string
}

var conversationalPrefix = regexp.MustCompile(`(?i)^((hola|buenos d[ií]as|buenas tardes|buenas noches|hey|hi|saludos|qu[eé] tal)[,.!\s]+)+`)

// StripConversationalPrefix quita saludos al inicio para mejorar la búsqueda.
func StripConversationalPrefix(q string) string {
	q = strings.TrimSpace(q)
	for i := 0; i < 3; i++ {
		next := conversationalPrefix.ReplaceAllString(q, "")
		next = strings.TrimSpace(next)
		if next == q {
			break
		}
		q = next
	}
	return q
}

// ResolveQuery devuelve la consulta de retrieval y la intención detectada.
func ResolveQuery(question string) (string, QueryIntent) {
	raw := strings.TrimSpace(question)
	core := StripConversationalPrefix(raw)
	lower := strings.ToLower(core)

	if IsUrgentCard(raw) {
		return "bloquear tarjeta robo perdida app whatsapp BLOQUEAR call center", QueryIntent{
			ID: "urgent_card",
		}
	}

	intent := detectIntent(lower, core)
	if intent.RetrieveQuery == "" {
		intent.RetrieveQuery = core
	}
	if intent.RetrieveQuery == "" {
		intent.RetrieveQuery = raw
	}

	if intent.ID == "" && (IsGreeting(raw) || (core != "" && IsConversational(core) && !HasBankIntent(core))) {
		return "serfinanza banco productos servicios canales atencion app web sucursal", QueryIntent{ID: "small_talk"}
	}

	return intent.RetrieveQuery, intent
}

func detectIntent(lower, core string) QueryIntent {
	switch {
	case containsAny(lower,
		"virtual personas", "serfinanza virtual", "banca en linea", "banca en línea",
		"portal web", "ingresar a serfinanza", "ingreso a serfinanza",
		"como ingreso", "cómo ingreso", "como entro", "cómo entro",
	):
		return QueryIntent{
			ID:            "portal_access",
			RetrieveQuery: "ingresar banca en linea serfinanza virtual usuario contraseña registrarse app login portal web",
			Keywords:      []string{"banca en línea", "banca en linea", "usuario y contraseña", "registr", "inicio de sesión", "ingresa a la app"},
			DocHints:      []string{"04_registro", "02_actualizacion"},
		}

	case containsAny(lower,
		"plan de ahorro", "ahorro programado", "radicacion de plan", "radicación de plan",
	) || (strings.Contains(lower, "radicacion") && strings.Contains(lower, "ahorro")) ||
		(strings.Contains(lower, "radicación") && strings.Contains(lower, "ahorro")):
		return QueryIntent{
			ID:            "plan_ahorro",
			RetrieveQuery: "plan ahorro programado radicacion cuenta ahorros domiciliacion",
			Keywords:      []string{"plan de ahorro", "ahorro programado", "cuenta de ahorros", "radicacion"},
			DocHints:      []string{"04_registro"},
		}

	case strings.Contains(lower, "extracto") || containsAny(lower, "generar mi extracto", "genero mi extracto", "leer mi extracto"):
		return QueryIntent{
			ID:            "extracto",
			RetrieveQuery: "generar extracto mensual pdf app documentos extractos leer saldo pago minimo movimientos",
			Keywords:      []string{"extracto", "Generar", "Documentos", "Extractos", "PDF", "pago mínimo", "movimientos"},
			DocHints:      []string{"05_extracto"},
		}

	case containsAny(lower,
		"actualizacion de datos", "actualización de datos", "actualizar mis datos",
		"actualizar datos", "por que canal", "por qué canal", "canales puedo",
		"en que canal", "en qué canal", "donde actualizo", "dónde actualizo",
	) || (strings.Contains(lower, "canal") && strings.Contains(lower, "datos")) ||
		(strings.Contains(lower, "actualizar") && containsAny(lower, "datos", "telefono", "teléfono", "correo", "direccion", "dirección")):
		return QueryIntent{
			ID:            "actualizacion_datos",
			RetrieveQuery: "actualizacion datos canales app portal web whatsapp call center sucursal ACTUALIZAR Mi Perfil",
			Keywords:      []string{"App Serfinanza", "Portal Web", "WhatsApp", "Call Center", "Sucursales", "ACTUALIZAR", "Mi Perfil", "Banca en Línea"},
			DocHints:      []string{"02_actualizacion"},
		}

	case containsAny(lower, "supercdt", "super cdt") || (strings.Contains(lower, "cdt") && containsAny(lower, "beneficio", "ventaja", "que tiene")):
		return QueryIntent{
			ID:            "cdt_beneficios",
			RetrieveQuery: "beneficios CDT Serfinanza supercdt rentabilidad fogafin plazos",
			Keywords:      []string{"beneficios", "Fogafín", "rentabilidad", "CDT", "tasa"},
			DocHints:      []string{"03_cdt"},
		}
	}

	return QueryIntent{RetrieveQuery: core}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// FilterChunksForIntent deja solo fragmentos coherentes con la intención (para el LLM).
func FilterChunksForIntent(intent QueryIntent, chunks []domain.Chunk) []domain.Chunk {
	if intent.ID == "" {
		return chunks
	}
	filtered := make([]domain.Chunk, 0, len(chunks))
	for _, c := range chunks {
		switch intent.ID {
		case "plan_ahorro":
			if IsPlanAhorroChunk(c.Contenido) {
				filtered = append(filtered, c)
			}
		case "extracto":
			if IsExtractoGuideChunk(c.Contenido) {
				filtered = append(filtered, c)
			}
		case "portal_access":
			if IsPortalLoginChunk(c.Contenido) {
				filtered = append(filtered, c)
			}
		case "actualizacion_datos":
			if IsActualizacionDatosChunk(c.Contenido) {
				filtered = append(filtered, c)
			}
		default:
			if !IsBoilerplate(c.Contenido) {
				filtered = append(filtered, c)
			}
		}
	}
	if len(filtered) == 0 {
		if intent.ID == "actualizacion_datos" || intent.ID == "plan_ahorro" || intent.ID == "extracto" || intent.ID == "portal_access" {
			return nil
		}
		return chunks
	}
	return filtered
}

// SelectBestChunk elige el fragmento más útil para fallback (sin pies de página).
func SelectBestChunk(intent QueryIntent, chunks []domain.Chunk) string {
	best := ""
	bestScore := -1.0

	for _, chunk := range chunks {
		if IsBoilerplate(chunk.Contenido) {
			continue
		}
		score := selectionScore(chunk, intent)
		if score > bestScore {
			bestScore = score
			best = chunk.Contenido
		}
	}
	return best
}

func selectionScore(chunk domain.Chunk, intent QueryIntent) float64 {
	content := strings.ToLower(chunk.Contenido)
	doc := strings.ToLower(chunk.Doc)
	seccion := strings.ToLower(chunk.Seccion)
	var score float64

	for _, hint := range intent.DocHints {
		if strings.Contains(doc, strings.ToLower(hint)) {
			score += 10
		}
	}
	for _, kw := range intent.Keywords {
		kw = strings.ToLower(kw)
		if strings.Contains(content, kw) {
			score += 3
		}
		if strings.Contains(seccion, kw) {
			score += 4
		}
	}
	if strings.Contains(content, "paso") || strings.Contains(content, "®") {
		score += 2
	}
	if strings.Contains(content, "sarlaft") && intent.ID != "actualizacion_datos" {
		score -= 8
	}
	if intent.ID == "plan_ahorro" {
		if strings.Contains(doc, "01_tarjeta") || strings.Contains(doc, "tarjeta_credito") {
			score -= 30
		}
		if strings.Contains(content, "pago mínimo") || strings.Contains(content, "pago total") ||
			strings.Contains(content, "pago libre") || strings.Contains(content, "factura") {
			score -= 25
		}
	}
	if intent.ID == "extracto" {
		if strings.Contains(doc, "05_extracto") {
			score += 8
		}
		if strings.Contains(content, "documentos") || strings.Contains(content, "generar") {
			score += 5
		}
		if strings.Contains(content, "pago mínimo") || strings.Contains(content, "movimientos del período") {
			score += 4
		}
	}
	if intent.ID == "portal_access" {
		if strings.Contains(doc, "05_extracto") {
			score -= 25
		}
		if strings.Contains(content, "extractos y documentos") || strings.Contains(content, "historial disponible") {
			score -= 20
		}
		if strings.Contains(content, "banca en línea") || strings.Contains(content, "banca en linea") {
			score += 12
		}
		if strings.Contains(doc, "04_registro") {
			score += 12
		}
		if strings.Contains(seccion, "registr") || strings.Contains(content, "regístrate") {
			score += 8
		}
	}
	return score
}

// IsActualizacionDatosChunk fragmento sobre canales o pasos de actualización de datos.
func IsActualizacionDatosChunk(content string) bool {
	lower := strings.ToLower(content)
	if IsBoilerplate(content) {
		return false
	}
	return strings.Contains(lower, "actualiz") &&
		(strings.Contains(lower, "app serfinanza") || strings.Contains(lower, "whatsapp") ||
			strings.Contains(lower, "call center") || strings.Contains(lower, "sucursal") ||
			strings.Contains(lower, "banca en línea") || strings.Contains(lower, "portal web") ||
			strings.Contains(lower, "mi perfil") || strings.Contains(lower, "actualizar"))
}

// IsPlanAhorroChunk evita confundir el plan de ahorro con el pago automático de tarjeta.
func IsPlanAhorroChunk(content string) bool {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "pago mínimo") || strings.Contains(lower, "pago total") ||
		strings.Contains(lower, "pago libre") || strings.Contains(lower, "factura") {
		return false
	}
	return strings.Contains(lower, "plan de ahorro") || strings.Contains(lower, "ahorro programado") ||
		(strings.Contains(lower, "cuenta de ahorros") && strings.Contains(lower, "radic"))
}

// IsExtractoGuideChunk fragmento útil para generar o leer extractos (no pies de página).
func IsExtractoGuideChunk(content string) bool {
	lower := strings.ToLower(content)
	if IsBoilerplate(content) {
		return false
	}
	return strings.Contains(lower, "extracto") &&
		(strings.Contains(lower, "generar") || strings.Contains(lower, "documentos") ||
			strings.Contains(lower, "pdf") || strings.Contains(lower, "pago mínimo") ||
			strings.Contains(lower, "movimientos"))
}

// IsPortalLoginChunk indica si el fragmento habla de ingreso/registro, no de extractos u otros trámites.
func IsPortalLoginChunk(content string) bool {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "extractos y documentos") || strings.Contains(lower, "historial disponible: hasta 36") {
		return false
	}
	hasAccess := strings.Contains(lower, "banca en línea") ||
		strings.Contains(lower, "banca en linea") ||
		strings.Contains(lower, "usuario y contraseña") ||
		strings.Contains(lower, "regístrate") ||
		strings.Contains(lower, "inicio de sesión")
	return hasAccess
}
