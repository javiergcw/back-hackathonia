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
			RetrieveQuery: "portal web banca en linea ingresar usuario contraseña www.serfinanza.com registro app",
			Keywords:      []string{"banca en línea", "www.serfinanza", "ingresa", "portal", "registr", "App Serfinanza"},
			DocHints:      []string{"04_registro", "02_actualizacion", "05_extracto"},
		}

	case containsAny(lower,
		"plan de ahorro", "ahorro programado", "debito automatico", "débito automático",
		"debito automatico", "radicacion de plan", "radicación de plan",
	):
		return QueryIntent{
			ID:            "plan_ahorro",
			RetrieveQuery: "plan ahorro debito automatico cuenta ahorros radicacion domiciliacion",
			Keywords:      []string{"ahorro", "automático", "débito", "cuenta"},
			DocHints:      []string{"01_tarjeta", "04_registro"},
		}

	case containsAny(lower, "extracto", "leerlo", "leer mi extracto", "generar mi extracto", "genero mi extracto"):
		return QueryIntent{
			ID:            "extracto",
			RetrieveQuery: "generar extracto mensual pdf app portal leer extracto tarjeta credito movimientos",
			Keywords:      []string{"extracto", "Generar", "PDF", "corte", "saldo", "Documentos"},
			DocHints:      []string{"05_extracto", "04_registro"},
		}

	case containsAny(lower,
		"actualizacion de datos", "actualización de datos", "actualizar mis datos",
		"actualizar datos", "por que canal", "por qué canal", "canales puedo",
	) || (strings.Contains(lower, "canal") && strings.Contains(lower, "datos")):
		return QueryIntent{
			ID:            "actualizacion_datos",
			RetrieveQuery: "actualizacion datos canales app portal whatsapp call center sucursal ACTUALIZAR",
			Keywords:      []string{"App Serfinanza", "WhatsApp", "Call Center", "sucursal", "ACTUALIZAR", "Mi Perfil"},
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
	return score
}
