package rag

import (
	"fmt"
	"strings"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

func ProfileContext(profile *domain.Profile) string {
	if profile == nil {
		return ""
	}

	productos := formatProductos(profile.Productos)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("- Nombre: %s\n", profile.Nombre))
	sb.WriteString(fmt.Sprintf("- Productos activos: %s\n", productos))
	if profile.Tarjeta != nil {
		sb.WriteString(fmt.Sprintf("- Tarjeta: %s (%d meses)\n", *profile.Tarjeta, profile.MesesTarjeta))
	}
	return sb.String()
}

func ProfileBoostTags(profile *domain.Profile) []string {
	if profile == nil {
		return nil
	}

	var tags []string
	for _, producto := range profile.Productos {
		switch producto {
		case "cuenta_ahorros":
			tags = append(tags, "cdt", "supercdt", "inversion", "ahorro")
		case "tarjeta_clasica", "tarjeta_gold", "tarjeta_platinum":
			tags = append(tags, "tarjeta", "cupo", "pago")
			if profile.MesesTarjeta >= 6 {
				tags = append(tags, "aumento", "bloquear")
			}
		case "credito_consumo":
			tags = append(tags, "debito automatico", "pago", "extracto", "credito")
		}
	}
	return tags
}

func ProactiveHint(profile *domain.Profile, query string) string {
	if profile == nil {
		return ""
	}

	q := strings.ToLower(query)
	switch {
	case HasProduct(profile.Productos, "cuenta_ahorros") && !HasAnyProduct(profile.Productos, "cdt"):
		if mentionsAny(q, "ahorr", "dinero", "invert", "cdt", "rentabil") {
			return "Si te interesa, puedo contarte más sobre el *superCDT* para hacer crecer tu plata con rentabilidad fija."
		}
	case HasProduct(profile.Productos, "tarjeta_clasica") && profile.MesesTarjeta >= 6:
		if mentionsAny(q, "tarjeta", "cupo", "limite") {
			return "También puedes solicitar un *aumento de cupo* desde la App si llevas buen historial de pagos."
		}
	case HasProduct(profile.Productos, "credito_consumo"):
		if mentionsAny(q, "credito", "pago", "cuota", "debito") {
			return "Te recomiendo activar el *débito automático* para no perder ningún pago."
		}
	}
	return ""
}

func formatProductos(productos []string) string {
	labels := map[string]string{
		"cuenta_ahorros":  "cuenta de ahorros",
		"tarjeta_clasica": "tarjeta clásica",
		"tarjeta_gold":    "tarjeta gold",
		"tarjeta_platinum": "tarjeta platinum",
		"credito_consumo": "crédito de consumo",
		"cdt":             "CDT",
	}
	var parts []string
	for _, p := range productos {
		if label, ok := labels[p]; ok {
			parts = append(parts, label)
		} else {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return "sin productos registrados"
	}
	return strings.Join(parts, ", ")
}

func HasProduct(productos []string, product string) bool {
	for _, p := range productos {
		if p == product {
			return true
		}
	}
	return false
}

func HasAnyProduct(productos []string, targets ...string) bool {
	for _, p := range productos {
		for _, target := range targets {
			if p == target {
				return true
			}
		}
	}
	return false
}

func mentionsAny(text string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}
