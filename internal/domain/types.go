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

type ExecutionStatus string

const (
	ExecutionStatusPending     ExecutionStatus = "pending"
	ExecutionStatusInProgress  ExecutionStatus = "in_progress"
	ExecutionStatusCompleted   ExecutionStatus = "completed"
	ExecutionStatusFailed      ExecutionStatus = "failed"
	ExecutionStatusCancelled   ExecutionStatus = "cancelled"
)

type Execution struct {
	ID            string            `json:"id"`
	JobName       string            `json:"jobName"`
	CampaignName  string            `json:"campaignName,omitempty"`
	Status        ExecutionStatus   `json:"status"`
	TotalItems    int               `json:"totalItems"`
	ProcessedItems int              `json:"processedItems"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
	CreatedBy     string            `json:"createdBy,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type IssueSeverity string

const (
	IssueSeverityLow      IssueSeverity = "low"
	IssueSeverityMedium   IssueSeverity = "medium"
	IssueSeverityHigh     IssueSeverity = "high"
	IssueSeverityCritical IssueSeverity = "critical"
)

type Issue struct {
	ID          string                  `json:"id"`
	ExecutionID string                  `json:"executionId"`
	IssueType   string                  `json:"issueType"`
	Severity    IssueSeverity           `json:"severity"`
	Description string                  `json:"description"`
	EntityID    string                  `json:"entityId,omitempty"`
	CreatedAt   string                  `json:"createdAt"`
	Resolved    bool                    `json:"resolved"`
	ResolvedAt  string                  `json:"resolvedAt,omitempty"`
	ResolvedBy  string                  `json:"resolvedBy,omitempty"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
}

type SEOAnalysis struct {
	ID                 string                  `json:"id"`
	ExecutionID        string                  `json:"executionId"`
	ContentHash        string                  `json:"contentHash,omitempty"`
	Title              string                  `json:"title,omitempty"`
	MetaDescription   string                  `json:"metaDescription,omitempty"`
	Keywords          string                  `json:"keywords,omitempty"`
	WordCount         int                     `json:"wordCount"`
	ReadabilityScore  float64                 `json:"readabilityScore"`
	SEOScore          float64                 `json:"seoScore"`
	Suggestions       []string                `json:"suggestions,omitempty"`
	CreatedAt         string                  `json:"createdAt"`
	Metadata          map[string]interface{}  `json:"metadata,omitempty"`
}

type ApprovalState string

const (
	ApprovalStatePending   ApprovalState = "pending"
	ApprovalStateApproved ApprovalState = "approved"
	ApprovalStateRejected ApprovalState = "rejected"
)

type ApprovalItem struct {
	ID           string                  `json:"id"`
	ExecutionID  string                  `json:"executionId"`
	ItemType     string                  `json:"itemType"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description,omitempty"`
	RequestedBy  string                  `json:"requestedBy"`
	RequestedAt  string                  `json:"requestedAt"`
	State        ApprovalState          `json:"state"`
	ReviewedBy   string                  `json:"reviewedBy,omitempty"`
	ReviewedAt   string                  `json:"reviewedAt,omitempty"`
	ReviewNotes  string                  `json:"reviewNotes,omitempty"`
	Priority     string                  `json:"priority,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type CreateExecutionRequest struct {
	JobName      string `json:"jobName"`
	CampaignName string `json:"campaignName,omitempty"`
	CreatedBy    string `json:"createdBy,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateExecutionRequest struct {
	Status        ExecutionStatus `json:"status,omitempty"`
	TotalItems    *int           `json:"totalItems,omitempty"`
	ProcessedItems *int          `json:"processedItems,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type BulkIssueRequest struct {
	Issues []CreateIssueRequest `json:"issues"`
}

type CreateIssueRequest struct {
	ExecutionID string            `json:"executionId"`
	IssueType   string            `json:"issueType"`
	Severity    IssueSeverity     `json:"severity"`
	Description string            `json:"description"`
	EntityID    string            `json:"entityId,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateIssueRequest struct {
	Resolved   *bool   `json:"resolved,omitempty"`
	ResolvedBy string  `json:"resolvedBy,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type BulkSEORequest struct {
	Analyses []CreateSEORequest `json:"analyses"`
}

type CreateSEORequest struct {
	ExecutionID       string                  `json:"executionId"`
	ContentHash       string                  `json:"contentHash,omitempty"`
	Title             string                  `json:"title,omitempty"`
	MetaDescription  string                  `json:"metaDescription,omitempty"`
	Keywords         string                  `json:"keywords,omitempty"`
	WordCount        int                     `json:"wordCount"`
	ReadabilityScore float64                 `json:"readabilityScore"`
	SEOScore         float64                 `json:"seoScore"`
	Suggestions      []string                `json:"suggestions,omitempty"`
	Metadata         map[string]interface{}  `json:"metadata,omitempty"`
}

type BulkApprovalRequest struct {
	Items []CreateApprovalRequest `json:"items"`
}

type CreateApprovalRequest struct {
	ExecutionID  string `json:"executionId"`
	ItemType     string `json:"itemType"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	RequestedBy  string `json:"requestedBy"`
	Priority     string `json:"priority,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateApprovalRequest struct {
	State       ApprovalState `json:"state,omitempty"`
	ReviewedBy  string        `json:"reviewedBy,omitempty"`
	ReviewNotes string        `json:"reviewNotes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}