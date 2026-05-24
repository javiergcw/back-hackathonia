package rag

import "strings"

var querySynonyms = map[string][]string{
	"cdt":           {"supercdt", "inversion", "invertir", "rentabilidad", "fogafin"},
	"supercdt":      {"cdt", "inversion", "rentabilidad"},
	"tarjeta":       {"cupo", "credito", "mastercard", "plastico", "tarjetas"},
	"tarjetas":      {"tarjeta", "cupo", "credito"},
	"bloquear":      {"robo", "perdida", "bloqueo", "reportar", "robar", "robado", "robaron"},
	"robo":          {"robar", "robado", "robaron", "hurto", "extraviada", "perdida"},
	"reportar":      {"reporte", "reportarlo", "denunciar"},
	"extracto":      {"factura", "estado", "movimientos", "generar", "pdf", "leer"},
	"app":           {"aplicacion", "celular", "movil", "serfinanza virtual"},
	"registro":      {"registrarme", "crear cuenta", "nuevo usuario"},
	"ingresar":      {"ingreso", "entrar", "acceder", "login", "iniciar sesion"},
	"virtual":       {"personas", "portal", "banca", "linea", "web"},
	"personas":      {"virtual", "banca", "portal"},
	"radicacion":    {"radicar", "solicitud", "tramite"},
	"ahorro":        {"plan", "programado", "meta"},
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
