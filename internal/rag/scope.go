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

const PlanAhorroReply = "Sobre la radicación de un plan de ahorro con débito automático: en la guía digital no está el paso a paso de ese producto específico (no confundirlo con el pago automático de la tarjeta de crédito). Para radicarlo correctamente: llama al 01 8000 123 456, visita una sucursal, o consulta en la App / Banca en Línea (www.serfinanza.com) si tienes habilitada la opción de ahorro programado en tu cuenta de ahorros."

const ActualizacionDatosReply = `Puedes actualizar tus datos por estos canales de Serfinanza:

1) App Serfinanza — datos básicos (dirección, celular, correo), actualización inmediata: menú Mi Perfil → Actualizar datos de contacto → confirma con OTP. No sirve para cambios laborales, financieros o de identificación con documentos.

2) Portal web — www.serfinanza.com, Banca en Línea, 24/7: datos de contacto y formulario para trámites con documentación (carga en Mis Trámites; respuesta hasta 3 días hábiles).

3) WhatsApp — +57 300 987 6543: escribe ACTUALIZAR; actualiza dirección, correo y teléfono. Trámites complejos te pasan con un asesor.

4) Call Center — 01 8000 123 456 o (601) 321-0000 (Bogotá), lun–sáb 7:00 a.m.–9:00 p.m.: teléfono y correo al instante; otros datos con ticket (1–2 días hábiles).

5) Sucursal — 45 sucursales a nivel nacional: canal para TODAS las actualizaciones, incluso con documentos originales; la mayoría se resuelve el mismo día.

¿Qué dato quieres cambiar? Te indico el canal más rápido.`

const ExtractoGuideReply = "Para generar tu extracto del mes en la App Serfinanza: (1) ingresa y elige tu producto (tarjeta o crédito), (2) toca Documentos o Extractos, (3) selecciona mes y año (hasta 24 meses atrás), (4) toca Generar y descarga el PDF. También puedes usar Banca en Línea en www.serfinanza.com o WhatsApp escribiendo EXTRACTO al +57 300 987 6543.\n\nPara leerlo, revisa: período de facturación, saldo total y pago mínimo, movimientos del mes, intereses y fecha límite de pago. Si no reconoces un cobro, reclama en máximo 30 días por Call Center (01 8000 123 456) o en la App en Atención → Reclamaciones."

const PortalAccessReply = "Para ingresar a Serfinanza Virtual (Banca en Línea): entra a www.serfinanza.com, sección Banca en Línea, con tu usuario y contraseña. Si aún no tienes acceso, regístrate primero en la App Serfinanza (descárgala en App Store o Google Play) con tu cédula y el celular registrado en el banco."

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
	retrieveQ, _ := ResolveQuery(query)
	return retrieveQ
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
