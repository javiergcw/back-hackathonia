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
	offTopicPattern  = regexp.MustCompile(`(?i)(presidente|capital de|clima|futbol|f[uú]tbol|receta|poema|chiste|historia de|quien invent[oó]|gpt|chatgpt|inteligencia artificial|pel[ií]cula|m[uú]sica|tiktok|instagram)`)
	greetingPattern  = regexp.MustCompile(`(?i)^(hola|buenos d[ií]as|buenas tardes|buenas noches|hey|hi|saludos)[\s!.?]*$`)
	bankHintPattern  = regexp.MustCompile(`(?i)(serfinanza|banco|cdt|supercdt|tarjeta|cr[eé]dito|cuenta|ahorro|app|extracto|fogaf[ií]n|cupo|pago|debito|d[eé]bito|inversi[oó]n|simul|bloque|otp|sucursal|call center)`)
)

const GreetingReply = "Hola, soy tu asesor de Banco Serfinanza. ¿En qué producto o trámite te puedo ayudar hoy?"

const OutOfScopeReply = "Solo puedo ayudarte con productos, trámites y servicios de Banco Serfinanza. ¿Tienes alguna consulta sobre tu cuenta, tarjeta, CDT, crédito o la App?"

func IsGreeting(query string) bool {
	return greetingPattern.MatchString(strings.TrimSpace(query))
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
