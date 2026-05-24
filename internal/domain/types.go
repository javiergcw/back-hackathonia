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

type KnowledgeItem struct {
	ID        string   `json:"id"`
	Doc       string   `json:"doc"`
	Seccion   string   `json:"seccion"`
	Tags      []string `json:"tags"`
	Contenido string   `json:"contenido"`
}

type AddKnowledgeRequest struct {
	Doc       string   `json:"doc"`
	Seccion   string   `json:"seccion"`
	Tags      []string `json:"tags"`
	Contenido string   `json:"contenido"`
}

type UpdateKnowledgeRequest struct {
	Tags      []string `json:"tags,omitempty"`
	Contenido string   `json:"contenido,omitempty"`
}

type KnowledgeListResponse struct {
	Items []KnowledgeItem `json:"items"`
}

type AddKnowledgeResponse struct {
	ID string `json:"id"`
}

type UpdateKnowledgeResponse struct {
	ID     string `json:"id"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type DeleteKnowledgeResponse struct {
	Deleted bool `json:"deleted"`
}

type ReloadKnowledgeResponse struct {
	Count int `json:"count"`
}

type Scope struct {
	StrictMode  bool     `json:"strictMode"`
	ActiveDocs  []string `json:"activeDocs"`
}

type SetScopeRequest struct {
	StrictMode  *bool    `json:"strictMode,omitempty"`
	ActiveDocs  []string `json:"activeDocs,omitempty"`
}

type Role string

const (
	RolePublico     Role = "publico"
	RoleCliente     Role = "cliente"
	RoleAsesor      Role = "asesor"
	RoleCoordinador Role = "coordinador"
)

type User struct {
	ID          string   `json:"id"`
	Nombre      string   `json:"nombre"`
	Phone       string   `json:"phone"`
	Role        Role     `json:"role"`
	ProfileID   string   `json:"profileId"`
	AllowedTags []string `json:"allowedTags"`
}

type IdentifyRequest struct {
	Phone     string `json:"phone,omitempty"`
	ProfileID string `json:"profileId,omitempty"`
}

type IdentifyResponse struct {
	UserID      string   `json:"userId"`
	Nombre      string   `json:"nombre"`
	Role        Role     `json:"role"`
	ProfileID   string   `json:"profileId"`
	AllowedTags []string `json:"allowedTags"`
}

type ClienteRiesgo struct {
	Nombre       string  `json:"nombre"`
	Producto     string  `json:"producto"`
	Monto        int     `json:"monto"`
	Probabilidad float64 `json:"probabilidad"`
	Nivel        string  `json:"nivel"`
}

type AnalyticsMorosidad struct {
	TotalEnRiesgo   int             `json:"totalEnRiesgo"`
	PerdidaEstimada int            `json:"perdidaEstimada"`
	ClientesRiesgo []ClienteRiesgo `json:"clientesRiesgo"`
}

type AnalyticsProyeccion struct {
	EscenarioOptimista int `json:"escenarioOptimista"`
	EscenarioBase      int `json:"escenarioBase"`
	EscenarioPesimista int `json:"escenarioPesimista"`
}

type PreguntaFrecuente struct {
	Pregunta string `json:"pregunta"`
	Conteo   int    `json:"conteo"`
}

type AnalyticsTopPreguntas struct {
	Preguntas []PreguntaFrecuente `json:"preguntas"`
}