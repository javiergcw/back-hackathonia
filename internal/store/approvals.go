package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

type ApprovalStore struct {
	db *sql.DB
}

func NewApprovalStore(db *sql.DB) *ApprovalStore {
	return &ApprovalStore{db: db}
}

func (s *ApprovalStore) Create(req domain.CreateApprovalRequest) (*domain.ApprovalItem, error) {
	metadata, _ := json.Marshal(req.Metadata)
	id := generateUUID()
	now := time.Now().UTC().Format(time.RFC3339)

	var item domain.ApprovalItem
	err := s.db.QueryRow(`
		INSERT INTO approval_items (id, execution_id, item_type, title, description, requested_by, requested_at, state, priority, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, execution_id, item_type, title, description, requested_by, requested_at, state, reviewed_by, reviewed_at, review_notes, priority, metadata
	`, id, req.ExecutionID, req.ItemType, req.Title, req.Description, req.RequestedBy, now, domain.ApprovalStatePending, req.Priority, metadata).
		Scan(&item.ID, &item.ExecutionID, &item.ItemType, &item.Title, &item.Description, &item.RequestedBy, &item.RequestedAt, &item.State, &item.ReviewedBy, &item.ReviewedAt, &item.ReviewNotes, &item.Priority, &metadata)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadata, &item.Metadata)
	return &item, nil
}

func (s *ApprovalStore) BulkCreate(reqs []domain.CreateApprovalRequest) ([]domain.ApprovalItem, error) {
	var items []domain.ApprovalItem
	for _, req := range reqs {
		item, err := s.Create(req)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (s *ApprovalStore) GetByID(id string) (*domain.ApprovalItem, error) {
	var item domain.ApprovalItem
	var metadata []byte
	err := s.db.QueryRow(`
		SELECT id, execution_id, item_type, title, description, requested_by, requested_at, state, reviewed_by, reviewed_at, review_notes, priority, metadata
		FROM approval_items WHERE id = $1
	`, id).Scan(&item.ID, &item.ExecutionID, &item.ItemType, &item.Title, &item.Description, &item.RequestedBy, &item.RequestedAt, &item.State, &item.ReviewedBy, &item.ReviewedAt, &item.ReviewNotes, &item.Priority, &metadata)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadata, &item.Metadata)
	return &item, nil
}

func (s *ApprovalStore) ListByExecution(executionID string) ([]domain.ApprovalItem, error) {
	rows, err := s.db.Query(`
		SELECT id, execution_id, item_type, title, description, requested_by, requested_at, state, reviewed_by, reviewed_at, review_notes, priority, metadata
		FROM approval_items WHERE execution_id = $1 ORDER BY requested_at DESC
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ApprovalItem
	for rows.Next() {
		var item domain.ApprovalItem
		var metadata []byte
		if err := rows.Scan(&item.ID, &item.ExecutionID, &item.ItemType, &item.Title, &item.Description, &item.RequestedBy, &item.RequestedAt, &item.State, &item.ReviewedBy, &item.ReviewedAt, &item.ReviewNotes, &item.Priority, &metadata); err != nil {
			return nil, err
		}
		json.Unmarshal(metadata, &item.Metadata)
		items = append(items, item)
	}
	return items, nil
}

func (s *ApprovalStore) ListPending() ([]domain.ApprovalItem, error) {
	rows, err := s.db.Query(`
		SELECT id, execution_id, item_type, title, description, requested_by, requested_at, state, reviewed_by, reviewed_at, review_notes, priority, metadata
		FROM approval_items WHERE state = $1 ORDER BY requested_at DESC
	`, domain.ApprovalStatePending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ApprovalItem{}
	for rows.Next() {
		var item domain.ApprovalItem
		var metadata []byte
		if err := rows.Scan(&item.ID, &item.ExecutionID, &item.ItemType, &item.Title, &item.Description, &item.RequestedBy, &item.RequestedAt, &item.State, &item.ReviewedBy, &item.ReviewedAt, &item.ReviewNotes, &item.Priority, &metadata); err != nil {
			return nil, err
		}
		json.Unmarshal(metadata, &item.Metadata)
		items = append(items, item)
	}
	return items, nil
}

func (s *ApprovalStore) Update(id string, req domain.UpdateApprovalRequest) (*domain.ApprovalItem, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if req.State != "" {
		_, err := s.db.Exec(`UPDATE approval_items SET state = $1, reviewed_by = $2, reviewed_at = $3, review_notes = $4 WHERE id = $5`,
			req.State, req.ReviewedBy, now, req.ReviewNotes, id)
		if err != nil {
			return nil, err
		}
	}
	return s.GetByID(id)
}

func (s *ApprovalStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM approval_items WHERE id = $1`, id)
	return err
}