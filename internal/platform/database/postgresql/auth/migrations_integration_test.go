package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authtestdb "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/database/postgresql/auth"
)

var testHelper = authtestdb.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMain(m)
}

func TestAuthMigrations_Up_CreatesAllTables(t *testing.T) {
	db, cleanup := testHelper.SetupTestDB(t)
	defer cleanup()

	expectedTables := []string{"credentials", "sessions", "admins", "audit_logs"}

	for _, table := range expectedTables {
		t.Run("table_"+table+"_exists", func(t *testing.T) {
			var exists bool
			err := db.QueryRow(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public' AND table_name = $1
				)
			`, table).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "expected table %q to exist after migrations", table)
		})
	}
}

func TestAuthMigrations_Up_SeedsPlatformAdmin(t *testing.T) {
	db, cleanup := testHelper.SetupTestDB(t)
	defer cleanup()

	var (
		id      string
		name    string
		email   string
		subRole string
		status  string
	)

	err := db.QueryRow(`
		SELECT id, name, email, sub_role, status
		FROM admins
		WHERE email = $1
	`, "admin@healing.local").Scan(&id, &name, &email, &subRole, &status)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-0000-0000-000000000001", id)
	assert.Equal(t, "Platform Admin", name)
	assert.Equal(t, "admin@healing.local", email)
	assert.Equal(t, "admin", subRole)
	assert.Equal(t, "active", status)
}

func TestAuthMigrations_CredentialsIndexes_Exist(t *testing.T) {
	db, cleanup := testHelper.SetupTestDB(t)
	defer cleanup()

	expectedIndexes := []string{
		"idx_credentials_email_provider_role",
		"idx_credentials_subject",
		"idx_credentials_provider_user",
		"idx_sessions_subject",
		"idx_sessions_hash",
		"idx_sessions_expires",
		"idx_admins_email",
		"idx_admins_sub_role",
		"idx_audit_subject",
		"idx_audit_event_type",
		"idx_audit_created_at",
	}

	for _, idx := range expectedIndexes {
		t.Run("index_"+idx+"_exists", func(t *testing.T) {
			var exists bool
			err := db.QueryRow(`
				SELECT EXISTS (
					SELECT FROM pg_indexes
					WHERE schemaname = 'public' AND indexname = $1
				)
			`, idx).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "expected index %q to exist after migrations", idx)
		})
	}
}
