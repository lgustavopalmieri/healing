-- +goose Up
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    subject_id VARCHAR(36),
    role VARCHAR(20),
    event_type VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_subject ON audit_logs(subject_id, role, created_at DESC);
CREATE INDEX idx_audit_event_type ON audit_logs(event_type, created_at DESC);
CREATE INDEX idx_audit_created_at ON audit_logs(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
