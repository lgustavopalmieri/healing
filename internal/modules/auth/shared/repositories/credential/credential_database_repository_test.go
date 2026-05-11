package credential_test

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
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/credential"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/password"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/domain/session"
	credentialrepo "github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/shared/repositories/credential"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/provider"
	"github.com/lgustavopalmieri/healing-specialist/pkg/healing-auth/role"
)

var testHelper = authtestdb.NewTestHelper()

func TestMain(m *testing.M) {
	testHelper.RunTestMainWithoutExit(m)
	code := m.Run()
	testHelper.TerminateContainer()
	os.Exit(code)
}

func credentialFactory(overrides ...func(*credential.Credential)) *credential.Credential {
	c := credential.NewCredential(credential.NewCredentialInput{
		SubjectID: uuid.New().String(),
		Role:      role.Specialist,
		Provider:  provider.Password,
		Email:     uuid.New().String() + "@healing.com",
	})
	for _, o := range overrides {
		o(c)
	}
	return c
}

func setupRepository(t *testing.T) (*credentialrepo.CredentialDatabaseRepository, *sql.DB, func()) {
	t.Helper()
	db, cleanup := testHelper.SetupTestDB(t)
	repo := credentialrepo.NewCredentialDatabaseRepository(db)
	return repo, db, cleanup
}

func TestCredentialDatabaseRepository_Save(t *testing.T) {
	tests := []struct {
		name           string
		input          func() *credential.Credential
		seed           func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository, input *credential.Credential)
		expectError    bool
		errMsgContains string
		validateResult func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository, input *credential.Credential)
	}{
		{
			name:  "happy path - salva credential pending e persiste",
			input: func() *credential.Credential { return credentialFactory() },
			validateResult: func(t *testing.T, _ *sql.DB, repo *credentialrepo.CredentialDatabaseRepository, input *credential.Credential) {
				found, err := repo.FindByEmailProviderRole(context.Background(), input.Email, input.Provider, input.Role)
				require.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, input.ID, found.ID)
				assert.Equal(t, input.SubjectID, found.SubjectID)
				assert.Equal(t, credential.StatusPending, found.Status)
			},
		},
		{
			name:  "happy path - salva credential sem providerUserID e sem passwordHash (NULLs)",
			input: func() *credential.Credential { return credentialFactory() },
			validateResult: func(t *testing.T, _ *sql.DB, repo *credentialrepo.CredentialDatabaseRepository, input *credential.Credential) {
				found, err := repo.FindByEmailProviderRole(context.Background(), input.Email, input.Provider, input.Role)
				require.NoError(t, err)
				require.NotNil(t, found)
				assert.Empty(t, found.ProviderUserID)
				assert.True(t, found.PasswordHash.IsEmpty())
				assert.Nil(t, found.LastUsedAt)
			},
		},
		{
			name: "happy path - salva credential com providerUserID e passwordHash preenchidos",
			input: func() *credential.Credential {
				return credentialFactory(func(c *credential.Credential) {
					c.ProviderUserID = "provider-user-id-123"
					c.PasswordHash = password.NewHashedPassword("bcrypt-hash-xyz")
					c.Status = credential.StatusActive
				})
			},
			validateResult: func(t *testing.T, _ *sql.DB, repo *credentialrepo.CredentialDatabaseRepository, input *credential.Credential) {
				found, err := repo.FindByEmailProviderRole(context.Background(), input.Email, input.Provider, input.Role)
				require.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, "provider-user-id-123", found.ProviderUserID)
				assert.Equal(t, "bcrypt-hash-xyz", found.PasswordHash.String())
			},
		},
		{
			name:  "failure - segundo Save para mesmos email/provider/role retorna CredentialAlreadyExistsErr",
			input: func() *credential.Credential { return credentialFactory() },
			seed: func(t *testing.T, _ *sql.DB, repo *credentialrepo.CredentialDatabaseRepository, input *credential.Credential) {
				existing := credentialFactory(func(c *credential.Credential) {
					c.Email = input.Email
					c.Role = input.Role
					c.Provider = input.Provider
				})
				require.NoError(t, repo.Save(context.Background(), existing))
			},
			expectError:    true,
			errMsgContains: credentialrepo.CredentialAlreadyExistsErr,
		},
		{
			name:  "happy path - Save permite novo registro quando existe credential deleted com mesmo trio",
			input: func() *credential.Credential { return credentialFactory() },
			seed: func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository, input *credential.Credential) {
				deleted := credentialFactory(func(c *credential.Credential) {
					c.Email = input.Email
					c.Role = input.Role
					c.Provider = input.Provider
				})
				require.NoError(t, repo.Save(context.Background(), deleted))

				_, err := db.ExecContext(
					context.Background(),
					"UPDATE credentials SET status = $1 WHERE id = $2",
					string(credential.StatusDeleted),
					deleted.ID,
				)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, db, cleanup := setupRepository(t)
			defer cleanup()

			input := tt.input()
			if tt.seed != nil {
				tt.seed(t, db, repo, input)
			}

			err := repo.Save(context.Background(), input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errMsgContains != "" {
					assert.Contains(t, err.Error(), tt.errMsgContains)
				}
				return
			}
			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, db, repo, input)
			}
		})
	}
}

