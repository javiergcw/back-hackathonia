package domain

type AskRequest struct {
	Question  string `json:"question"`
	Channel   string `json:"channel"`
	SessionID string `json:"sessionId"`
}

type AskResponse struct {
	Answer     string     `json:"answer"`
	Citations  []Citation `json:"citations"`
	Grounded   bool       `json:"grounded"`
	SessionID  string     `json:"sessionId"`
}

type Citation struct {
	Doc     string `json:"doc"`
	Seccion string `json:"seccion"`
}

type SimulateCDTRequest struct {
	Monto     int `json:"monto"`
	PlazoDias int `json:"plazoDias"`
}

type SimulateCDTResponse struct {
	Monto            int       `json:"monto"`
	PlazoDias        int       `json:"plazoDias"`
	TasaEA           float64   `json:"tasaEA"`
	InteresBruto     int       `json:"interesBruto"`
	RetencionFuente   int       `json:"retencionFuente"`
	InteresNeto      int       `json:"interesNeto"`
	TotalVencimiento int       `json:"totalVencimiento"`
	Citation         Citation  `json:"citation"`
}

type RecommendRequest struct {
	ProfileID string `json:"profileId"`
}

type RecommendResponse struct {
	Recomendacion string     `json:"recomendacion"`
	Producto      string     `json:"producto"`
	Accion        string     `json:"accion"`
	Citations     []Citation `json:"citations"`
}

type Chunk struct {
	ID        string   `json:"id"`
	Doc       string   `json:"doc"`
	Seccion   string   `json:"seccion"`
	Tags      []string `json:"tags"`
	Contenido string   `json:"contenido"`
}

type Profile struct {
	ID           string   `json:"id"`
	Nombre       string   `json:"nombre"`
	Productos    []string `json:"productos"`
	Tarjeta      *string  `json:"tarjeta"`
	MesesTarjeta int      `json:"meses_tarjeta"`
}

type WhatsAppPayload struct {
	Event     string `json:"event"`
	Instance  string `json:"instance"`
	Phone     string `json:"phone"`
	Message   string `json:"message"`
	ProfileID string `json:"profileId"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}