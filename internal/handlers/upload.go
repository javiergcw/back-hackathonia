package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/javierg/hackathon-bqia/internal/domain"
	"github.com/javierg/hackathon-bqia/internal/rag"
)

type UploadResponse struct {
	FilesProcessed int             `json:"filesProcessed"`
	ChunksCreated  int             `json:"chunksCreated"`
	ChunksReplaced int             `json:"chunksReplaced"`
	Chunks         []domain.Chunk `json:"chunks"`
	AutoReloaded   bool           `json:"autoReloaded"`
}

type ScanFolderRequest struct {
	FolderPath string `json:"folderPath"`
}

func (h *Handler) UploadKnowledge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_FORM", "no se pudo parsear el formulario")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		h.error(w, http.StatusBadRequest, "NO_FILES", "no se encontraron archivos")
		return
	}

	knowledgeFolder := os.Getenv("KNOWLEDGE_FOLDER")
	if knowledgeFolder == "" {
		knowledgeFolder = "./data/knowledge"
	}

	if err := os.MkdirAll(knowledgeFolder, 0755); err != nil {
		h.error(w, http.StatusInternalServerError, "FOLDER_CREATE_FAILED", "no se pudo crear la carpeta")
		return
	}

	maxChunkSize := 500
	if env := os.Getenv("CHUNK_MAX_SIZE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil && val > 0 {
			maxChunkSize = val
		}
	}

	overlap := maxChunkSize / 7
	if env := os.Getenv("CHUNK_OVERLAP"); env != "" {
		if val, err := strconv.Atoi(env); err == nil && val >= 0 {
			overlap = val
		}
	}

	var allChunks []domain.Chunk
	filesProcessed := 0
	totalReplaced := 0

	for _, fileHeader := range files {
		docName := fileHeader.Filename

		removedCount, _ := h.ragClient.RemoveChunksByDoc(docName)
		totalReplaced += removedCount

		src, err := fileHeader.Open()
		if err != nil {
			continue
		}

		dstPath := filepath.Join(knowledgeFolder, docName)
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			continue
		}

		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			continue
		}

		text, err := rag.ExtractText(dstPath)
		if err != nil {
			continue
		}

		chunks := rag.ChunkText(text, docName, maxChunkSize, overlap)

		for _, chunk := range chunks {
			h.ragClient.AddChunk(chunk)
			allChunks = append(allChunks, chunk)
		}

		filesProcessed++
	}

	h.ok(w, UploadResponse{
		FilesProcessed: filesProcessed,
		ChunksCreated:  len(allChunks),
		ChunksReplaced: totalReplaced,
		Chunks:        allChunks,
		AutoReloaded:   true,
	})
}

func (h *Handler) ScanFolderKnowledge(w http.ResponseWriter, r *http.Request) {
	var req ScanFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "INVALID_JSON", "cuerpo JSON inválido")
		return
	}

	if req.FolderPath == "" {
		h.error(w, http.StatusBadRequest, "MISSING_FOLDER", "folderPath es requerido")
		return
	}

	entries, err := os.ReadDir(req.FolderPath)
	if err != nil {
		h.error(w, http.StatusNotFound, "FOLDER_NOT_FOUND", "carpeta no encontrada")
		return
	}

	knowledgeFolder := os.Getenv("KNOWLEDGE_FOLDER")
	if knowledgeFolder == "" {
		knowledgeFolder = "./data/knowledge"
	}
	os.MkdirAll(knowledgeFolder, 0755)

	maxChunkSize := 500
	if env := os.Getenv("CHUNK_MAX_SIZE"); env != "" {
		if val, err := strconv.Atoi(env); err == nil && val > 0 {
			maxChunkSize = val
		}
	}

	overlap := maxChunkSize / 7
	if env := os.Getenv("CHUNK_OVERLAP"); env != "" {
		if val, err := strconv.Atoi(env); err == nil && val >= 0 {
			overlap = val
		}
	}

	var allChunks []domain.Chunk
	filesProcessed := 0
	totalReplaced := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".pdf" && ext != ".txt" && ext != ".md" {
			continue
		}

		docName := entry.Name()

		removedCount, _ := h.ragClient.RemoveChunksByDoc(docName)
		totalReplaced += removedCount

		filePath := filepath.Join(req.FolderPath, entry.Name())

		text, err := rag.ExtractText(filePath)
		if err != nil {
			continue
		}

		chunks := rag.ChunkText(text, docName, maxChunkSize, overlap)

		for _, chunk := range chunks {
			h.ragClient.AddChunk(chunk)
			allChunks = append(allChunks, chunk)
		}

		filesProcessed++
	}

	h.ok(w, UploadResponse{
		FilesProcessed: filesProcessed,
		ChunksCreated:  len(allChunks),
		ChunksReplaced: totalReplaced,
		Chunks:        allChunks,
		AutoReloaded:   true,
	})
}