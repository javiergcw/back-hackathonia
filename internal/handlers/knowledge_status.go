package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetKnowledgeStatus(w http.ResponseWriter, r *http.Request) {
	chunks := h.ragClient.ListChunks()

	type DocInfo struct {
		Name   string `json:"name"`
		Chunks int    `json:"chunks"`
	}

	docMap := make(map[string]int)
	totalChars := 0
	for _, c := range chunks {
		docMap[c.Doc]++
		totalChars += len(c.Contenido)
	}

	var docs []DocInfo
	for name, count := range docMap {
		docs = append(docs, DocInfo{Name: name, Chunks: count})
	}

	avgChunkSize := 0
	if len(chunks) > 0 {
		avgChunkSize = totalChars / len(chunks)
	}

	resp := map[string]interface{}{
		"totalChunks":  len(chunks),
		"totalDocs":    len(docMap),
		"docs":         docs,
		"avgChunkSize": avgChunkSize,
		"totalTokens":  totalChars / 4,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": resp,
	})
}

func (h *Handler) DeleteKnowledgeByDoc(w http.ResponseWriter, r *http.Request) {
	docName := chi.URLParam(r, "docName")
	if docName == "" {
		h.error(w, http.StatusBadRequest, "MISSING_DOCNAME", "docName es requerido")
		return
	}

	count, err := h.ragClient.RemoveChunksByDoc(docName)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	h.ok(w, map[string]interface{}{
		"deleted": true,
		"count":   count,
	})
}

func (h *Handler) ClearAllKnowledge(w http.ResponseWriter, r *http.Request) {
	count, err := h.ragClient.ClearAllChunks()
	if err != nil {
		h.error(w, http.StatusInternalServerError, "CLEAR_FAILED", err.Error())
		return
	}

	h.ok(w, map[string]interface{}{
		"deleted": true,
		"count":   count,
	})
}