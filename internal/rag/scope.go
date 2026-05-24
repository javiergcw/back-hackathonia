package rag

import (
	"regexp"
	"strings"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

const minRelevanceScore = 3.0

type RetrieveResult struct {
	Chunks   []domain.Chunk
	TopScore float64
}

var (
	mathQueryPattern = regexp.MustCompile(`(?i)(cu[aá]nto es|\d+\s*[\+\-\*x×/÷]\s*\d+)`)
	offTopicPattern  = regexp.MustCompile(`(?i)(presidente|capital de|clima|futbol|f[uú]tbol|receta|poema|chiste|historia de|quien invent[oó]|gpt|chatgpt|pel[ií]cula|m[uú]sica|tiktok|instagram)`)
	greetingPattern  = regexp.MustCompile(`(?i)^(hola|buenos d[ií]as|buenas tardes|buenas noches|hey|hi|hello|saludos|qu[eé] tal|que tal)[\s!.?]*$`)
	thanksPattern    = regexp.MustCompile(`(?i)^(gracias|muchas gracias|mil gracias|te agradezco|ok gracias|vale gracias)[\s!.?]*$`)
	farewellPattern  = regexp.MustCompile(`(?i)^(chao|adi[oó]s|hasta luego|nos vemos|bye|bueno gracias)[\s!.?]*$`)
	smallTalkPattern  = regexp.MustCompile(`(?i)^(c[oó]mo est[aá]s|como estas|todo bien|qu[eé] hay|que hay|me ayudas|me puedes ayudar|necesito ayuda)[\s!.?]*$`)
	ackPattern        = regexp.MustCompile(`(?i)^(ok|okay|vale|listo|perfecto|entendido|de acuerdo|dale|bueno)[\s!.?]*$`)
	confusedPattern   = regexp.MustCompile(`(?i)(no entend[ií]|no me qued[oó] claro|expl[ií]came mejor|estoy confundid|no comprend[ií])`)
	urgentCardPattern = regexp.MustCompile(`(?i)((robo|robar|robaron|robado|hurto|extravi|perd[ií]|clonaron|fraude|no reconozco).*(tarjeta|pl[aá]stico|tarjetas))|(tarjeta.*(robo|robar|robaron|perd[ií]|bloque|urgente|reportar))|(reportar.*tarjeta)|(bloque(ar|o).*(tarjeta|ya|urgente))`)
	bankHintPattern   = regexp.MustCompile(`(?i)(serfinanza|banco|cdt|supercdt|tarjeta|cr[eé]dito|cuenta|ahorro|app|extracto|fogaf[ií]n|cupo|pago|debito|d[eé]bito|inversi[oó]n|simul|bloque|otp|sucursal|call center|robo|robar|reportar)`)
)

const GreetingReply = "¡Hola! Soy tu asesor de Banco Serfinanza. ¿En qué te puedo ayudar hoy? Puedo orientarte con CDT, tarjeta, crédito, la App o cualquier trámite."

const ThanksReply = "¡Con gusto! Si más adelante tienes otra duda sobre tus productos o trámites en Serfinanza, aquí estoy."

const FarewellReply = "¡Que tengas un excelente día! Cuando quieras, escríbeme si necesitas algo de Serfinanza."

const OutOfScopeReply = "Ese tema no lo manejo yo, pero con gusto te ayudo con Serfinanza: cuenta, tarjeta, CDT, crédito, App o trámites. ¿Qué necesitas?"

const CasualFallbackReply = "Cuéntame qué necesitas y te oriento: CDT, tarjeta, crédito, pagos, extractos o la App de Serfinanza. También puedes llamar al 01 8000 123 456 o escribir por WhatsApp oficial del banco."

const UrgentCardReply = "Entiendo, vamos de inmediato. Bloquea tu tarjeta ya: (1) App Serfinanza → Tarjetas → Bloquear tarjeta, (2) WhatsApp oficial escribiendo BLOQUEAR, o (3) llama al 01 8000 123 456. El bloqueo es instantáneo y sin costo. ¿Ya pudiste bloquearla?"

const ConfusedReply = "Sin problema, lo vemos más simple. ¿Tu consulta es sobre tarjeta, CDT, crédito, la App o un trámite? Cuéntame en una frase y te guío paso a paso."

const AckReply = "Perfecto. ¿Te ayudo con algo más de Serfinanza?"

func IsGreeting(query string) bool {
	return greetingPattern.MatchString(strings.TrimSpace(query))
}

func IsThanks(query string) bool {
	return thanksPattern.MatchString(strings.TrimSpace(query))
}

func IsFarewell(query string) bool {
	return farewellPattern.MatchString(strings.TrimSpace(query))
}

func IsSmallTalk(query string) bool {
	return smallTalkPattern.MatchString(strings.TrimSpace(query))
}

func IsAcknowledgment(query string) bool {
	return ackPattern.MatchString(strings.TrimSpace(query))
}

func IsConfused(query string) bool {
	return confusedPattern.MatchString(strings.TrimSpace(query))
}

func IsUrgentCard(query string) bool {
	return urgentCardPattern.MatchString(strings.TrimSpace(query))
}

func HasBankIntent(query string) bool {
	return bankHintPattern.MatchString(strings.TrimSpace(query))
}

func IsConversational(query string) bool {
	q := strings.TrimSpace(query)
	return IsGreeting(q) || IsThanks(q) || IsFarewell(q) || IsSmallTalk(q) || IsAcknowledgment(q)
}

// RetrieveQuery amplía la búsqueda RAG según la intención del mensaje.
func RetrieveQuery(query string) string {
	q := strings.TrimSpace(query)
	if IsUrgentCard(q) {
		return "bloquear tarjeta robo perdida app whatsapp BLOQUEAR call center"
	}
	if IsConversational(q) && !HasBankIntent(q) {
		return "serfinanza banco productos servicios canales atencion app web sucursal"
	}
	return q
}

// BestChunkForKeywords devuelve el contenido del fragmento que mencione alguna palabra clave.
func BestChunkForKeywords(chunks []domain.Chunk, keywords ...string) string {
	for _, kw := range keywords {
		kw = strings.ToLower(kw)
		for _, chunk := range chunks {
			if strings.Contains(strings.ToLower(chunk.Contenido), kw) {
				return chunk.Contenido
			}
		}
	}
	return ""
}

func IsOffTopic(query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return true
	}
	if bankHintPattern.MatchString(q) {
		return false
	}
	if mathQueryPattern.MatchString(q) {
		return true
	}
	return offTopicPattern.MatchString(q)
}

func IsInScope(query string, result RetrieveResult) bool {
	if IsGreeting(query) {
		return true
	}
	if IsOffTopic(query) {
		return false
	}
	if bankHintPattern.MatchString(query) {
		return true
	}
	if len(result.Chunks) == 0 {
		return false
	}
	return result.TopScore >= minRelevanceScore
}
