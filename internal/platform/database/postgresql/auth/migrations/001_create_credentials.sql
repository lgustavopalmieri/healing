-- +goose Up
CREATE TABLE IF NOT EXISTS credentials (
    id VARCHAR(36) PRIMARY KEY,
    subject_id VARCHAR(36) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('specialist', 'patient', 'admin')),
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('password', 'google', 'instagram', 'biometric')),
    provider_user_id VARCHAR(255),
    password_hash VARCHAR(128),
    email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'locked', 'deleted')),
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_credentials_email_provider_role ON credentials(email, provider, role) WHERE status != 'deleted';
CREATE INDEX idx_credentials_subject ON credentials(subject_id, role);
CREATE INDEX idx_credentials_provider_user ON credentials(provider, provider_user_id) WHERE provider_user_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS credentials;
