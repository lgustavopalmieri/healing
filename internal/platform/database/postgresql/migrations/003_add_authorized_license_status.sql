-- +goose Up
ALTER TABLE specialists DROP CONSTRAINT IF EXISTS specialists_status_check;
ALTER TABLE specialists ADD CONSTRAINT specialists_status_check 
CHECK (status IN ('pending', 'authorized_license', 'active', 'unavailable', 'deleted', 'banned'));

-- +goose Down
ALTER TABLE specialists DROP CONSTRAINT IF EXISTS specialists_status_check;
ALTER TABLE specialists ADD CONSTRAINT specialists_status_check 
CHECK (status IN ('pending', 'active', 'unavailable', 'deleted', 'banned'));
