-- Migration: 002_issues.sql
-- Issues found during job execution
CREATE TABLE IF NOT EXISTS issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    issue_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    description TEXT NOT NULL,
    entity_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(255),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_issues_execution_id ON issues(execution_id);
CREATE INDEX idx_issues_severity ON issues(severity);
CREATE INDEX idx_issues_resolved ON issues(resolved);