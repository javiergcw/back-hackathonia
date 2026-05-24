-- Migration: 004_approvals.sql
-- Approval workflow items
CREATE TABLE IF NOT EXISTS approval_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    requested_by VARCHAR(255) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    state VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewed_by VARCHAR(255),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,
    priority VARCHAR(20) DEFAULT 'normal',
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_approvals_execution_id ON approval_items(execution_id);
CREATE INDEX idx_approvals_state ON approval_items(state);
CREATE INDEX idx_approvals_requested_at ON approval_items(requested_at);