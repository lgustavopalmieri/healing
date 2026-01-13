-- +goose Up
CREATE TABLE IF NOT EXISTS specialists (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    phone VARCHAR(20),
    specialty VARCHAR(100) NOT NULL,
    license_number VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    keywords TEXT[] DEFAULT '{}',
    agreed_to_share BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_specialists_email ON specialists(email);
CREATE INDEX IF NOT EXISTS idx_specialists_license_number ON specialists(license_number);
CREATE INDEX IF NOT EXISTS idx_specialists_specialty ON specialists(specialty);
CREATE INDEX IF NOT EXISTS idx_specialists_keywords ON specialists USING GIN(keywords);

-- +goose Down
DROP TABLE IF EXISTS specialists;