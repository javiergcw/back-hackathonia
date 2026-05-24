-- Migration: 003_seo_analysis.sql
-- SEO analysis results for content/jobs
CREATE TABLE IF NOT EXISTS seo_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    content_hash VARCHAR(64),
    title VARCHAR(500),
    meta_description TEXT,
    keywords VARCHAR(500),
    word_count INTEGER DEFAULT 0,
    readability_score DECIMAL(5,2),
    seo_score DECIMAL(5,2),
    suggestions TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_seo_execution_id ON seo_analysis(execution_id);
CREATE INDEX idx_seo_content_hash ON seo_analysis(content_hash);