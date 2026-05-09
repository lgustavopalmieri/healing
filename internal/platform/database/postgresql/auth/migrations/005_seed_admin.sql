-- +goose Up
INSERT INTO admins (id, name, email, sub_role, status, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Platform Admin',
    'admin@healing.local',
    'admin',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM admins WHERE email = 'admin@healing.local';
