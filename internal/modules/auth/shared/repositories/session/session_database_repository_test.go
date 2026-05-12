package session_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authtestdb "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/database/postgresql/auth"
	domainsession "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	sessionrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/session"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

var testHelper = authtestdb.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMainWithoutExit(m)
	code := m.Run()
	testHelper.TerminateContainer()
	os.Exit(code)
}

func sessionFactory(overrides ...func(*domainsession.Session)) *domainsession.Session {
	s := domainsession.NewSession(domainsession.NewSessionInput{
		SubjectID:        uuid.New().String(),
		Role:             role.Specialist,
		RefreshTokenHash: uuid.New().String(),
		DeviceInfo:       "web",
		IPAddress:        "1.2.3.4",
		UserAgent:        "go-test",
		ExpiresAt:        time.Now().UTC().Add(24 * time.Hour),
	})
	for _, o := range overrides {
		o(s)
	}
	return s
}

func setupRepository(t *testing.T) (*sessionrepo.SessionDatabaseRepository, *sql.DB, func()) {
	t.Helper()
	db, cleanup := testHelper.SetupTestDB(t)
	return sessionrepo.NewSessionDatabaseRepository(db), db, cleanup
}

func TestSessionDatabaseRepository_Save(t *testing.T) {
	tests := []struct {
		name        string
		input       func() *domainsession.Session
		expectError bool
	}{
		{
			name:  "happy path - salva session com todos os campos preenchidos",
			input: func() *domainsession.Session { return sessionFactory() },
		},
		{
			name: "happy path - salva session com opcionais vazios (device_info/ip/ua = NULL)",
			input: func() *domainsession.Session {
				return sessionFactory(func(s *domainsession.Session) {
					s.DeviceInfo = ""
					s.IPAddress = ""
					s.UserAgent = ""
				})
			},
		},
		{
			name: "failure - refresh_token_hash duplicado retorna erro (unique)",
			input: func() *domainsession.Session {
				return sessionFactory()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _, cleanup := setupRepository(t)
			defer cleanup()

			s := tt.input()
			require.NoError(t, repo.Save(context.Background(), s))

			if tt.expectError {
				duplicate := sessionFactory(func(d *domainsession.Session) {
					d.RefreshTokenHash = s.RefreshTokenHash
				})
				err := repo.Save(context.Background(), duplicate)
				require.Error(t, err)
			}
		})
	}
}

func TestSessionDatabaseRepository_FindByRefreshTokenHash(t *testing.T) {
	tests := []struct {
		name      string
		seed      bool
		hash      string
		expectNil bool
	}{
		{
			name: "happy path - retorna session pelo hash",
			seed: true,
		},
		{
			name:      "happy path - retorna nil sem erro quando hash nao existe",
			seed:      false,
			hash:      "hash-missing",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _, cleanup := setupRepository(t)
			defer cleanup()

			var seededHash string
			if tt.seed {
				s := sessionFactory()
				require.NoError(t, repo.Save(context.Background(), s))
				seededHash = s.RefreshTokenHash
			}

			lookup := seededHash
			if tt.hash != "" {
				lookup = tt.hash
			}

			found, err := repo.FindByRefreshTokenHash(context.Background(), lookup)
			require.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, found)
				return
			}
			require.NotNil(t, found)
			assert.Equal(t, seededHash, found.RefreshTokenHash)
			assert.Equal(t, role.Specialist, found.Role)
		})
	}
}

func TestSessionDatabaseRepository_Revoke(t *testing.T) {
	tests := []struct {
		name        string
		seed        bool
		sessionID   string
		revokeFirst bool
		expectError bool
	}{
		{
			name: "happy path - session ativa vira revoked",
			seed: true,
		},
		{
			name:        "failure - session inexistente retorna erro",
			seed:        false,
			sessionID:   uuid.New().String(),
			expectError: true,
		},
		{
			name:        "failure - session ja revogada retorna erro (WHERE revoked_at IS NULL)",
			seed:        true,
			revokeFirst: true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _, cleanup := setupRepository(t)
			defer cleanup()

			var id string
			if tt.seed {
				s := sessionFactory()
				require.NoError(t, repo.Save(context.Background(), s))
				id = s.ID
			} else {
				id = tt.sessionID
			}

			if tt.revokeFirst {
				require.NoError(t, repo.Revoke(context.Background(), id))
			}

			err := repo.Revoke(context.Background(), id)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			found, err := repo.FindByRefreshTokenHash(context.Background(), "")
			_ = found
			_ = err
		})
	}
}

func TestSessionDatabaseRepository_RevokeAllForSubject(t *testing.T) {
	tests := []struct {
		name           string
		seedCount      int
		preRevoke      bool
		subjectIDOther bool
		roleOther      bool
		expectCount    int64
	}{
		{
			name:        "happy path - revoga todas as 3 sessions do subject+role ativas",
			seedCount:   3,
			expectCount: 3,
		},
		{
			name:        "happy path - 0 quando nao ha sessions",
			seedCount:   0,
			expectCount: 0,
		},
		{
			name:        "happy path - ignora sessions ja revogadas",
			seedCount:   2,
			preRevoke:   true,
			expectCount: 0,
		},
		{
			name:           "happy path - nao mexe em sessions de outro subject",
			seedCount:      2,
			subjectIDOther: true,
			expectCount:    0,
		},
		{
			name:        "happy path - nao mexe em sessions de outro role",
			seedCount:   2,
			roleOther:   true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _, cleanup := setupRepository(t)
			defer cleanup()

			subjectID := uuid.New().String()
			for i := 0; i < tt.seedCount; i++ {
				s := sessionFactory(func(s *domainsession.Session) {
					s.SubjectID = subjectID
				})
				require.NoError(t, repo.Save(context.Background(), s))
				if tt.preRevoke {
					require.NoError(t, repo.Revoke(context.Background(), s.ID))
				}
			}

			targetSubject := subjectID
			targetRole := role.Specialist
			if tt.subjectIDOther {
				targetSubject = uuid.New().String()
			}
			if tt.roleOther {
				targetRole = role.Patient
			}

			count, err := repo.RevokeAllForSubject(context.Background(), targetSubject, targetRole)
			require.NoError(t, err)
			assert.Equal(t, tt.expectCount, count)
		})
	}
}