func TestCredentialDatabaseRepository_FindByEmailProviderRole(t *testing.T) {
	tests := []struct {
		name           string
		seed           func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential
		query          func(seeded *credential.Credential) (email string, p provider.Provider, r role.Role)
		expectNil      bool
		validateResult func(t *testing.T, seeded, found *credential.Credential)
	}{
		{
			name: "happy path - retorna credential com todos os campos hidratados",
			seed: func(t *testing.T, _ *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				cred := credentialFactory()
				require.NoError(t, repo.Save(context.Background(), cred))
				return cred
			},
			query: func(seeded *credential.Credential) (string, provider.Provider, role.Role) {
				return seeded.Email, seeded.Provider, seeded.Role
			},
			validateResult: func(t *testing.T, seeded, found *credential.Credential) {
				assert.Equal(t, seeded.ID, found.ID)
				assert.Equal(t, seeded.SubjectID, found.SubjectID)
				assert.Equal(t, role.Specialist, found.Role)
				assert.Equal(t, provider.Password, found.Provider)
				assert.Equal(t, seeded.Email, found.Email)
				assert.Equal(t, credential.StatusPending, found.Status)
				assert.False(t, found.CreatedAt.IsZero())
				assert.False(t, found.UpdatedAt.IsZero())
			},
		},
		{
			name: "happy path - retorna credential com LastUsedAt quando existe",
			seed: func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				cred := credentialFactory()
				require.NoError(t, repo.Save(context.Background(), cred))

				lastUsed := time.Now().UTC().Truncate(time.Second)
				_, err := db.ExecContext(
					context.Background(),
					"UPDATE credentials SET last_used_at = $1 WHERE id = $2",
					lastUsed,
					cred.ID,
				)
				require.NoError(t, err)
				cred.LastUsedAt = &lastUsed
				return cred
			},
			query: func(seeded *credential.Credential) (string, provider.Provider, role.Role) {
				return seeded.Email, seeded.Provider, seeded.Role
			},
			validateResult: func(t *testing.T, seeded, found *credential.Credential) {
				require.NotNil(t, found.LastUsedAt)
				assert.Equal(t, seeded.LastUsedAt.Unix(), found.LastUsedAt.Unix())
			},
		},
		{
			name: "happy path - retorna nil sem erro quando nenhuma credencial existe",
			seed: func(t *testing.T, _ *sql.DB, _ *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				return &credential.Credential{
					Email:    "ghost@healing.com",
					Provider: provider.Password,
					Role:     role.Specialist,
				}
			},
			query: func(seeded *credential.Credential) (string, provider.Provider, role.Role) {
				return seeded.Email, seeded.Provider, seeded.Role
			},
			expectNil: true,
		},
		{
			name: "happy path - ignora credencial com status=deleted",
			seed: func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				cred := credentialFactory()
				require.NoError(t, repo.Save(context.Background(), cred))

				_, err := db.ExecContext(
					context.Background(),
					"UPDATE credentials SET status = $1 WHERE id = $2",
					string(credential.StatusDeleted),
					cred.ID,
				)
				require.NoError(t, err)
				return cred
			},
			query: func(seeded *credential.Credential) (string, provider.Provider, role.Role) {
				return seeded.Email, seeded.Provider, seeded.Role
			},
			expectNil: true,
		},
		{
			name: "happy path - role diferente retorna nil mesmo com email e provider batendo",
			seed: func(t *testing.T, _ *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				specialist := credentialFactory()
				require.NoError(t, repo.Save(context.Background(), specialist))

				patient := credentialFactory(func(c *credential.Credential) {
					c.Email = specialist.Email
					c.Role = role.Patient
				})
				require.NoError(t, repo.Save(context.Background(), patient))
				return specialist
			},
			query: func(seeded *credential.Credential) (string, provider.Provider, role.Role) {
				return seeded.Email, provider.Password, role.Admin
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, db, cleanup := setupRepository(t)
			defer cleanup()

			seeded := tt.seed(t, db, repo)
			email, p, r := tt.query(seeded)

			found, err := repo.FindByEmailProviderRole(context.Background(), email, p, r)

			require.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, found)
				return
			}
			require.NotNil(t, found)
			if tt.validateResult != nil {
				tt.validateResult(t, seeded, found)
			}
		})
	}
}

