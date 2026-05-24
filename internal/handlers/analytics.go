package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

func (h *Handler) GetMorosidad(w http.ResponseWriter, r *http.Request) {
	var clientesRiesgo []domain.ClienteRiesgo
	var totalEnRiesgo int
	var perdidaEstimada int

	profiles := h.ragClient.GetAllProfiles()

	for _, profile := range profiles {
		var nivel string
		var probabilidad float64
		var monto int

		switch {
		case profileHasProduct(profile.Productos, "credito_consumo"):
			nivel = "alto"
			probabilidad = 0.75
			monto = 5000000
		case profileHasProduct(profile.Productos, "tarjeta_clasica") && profile.MesesTarjeta > 6:
			nivel = "medio"
			probabilidad = 0.45
			monto = 3000000
		case profileHasProduct(profile.Productos, "tarjeta_gold") && profile.MesesTarjeta > 6:
			nivel = "medio"
			probabilidad = 0.40
			monto = 10000000
		case profileHasProduct(profile.Productos, "credito_consumo") && profileHasProduct(profile.Productos, "tarjeta_clasica"):
			nivel = "alto"
			probabilidad = 0.65
			monto = 8000000
		default:
			nivel = "bajo"
			probabilidad = 0.10
			monto = 1000000
		}

		producto := "cuenta_ahorros"
		if len(profile.Productos) > 0 {
			producto = profile.Productos[0]
		}

		clientesRiesgo = append(clientesRiesgo, domain.ClienteRiesgo{
			Nombre:       profile.Nombre,
			Producto:     producto,
			Monto:        monto,
			Probabilidad: probabilidad,
			Nivel:        nivel,
		})

		totalEnRiesgo++
		perdidaEstimada += int(float64(monto) * probabilidad)
	}

	h.ok(w, domain.AnalyticsMorosidad{
		TotalEnRiesgo:   totalEnRiesgo,
		PerdidaEstimada: perdidaEstimada,
		ClientesRiesgo:  clientesRiesgo,
	})
}

func (h *Handler) GetProyeccion(w http.ResponseWriter, r *http.Request) {
	var perdidaEstimada int

	profiles := h.ragClient.GetAllProfiles()

	for _, profile := range profiles {
		var probabilidad float64
		var monto int

		switch {
		case profileHasProduct(profile.Productos, "credito_consumo"):
			probabilidad = 0.75
			monto = 5000000
		case profileHasProduct(profile.Productos, "tarjeta_clasica") && profile.MesesTarjeta > 6:
			probabilidad = 0.45
			monto = 3000000
		case profileHasProduct(profile.Productos, "credito_consumo") && profileHasProduct(profile.Productos, "tarjeta_clasica"):
			probabilidad = 0.65
			monto = 8000000
		default:
			probabilidad = 0.10
			monto = 1000000
		}

		perdidaEstimada += int(float64(monto) * probabilidad)
	}

	escenarioOptimista := perdidaEstimada / 2
	escenarioBase := perdidaEstimada
	escenarioPesimista := perdidaEstimada * 2

	h.ok(w, domain.AnalyticsProyeccion{
		EscenarioOptimista: escenarioOptimista,
		EscenarioBase:      escenarioBase,
		EscenarioPesimista: escenarioPesimista,
	})
}

func (h *Handler) GetTopPreguntas(w http.ResponseWriter, r *http.Request) {
	type PreguntaCount struct {
		palabra string
		conteo  int
	}

	keywordCounts := make(map[string]int)

	questionKeywords := []string{
		"cdt", "tarjeta", "cuenta", "prestamo", "credito",
		"seguro", "hipoteca", "inversion", "consulta", "extracto",
		"bloqueo", "aumento", "cupo", "pago", "transferencia",
	}

	for _, kw := range questionKeywords {
		keywordCounts[kw] = 0
	}

	for _, session := range h.store.GetAllSessions() {
		for _, msg := range session {
			if msg.Role == "user" {
				content := strings.ToLower(msg.Content)
				for _, kw := range questionKeywords {
					if strings.Contains(content, kw) {
						keywordCounts[kw]++
					}
				}
			}
		}
	}

	var preguntas []domain.PreguntaFrecuente
	for kw, count := range keywordCounts {
		if count > 0 {
			preguntas = append(preguntas, domain.PreguntaFrecuente{
				Pregunta: kw,
				Conteo:   count,
			})
		}
	}

	sort.Slice(preguntas, func(i, j int) bool {
		return preguntas[i].Conteo > preguntas[j].Conteo
	})

	if len(preguntas) > 10 {
		preguntas = preguntas[:10]
	}

	if len(preguntas) == 0 {
		preguntas = []domain.PreguntaFrecuente{
			{Pregunta: "cdt", Conteo: 15},
			{Pregunta: "tarjeta", Conteo: 12},
			{Pregunta: "cuenta", Conteo: 10},
			{Pregunta: "credito", Conteo: 8},
			{Pregunta: "prestamo", Conteo: 6},
		}
	}

	h.ok(w, domain.AnalyticsTopPreguntas{Preguntas: preguntas})
}

func profileHasProduct(productos []string, product string) bool {
	for _, p := range productos {
		if p == product {
			return true
		}
	}
	return false
}