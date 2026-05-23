package rag

import "strings"

var querySynonyms = map[string][]string{
	"cdt":           {"supercdt", "inversion", "invertir", "rentabilidad", "fogafin"},
	"supercdt":      {"cdt", "inversion", "rentabilidad"},
	"tarjeta":       {"cupo", "credito", "mastercard", "plastico"},
	"bloquear":      {"robo", "perdida", "bloqueo"},
	"extracto":      {"factura", "estado", "movimientos"},
	"app":           {"aplicacion", "celular", "movil", "serfinanza virtual"},
	"registro":      {"registrarme", "crear cuenta", "nuevo usuario"},
	"actualizar":    {"cambiar", "modificar", "datos"},
	"pago":          {"cuota", "abono", "pagar"},
	"debito":        {"automatico", "domiciliacion"},
	"simular":       {"simulador", "calcular", "cuanto gano"},
}

func expandTerms(terms []string) []string {
	seen := make(map[string]bool)
	var expanded []string

	add := func(term string) {
		term = strings.TrimSpace(term)
		if term == "" || seen[term] {
			return
		}
		seen[term] = true
		expanded = append(expanded, term)
	}

	for _, term := range terms {
		add(term)
		for _, syn := range querySynonyms[term] {
			add(syn)
		}
	}

	return expanded
}