func TestCredentialDatabaseRepository_FindBySubjectAndRole(t *testing.T) {
	tests := []struct {
		name           string
		seed           func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential
		query          func(seeded *credential.Credential) (subjectID string, r role.Role)
		expectNil      bool
		validateResult func(t *testing.T, seeded, found *credential.Credential)
	}{
		{
			name: "happy path - retorna credential pelo subject + role",
			seed: func(t *testing.T, _ *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				cred := credentialFactory()
				require.NoError(t, repo.Save(context.Background(), cred))
				return cred
			},
			query: func(seeded *credential.Credential) (string, role.Role) {
				return seeded.SubjectID, seeded.Role
			},
			validateResult: func(t *testing.T, seeded, found *credential.Credential) {
				assert.Equal(t, seeded.ID, found.ID)
				assert.Equal(t, seeded.Email, found.Email)
			},
		},
		{
			name: "happy path - retorna nil quando subject nao existe",
			seed: func(t *testing.T, _ *sql.DB, _ *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				return &credential.Credential{SubjectID: uuid.New().String(), Role: role.Specialist}
			},
			query: func(seeded *credential.Credential) (string, role.Role) {
				return seeded.SubjectID, role.Specialist
			},
			expectNil: true,
		},
		{
			name: "happy path - ignora credencial deleted",
			seed: func(t *testing.T, db *sql.DB, repo *credentialrepo.CredentialDatabaseRepository) *credential.Credential {
				cred := credentialFactory()
				require.NoError(t, repo.Save(context.Background(), cred))
				_, err := db.ExecContext(context.Background(),
					"UPDATE credentials SET status = $1 WHERE id = $2",
					string(credential.StatusDeleted), cred.ID)
				require.NoError(t, err)
				return cred
			},
			query: func(seeded *credential.Credential) (string, role.Role) {
				return seeded.SubjectID, seeded.Role
			},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, db, cleanup := setupRepository(t)
			defer cleanup()

			seeded := tt.seed(t, db, repo)
			subjectID, r := tt.query(seeded)

			found, err := repo.FindBySubjectAndRole(context.Background(), subjectID, r)

			require.NoError(t, err)
			if tt.expectNil {
				assert.Nil(t, found)
				return
			}
			require.NotNil(t, found)
			if tt.validateResult != nil {
				tt.validateResult(t, seeded, found)
			}
		})
	}
}

