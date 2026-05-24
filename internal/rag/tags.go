package rag

import (
	"strings"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

// chunkAllowedForTags decide si un chunk es visible según los tags del usuario.
// publico/general y * abren todo el corpus; tarjetas/cdt/cuentas también por nombre de documento.
func chunkAllowedForTags(allowedTags []string, chunk domain.Chunk) bool {
	for _, allowed := range allowedTags {
		if allowed == "*" || allowed == "publico" || allowed == "general" {
			return true
		}
	}

	for _, chunkTag := range chunk.Tags {
		for _, allowed := range allowedTags {
			if tagsMatch(allowed, chunkTag) {
				return true
			}
		}
	}

	doc := strings.ToLower(chunk.Doc)
	for _, allowed := range allowedTags {
		switch allowed {
		case "tarjetas", "tarjeta":
			if strings.Contains(doc, "tarjeta") {
				return true
			}
		case "cdt", "supercdt":
			if strings.Contains(doc, "cdt") {
				return true
			}
		case "cuentas", "cuenta", "ahorros":
			if strings.Contains(doc, "cuenta") || strings.Contains(doc, "ahorro") {
				return true
			}
		}
	}

	return false
}

func tagsMatch(allowed, chunkTag string) bool {
	if allowed == chunkTag {
		return true
	}
	a := strings.TrimSuffix(strings.ToLower(allowed), "s")
	b := strings.TrimSuffix(strings.ToLower(chunkTag), "s")
	return a == b
}
