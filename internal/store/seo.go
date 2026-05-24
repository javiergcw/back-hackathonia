package store

import (
	"database/sql"
	"encoding/json"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

type SEOStore struct {
	db *sql.DB
}

func NewSEOStore(db *sql.DB) *SEOStore {
	return &SEOStore{db: db}
}

func (s *SEOStore) Create(req domain.CreateSEORequest) (*domain.SEOAnalysis, error) {
	metadata, _ := json.Marshal(req.Metadata)
	suggestions, _ := json.Marshal(req.Suggestions)
	id := generateUUID()

	var seo domain.SEOAnalysis
	err := s.db.QueryRow(`
		INSERT INTO seo_analysis (id, execution_id, content_hash, title, meta_description, keywords, word_count, readability_score, seo_score, suggestions, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, execution_id, content_hash, title, meta_description, keywords, word_count, readability_score, seo_score, suggestions, created_at, metadata
	`, id, req.ExecutionID, req.ContentHash, req.Title, req.MetaDescription, req.Keywords, req.WordCount, req.ReadabilityScore, req.SEOScore, suggestions, metadata).
		Scan(&seo.ID, &seo.ExecutionID, &seo.ContentHash, &seo.Title, &seo.MetaDescription, &seo.Keywords, &seo.WordCount, &seo.ReadabilityScore, &seo.SEOScore, &suggestions, &seo.CreatedAt, &seo.Metadata)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(suggestions, &seo.Suggestions)
	json.Unmarshal(metadata, &seo.Metadata)
	return &seo, nil
}

func (s *SEOStore) BulkCreate(reqs []domain.CreateSEORequest) ([]domain.SEOAnalysis, error) {
	var analyses []domain.SEOAnalysis
	for _, req := range reqs {
		seo, err := s.Create(req)
		if err != nil {
			return nil, err
		}
		analyses = append(analyses, *seo)
	}
	return analyses, nil
}

func (s *SEOStore) GetByID(id string) (*domain.SEOAnalysis, error) {
	var seo domain.SEOAnalysis
	var metadata, suggestions []byte
	err := s.db.QueryRow(`
		SELECT id, execution_id, content_hash, title, meta_description, keywords, word_count, readability_score, seo_score, suggestions, created_at, metadata
		FROM seo_analysis WHERE id = $1
	`, id).Scan(&seo.ID, &seo.ExecutionID, &seo.ContentHash, &seo.Title, &seo.MetaDescription, &seo.Keywords, &seo.WordCount, &seo.ReadabilityScore, &seo.SEOScore, &suggestions, &seo.CreatedAt, &metadata)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(suggestions, &seo.Suggestions)
	json.Unmarshal(metadata, &seo.Metadata)
	return &seo, nil
}

func (s *SEOStore) ListByExecution(executionID string) ([]domain.SEOAnalysis, error) {
	rows, err := s.db.Query(`
		SELECT id, execution_id, content_hash, title, meta_description, keywords, word_count, readability_score, seo_score, suggestions, created_at, metadata
		FROM seo_analysis WHERE execution_id = $1 ORDER BY created_at DESC
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyses []domain.SEOAnalysis
	for rows.Next() {
		var seo domain.SEOAnalysis
		var metadata, suggestions []byte
		if err := rows.Scan(&seo.ID, &seo.ExecutionID, &seo.ContentHash, &seo.Title, &seo.MetaDescription, &seo.Keywords, &seo.WordCount, &seo.ReadabilityScore, &seo.SEOScore, &suggestions, &seo.CreatedAt, &metadata); err != nil {
			return nil, err
		}
		json.Unmarshal(suggestions, &seo.Suggestions)
		json.Unmarshal(metadata, &seo.Metadata)
		analyses = append(analyses, seo)
	}
	return analyses, nil
}

func (s *SEOStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM seo_analysis WHERE id = $1`, id)
	return err
}