func TestCredentialDatabaseRepository_Update(t *testing.T) {
	tests := []struct {
		name           string
		mutate         func(c *credential.Credential)
		expectError    bool
		errMsg         string
		validateResult func(t *testing.T, updated, found *credential.Credential)
	}{
		{
			name: "happy path - Activate pending -> active com password hash",
			mutate: func(c *credential.Credential) {
				require.NoError(t, c.Activate(password.NewHashedPassword("new-hash")))
			},
			validateResult: func(t *testing.T, updated, found *credential.Credential) {
				assert.Equal(t, credential.StatusActive, found.Status)
				assert.Equal(t, "new-hash", found.PasswordHash.String())
			},
		},
		{
			name: "happy path - Lock active -> locked",
			mutate: func(c *credential.Credential) {
				require.NoError(t, c.Activate(password.NewHashedPassword("any-hash")))
				require.NoError(t, c.Lock())
			},
			validateResult: func(t *testing.T, _, found *credential.Credential) {
				assert.Equal(t, credential.StatusLocked, found.Status)
			},
		},
		{
			name: "failure - Update em credential inexistente retorna sql.ErrNoRows wrapping",
			mutate: func(c *credential.Credential) {
				c.ID = uuid.New().String()
				require.NoError(t, c.Activate(password.NewHashedPassword("hash")))
			},
			expectError: true,
			errMsg:      "no rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _, cleanup := setupRepository(t)
			defer cleanup()

			cred := credentialFactory()
			require.NoError(t, repo.Save(context.Background(), cred))

			tt.mutate(cred)

			err := repo.Update(context.Background(), cred)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)

			found, err := repo.FindBySubjectAndRole(context.Background(), cred.SubjectID, cred.Role)
			require.NoError(t, err)
			require.NotNil(t, found)
			if tt.validateResult != nil {
				tt.validateResult(t, cred, found)
			}
		})
	}
}

func TestCredentialDatabaseRepository_UpdateWithSessionInTransaction(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(cred *credential.Credential, sess *session.Session)
		breakAt       func(sess *session.Session)
		expectError   bool
		expectPersist bool
	}{
		{
			name: "happy path - commit persiste credential active + session",
			setup: func(cred *credential.Credential, _ *session.Session) {
				require.NoError(t, cred.Activate(password.NewHashedPassword("hashed-value")))
			},
			expectPersist: true,
		},
		{
			name: "failure - violação de constraint faz rollback: credential volta ao estado pending",
			setup: func(cred *credential.Credential, sess *session.Session) {
				require.NoError(t, cred.Activate(password.NewHashedPassword("hashed-value")))
			},
			breakAt: func(sess *session.Session) {
				sess.IPAddress = "this-ip-address-is-way-longer-than-forty-five-characters-limit"
			},
			expectError:   true,
			expectPersist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, db, cleanup := setupRepository(t)
			defer cleanup()

			cred := credentialFactory()
			require.NoError(t, repo.Save(context.Background(), cred))

			sess := session.NewSession(session.NewSessionInput{
				SubjectID:        cred.SubjectID,
				Role:             cred.Role,
				RefreshTokenHash: uuid.New().String(),
				ExpiresAt:        time.Now().Add(24 * time.Hour),
			})

			if tt.setup != nil {
				tt.setup(cred, sess)
			}
			if tt.breakAt != nil {
				tt.breakAt(sess)
			}

			err := repo.UpdateWithSessionInTransaction(context.Background(), cred, sess)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			found, err := repo.FindBySubjectAndRole(context.Background(), cred.SubjectID, cred.Role)
			require.NoError(t, err)
			require.NotNil(t, found)

			var sessionCount int
			err = db.QueryRowContext(context.Background(),
				"SELECT COUNT(*) FROM sessions WHERE subject_id = $1", cred.SubjectID).Scan(&sessionCount)
			require.NoError(t, err)

			if tt.expectPersist {
				assert.Equal(t, credential.StatusActive, found.Status)
				assert.False(t, found.PasswordHash.IsEmpty())
				assert.Equal(t, 1, sessionCount)
			} else {
				assert.Equal(t, credential.StatusPending, found.Status)
				assert.True(t, found.PasswordHash.IsEmpty())
				assert.Equal(t, 0, sessionCount)
			}
		})
	}
}
