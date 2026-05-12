package audit_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authtestdb "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/database/postgresql/auth"
	domainaudit "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/audit"
	auditrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/audit"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

var testHelper = authtestdb.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMainWithoutExit(m)
	code := m.Run()
	testHelper.TerminateContainer()
	os.Exit(code)
}

func setupRepository(t *testing.T) (*auditrepo.AuditDatabaseRepository, *sql.DB, func()) {
	t.Helper()
	db, cleanup := testHelper.SetupTestDB(t)
	return auditrepo.NewAuditDatabaseRepository(db), db, cleanup
}

func TestAuditDatabaseRepository_Save(t *testing.T) {
	tests := []struct {
		name         string
		input        func() *domainaudit.AuditLog
		validateRows func(t *testing.T, db *sql.DB, log *domainaudit.AuditLog)
	}{
		{
			name: "happy path - salva log com todos os campos preenchidos",
			input: func() *domainaudit.AuditLog {
				return domainaudit.NewAuditLog(domainaudit.NewAuditLogInput{
					SubjectID: uuid.New().String(),
					Role:      role.Specialist,
					EventType: domainaudit.EventLoginSuccess,
					IPAddress: "1.2.3.4",
					UserAgent: "go-test",
					Metadata:  map[string]any{"ip_city": "Sao Paulo", "reason": "ok"},
				})
			},
			validateRows: func(t *testing.T, db *sql.DB, log *domainaudit.AuditLog) {
				var (
					subjectID sql.NullString
					roleValue sql.NullString
					ipAddress sql.NullString
					userAgent sql.NullString
					metadata  []byte
				)
				err := db.QueryRowContext(context.Background(),
					"SELECT subject_id, role, ip_address, user_agent, metadata FROM audit_logs WHERE id = $1",
					log.ID,
				).Scan(&subjectID, &roleValue, &ipAddress, &userAgent, &metadata)
				require.NoError(t, err)
				assert.Equal(t, log.SubjectID, subjectID.String)
				assert.Equal(t, role.Specialist.String(), roleValue.String)
				assert.Equal(t, "1.2.3.4", ipAddress.String)
				assert.Equal(t, "go-test", userAgent.String)

				var decoded map[string]any
				require.NoError(t, json.Unmarshal(metadata, &decoded))
				assert.Equal(t, "Sao Paulo", decoded["ip_city"])
				assert.Equal(t, "ok", decoded["reason"])
			},
		},
		{
			name: "happy path - salva log com metadata nil (coluna fica NULL)",
			input: func() *domainaudit.AuditLog {
				return domainaudit.NewAuditLog(domainaudit.NewAuditLogInput{
					SubjectID: uuid.New().String(),
					Role:      role.Specialist,
					EventType: domainaudit.EventPasswordSet,
					IPAddress: "1.2.3.4",
					UserAgent: "go-test",
				})
			},
			validateRows: func(t *testing.T, db *sql.DB, log *domainaudit.AuditLog) {
				var metadata sql.NullString
				err := db.QueryRowContext(context.Background(),
					"SELECT metadata::text FROM audit_logs WHERE id = $1",
					log.ID,
				).Scan(&metadata)
				require.NoError(t, err)
				assert.False(t, metadata.Valid, "metadata should be NULL")
			},
		},
		{
			name: "happy path - salva log sem subject (access_denied anonymous) com campos nulos",
			input: func() *domainaudit.AuditLog {
				return domainaudit.NewAuditLog(domainaudit.NewAuditLogInput{
					EventType: domainaudit.EventAccessDenied,
					IPAddress: "1.2.3.4",
				})
			},
			validateRows: func(t *testing.T, db *sql.DB, log *domainaudit.AuditLog) {
				var (
					subjectID sql.NullString
					roleValue sql.NullString
					userAgent sql.NullString
				)
				err := db.QueryRowContext(context.Background(),
					"SELECT subject_id, role, user_agent FROM audit_logs WHERE id = $1",
					log.ID,
				).Scan(&subjectID, &roleValue, &userAgent)
				require.NoError(t, err)
				assert.False(t, subjectID.Valid)
				assert.False(t, roleValue.Valid)
				assert.False(t, userAgent.Valid)
			},
		},
		{
			name: "happy path - metadata aninhada (nested map) salva como JSON",
			input: func() *domainaudit.AuditLog {
				return domainaudit.NewAuditLog(domainaudit.NewAuditLogInput{
					SubjectID: uuid.New().String(),
					Role:      role.Admin,
					EventType: domainaudit.EventAdminAccessResource,
					Metadata: map[string]any{
						"resource": map[string]any{"type": "specialist", "id": "abc"},
						"actor":    "admin-1",
					},
				})
			},
			validateRows: func(t *testing.T, db *sql.DB, log *domainaudit.AuditLog) {
				var metadata []byte
				err := db.QueryRowContext(context.Background(),
					"SELECT metadata FROM audit_logs WHERE id = $1",
					log.ID,
				).Scan(&metadata)
				require.NoError(t, err)

				var decoded map[string]any
				require.NoError(t, json.Unmarshal(metadata, &decoded))
				resource := decoded["resource"].(map[string]any)
				assert.Equal(t, "specialist", resource["type"])
				assert.Equal(t, "abc", resource["id"])
				assert.Equal(t, "admin-1", decoded["actor"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, db, cleanup := setupRepository(t)
			defer cleanup()

			log := tt.input()
			err := repo.Save(context.Background(), log)
			require.NoError(t, err)

			tt.validateRows(t, db, log)
		})
	}
}
