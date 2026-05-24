package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/javierg/hackathon-bqia/internal/domain"
	"github.com/javierg/hackathon-bqia/internal/store"
)

type ExecutionHandler struct {
	execStore *store.ExecutionStore
	issueStore *store.IssueStore
	seoStore *store.SEOStore
	approvalStore *store.ApprovalStore
}

func NewExecutionHandler(db *sql.DB) *ExecutionHandler {
	return &ExecutionHandler{
		execStore: store.NewExecutionStore(db),
		issueStore: store.NewIssueStore(db),
		seoStore: store.NewSEOStore(db),
		approvalStore: store.NewApprovalStore(db),
	}
}

func (h *ExecutionHandler) CreateExecution(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	if req.JobName == "" {
		h.error(w, http.StatusBadRequest, "MISSING_JOB_NAME", "jobName es requerido")
		return
	}
	exec, err := h.execStore.Create(req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}
	h.ok(w, exec)
}

func (h *ExecutionHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "MISSING_ID", "id es requerido")
		return
	}
	exec, err := h.execStore.GetByID(id)
	if err != nil {
		h.error(w, http.StatusNotFound, "NOT_FOUND", "ejecución no encontrada")
		return
	}
	h.ok(w, exec)
}

func (h *ExecutionHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	execs, err := h.execStore.List(limit, offset)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	count, _ := h.execStore.Count()
	h.ok(w, map[string]interface{}{"items": execs, "total": count, "limit": limit, "offset": offset})
}

func (h *ExecutionHandler) UpdateExecution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "MISSING_ID", "id es requerido")
		return
	}
	var req domain.UpdateExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	exec, err := h.execStore.Update(id, req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}
	h.ok(w, exec)
}

func (h *ExecutionHandler) DeleteExecution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "MISSING_ID", "id es requerido")
		return
	}
	if err := h.execStore.Delete(id); err != nil {
		h.error(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]bool{"deleted": true})
}

func (h *ExecutionHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	if req.ExecutionID == "" || req.IssueType == "" || req.Description == "" {
		h.error(w, http.StatusBadRequest, "MISSING_FIELDS", "executionId, issueType y description son requeridos")
		return
	}
	issue, err := h.issueStore.Create(req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}
	h.ok(w, issue)
}

func (h *ExecutionHandler) BulkCreateIssues(w http.ResponseWriter, r *http.Request) {
	var req domain.BulkIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	issues, err := h.issueStore.BulkCreate(req.Issues)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "BULK_CREATE_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]interface{}{"items": issues, "count": len(issues)})
}

func (h *ExecutionHandler) ListIssuesByExecution(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionID")
	if executionID == "" {
		h.error(w, http.StatusBadRequest, "MISSING_EXECUTION_ID", "executionID es requerido")
		return
	}
	issues, err := h.issueStore.ListByExecution(executionID)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]interface{}{"items": issues, "count": len(issues)})
}

func (h *ExecutionHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "MISSING_ID", "id es requerido")
		return
	}
	var req domain.UpdateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	issue, err := h.issueStore.Update(id, req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}
	h.ok(w, issue)
}

func (h *ExecutionHandler) CreateSEO(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSEORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	if req.ExecutionID == "" {
		h.error(w, http.StatusBadRequest, "MISSING_EXECUTION_ID", "executionId es requerido")
		return
	}
	seo, err := h.seoStore.Create(req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}
	h.ok(w, seo)
}

func (h *ExecutionHandler) BulkCreateSEO(w http.ResponseWriter, r *http.Request) {
	var req domain.BulkSEORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	analyses, err := h.seoStore.BulkCreate(req.Analyses)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "BULK_CREATE_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]interface{}{"items": analyses, "count": len(analyses)})
}

func (h *ExecutionHandler) ListSEOByExecution(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionID")
	if executionID == "" {
		h.error(w, http.StatusBadRequest, "MISSING_EXECUTION_ID", "executionID es requerido")
		return
	}
	analyses, err := h.seoStore.ListByExecution(executionID)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]interface{}{"items": analyses, "count": len(analyses)})
}

func (h *ExecutionHandler) CreateApproval(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	if req.ExecutionID == "" || req.ItemType == "" || req.Title == "" || req.RequestedBy == "" {
		h.error(w, http.StatusBadRequest, "MISSING_FIELDS", "executionId, itemType, title y requestedBy son requeridos")
		return
	}
	item, err := h.approvalStore.Create(req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}
	h.ok(w, item)
}

func (h *ExecutionHandler) BulkCreateApprovals(w http.ResponseWriter, r *http.Request) {
	var req domain.BulkApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	items, err := h.approvalStore.BulkCreate(req.Items)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "BULK_CREATE_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]interface{}{"items": items, "count": len(items)})
}

func (h *ExecutionHandler) ListApprovalsByExecution(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionID")
	if executionID == "" {
		h.error(w, http.StatusBadRequest, "MISSING_EXECUTION_ID", "executionID es requerido")
		return
	}
	items, err := h.approvalStore.ListByExecution(executionID)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]interface{}{"items": items, "count": len(items)})
}

func (h *ExecutionHandler) ListPendingApprovals(w http.ResponseWriter, r *http.Request) {
	items, err := h.approvalStore.ListPending()
	if err != nil {
		h.error(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}
	h.ok(w, map[string]interface{}{"items": items, "count": len(items)})
}

func (h *ExecutionHandler) UpdateApproval(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "MISSING_ID", "id es requerido")
		return
	}
	var req domain.UpdateApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}
	item, err := h.approvalStore.Update(id, req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}
	h.ok(w, item)
}

func (h *ExecutionHandler) ok(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
}

func (h *ExecutionHandler) error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(domain.ErrorResponse{
		Error: domain.ErrorDetail{Code: code, Message: message},
	})
}