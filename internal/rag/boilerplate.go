package rag

import "strings"

// IsBoilerplate detecta fragmentos legales, pies de página o texto de relleno del PDF.
func IsBoilerplate(content string) bool {
	lower := strings.ToLower(content)
	markers := []string{
		"superintendencia financiera",
		"calle 72 no. 10-07",
		"vigilado superintendencia",
		"entrenamientocondiciones sujetas",
		"carácter informativo y de entrenamiento",
		"caracter informativo y de entrenamiento",
	}
	hits := 0
	for _, m := range markers {
		if strings.Contains(lower, m) {
			hits++
		}
	}
	if hits >= 1 {
		return true
	}
	// Pie repetido con línea de atención duplicada
	if strings.Count(lower, "www.serfinanza.com") >= 2 && strings.Contains(lower, "8000 123 456") {
		return true
	}
	if strings.Contains(lower, "sarlaft") && strings.Contains(lower, "bogotá") {
		return true
	}
	if strings.HasPrefix(strings.TrimSpace(content), "D.C.") || strings.HasPrefix(strings.TrimSpace(content), "d.c.") {
		return true
	}
	if strings.Contains(lower, "bloqueados por sarlaft") {
		return true
	}
	return false
}
