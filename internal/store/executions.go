package store

import (
	"database/sql"
	"encoding/json"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

type ExecutionStore struct {
	db *sql.DB
}

func NewExecutionStore(db *sql.DB) *ExecutionStore {
	return &ExecutionStore{db: db}
}

func (s *ExecutionStore) Create(req domain.CreateExecutionRequest) (*domain.Execution, error) {
	metadata, _ := json.Marshal(req.Metadata)
	id := generateUUID()

	var exec domain.Execution
	err := s.db.QueryRow(`
		INSERT INTO executions (id, job_name, campaign_name, status, created_by, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, job_name, campaign_name, status, total_items, processed_items, created_at, updated_at, created_by, metadata
	`, id, req.JobName, req.CampaignName, domain.ExecutionStatusInProgress, req.CreatedBy, metadata).
		Scan(&exec.ID, &exec.JobName, &exec.CampaignName, &exec.Status, &exec.TotalItems, &exec.ProcessedItems, &exec.CreatedAt, &exec.UpdatedAt, &exec.CreatedBy, &exec.Metadata)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

func (s *ExecutionStore) GetByID(id string) (*domain.Execution, error) {
	var exec domain.Execution
	var metadata []byte
	err := s.db.QueryRow(`
		SELECT id, job_name, campaign_name, status, total_items, processed_items, created_at, updated_at, created_by, metadata
		FROM executions WHERE id = $1
	`, id).Scan(&exec.ID, &exec.JobName, &exec.CampaignName, &exec.Status, &exec.TotalItems, &exec.ProcessedItems, &exec.CreatedAt, &exec.UpdatedAt, &exec.CreatedBy, &metadata)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadata, &exec.Metadata)
	return &exec, nil
}

func (s *ExecutionStore) List(limit, offset int) ([]domain.Execution, error) {
	rows, err := s.db.Query(`
		SELECT id, job_name, campaign_name, status, total_items, processed_items, created_at, updated_at, created_by, metadata
		FROM executions ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []domain.Execution
	for rows.Next() {
		var exec domain.Execution
		var metadata []byte
		if err := rows.Scan(&exec.ID, &exec.JobName, &exec.CampaignName, &exec.Status, &exec.TotalItems, &exec.ProcessedItems, &exec.CreatedAt, &exec.UpdatedAt, &exec.CreatedBy, &metadata); err != nil {
			return nil, err
		}
		json.Unmarshal(metadata, &exec.Metadata)
		executions = append(executions, exec)
	}
	return executions, nil
}

func (s *ExecutionStore) Update(id string, req domain.UpdateExecutionRequest) (*domain.Execution, error) {
	if req.Status != "" {
		_, err := s.db.Exec(`UPDATE executions SET status = $1, updated_at = NOW() WHERE id = $2`, req.Status, id)
		if err != nil {
			return nil, err
		}
	}
	if req.TotalItems != nil {
		_, err := s.db.Exec(`UPDATE executions SET total_items = $1, updated_at = NOW() WHERE id = $2`, *req.TotalItems, id)
		if err != nil {
			return nil, err
		}
	}
	if req.ProcessedItems != nil {
		_, err := s.db.Exec(`UPDATE executions SET processed_items = $1, updated_at = NOW() WHERE id = $2`, *req.ProcessedItems, id)
		if err != nil {
			return nil, err
		}
	}
	if req.Metadata != nil {
		metadata, _ := json.Marshal(req.Metadata)
		_, err := s.db.Exec(`UPDATE executions SET metadata = $1, updated_at = NOW() WHERE id = $2`, metadata, id)
		if err != nil {
			return nil, err
		}
	}
	return s.GetByID(id)
}

func (s *ExecutionStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM executions WHERE id = $1`, id)
	return err
}

func (s *ExecutionStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM executions`).Scan(&count)
	return count, err
}