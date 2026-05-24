package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

type IssueStore struct {
	db *sql.DB
}

func NewIssueStore(db *sql.DB) *IssueStore {
	return &IssueStore{db: db}
}

func (s *IssueStore) Create(req domain.CreateIssueRequest) (*domain.Issue, error) {
	metadata, _ := json.Marshal(req.Metadata)
	id := generateUUID()

	var issue domain.Issue
	var metaBytes []byte
	err := s.db.QueryRow(`
		INSERT INTO issues (id, execution_id, issue_type, severity, description, entity_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, execution_id, issue_type, severity, description, entity_id, created_at, resolved, resolved_at, resolved_by, metadata
	`, id, req.ExecutionID, req.IssueType, req.Severity, req.Description, req.EntityID, metadata).
		Scan(&issue.ID, &issue.ExecutionID, &issue.IssueType, &issue.Severity, &issue.Description, &issue.EntityID, &issue.CreatedAt, &issue.Resolved, &issue.ResolvedAt, &issue.ResolvedBy, &metaBytes)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metaBytes, &issue.Metadata)
	return &issue, nil
}

func (s *IssueStore) BulkCreate(reqs []domain.CreateIssueRequest) ([]domain.Issue, error) {
	issues := []domain.Issue{}
	for _, req := range reqs {
		issue, err := s.Create(req)
		if err != nil {
			return nil, err
		}
		issues = append(issues, *issue)
	}
	return issues, nil
}

func (s *IssueStore) GetByID(id string) (*domain.Issue, error) {
	var issue domain.Issue
	var metadata []byte
	err := s.db.QueryRow(`
		SELECT id, execution_id, issue_type, severity, description, entity_id, created_at, resolved, resolved_at, resolved_by, metadata
		FROM issues WHERE id = $1
	`, id).Scan(&issue.ID, &issue.ExecutionID, &issue.IssueType, &issue.Severity, &issue.Description, &issue.EntityID, &issue.CreatedAt, &issue.Resolved, &issue.ResolvedAt, &issue.ResolvedBy, &metadata)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(metadata, &issue.Metadata)
	return &issue, nil
}

func (s *IssueStore) ListByExecution(executionID string) ([]domain.Issue, error) {
	rows, err := s.db.Query(`
		SELECT id, execution_id, issue_type, severity, description, entity_id, created_at, resolved, resolved_at, resolved_by, metadata
		FROM issues WHERE execution_id = $1 ORDER BY created_at DESC
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := []domain.Issue{}
	for rows.Next() {
		var issue domain.Issue
		var metadata []byte
		if err := rows.Scan(&issue.ID, &issue.ExecutionID, &issue.IssueType, &issue.Severity, &issue.Description, &issue.EntityID, &issue.CreatedAt, &issue.Resolved, &issue.ResolvedAt, &issue.ResolvedBy, &metadata); err != nil {
			return nil, err
		}
		json.Unmarshal(metadata, &issue.Metadata)
		issues = append(issues, issue)
	}
	return issues, nil
}

func (s *IssueStore) Update(id string, req domain.UpdateIssueRequest) (*domain.Issue, error) {
	if req.Resolved != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := s.db.Exec(`UPDATE issues SET resolved = $1, resolved_at = $2, resolved_by = $3 WHERE id = $4`, *req.Resolved, now, req.ResolvedBy, id)
		if err != nil {
			return nil, err
		}
	}
	return s.GetByID(id)
}

func (s *IssueStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM issues WHERE id = $1`, id)
	return err
}