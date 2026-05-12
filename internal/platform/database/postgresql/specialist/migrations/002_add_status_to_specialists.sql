-- +goose Up
ALTER TABLE specialists 
ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending' 
CHECK (status IN ('pending', 'active', 'unavailable', 'deleted', 'banned'));

CREATE INDEX IF NOT EXISTS idx_specialists_status ON specialists(status);

-- +goose Down
DROP INDEX IF EXISTS idx_specialists_status;
ALTER TABLE specialists DROP COLUMN IF EXISTS status;
