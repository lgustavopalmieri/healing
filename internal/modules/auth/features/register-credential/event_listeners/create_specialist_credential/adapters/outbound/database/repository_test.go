package database_test

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
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/auth/features/register-credential/event_listeners/create_specialist_credential/adapters/outbound/database"
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

func setupRepository(t *testing.T) (*database.CredentialDatabaseRepository, *sql.DB, func()) {
	t.Helper()
	db, cleanup := testHelper.SetupTestDB(t)
	repoIface := database.NewCredentialDatabaseRepository(db)
	repo, ok := repoIface.(*database.CredentialDatabaseRepository)
	require.True(t, ok)
	return repo, db, cleanup
}

func TestCredentialDatabaseRepository_Save(t *testing.T) {
	tests := []struct {
		name           string
		input          func() *credential.Credential
		seed           func(t *testing.T, db *sql.DB, repo *database.CredentialDatabaseRepository, input *credential.Credential)
		expectError    bool
		errMsgContains string
		validateResult func(t *testing.T, db *sql.DB, repo *database.CredentialDatabaseRepository, input *credential.Credential)
	}{
		{
			name:  "happy path - salva credential pending e persiste",
			input: func() *credential.Credential { return credentialFactory() },
			validateResult: func(t *testing.T, _ *sql.DB, repo *database.CredentialDatabaseRepository, input *credential.Credential) {
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
			validateResult: func(t *testing.T, _ *sql.DB, repo *database.CredentialDatabaseRepository, input *credential.Credential) {
				found, err := repo.FindByEmailProviderRole(context.Background(), input.Email, input.Provider, input.Role)
				require.NoError(t, err)
				require.NotNil(t, found)
				assert.Empty(t, found.ProviderUserID)
				assert.Empty(t, found.PasswordHash)
				assert.Nil(t, found.LastUsedAt)
			},
		},
		{
			name: "happy path - salva credential com providerUserID e passwordHash preenchidos",
			input: func() *credential.Credential {
				return credentialFactory(func(c *credential.Credential) {
					c.ProviderUserID = "provider-user-id-123"
					c.PasswordHash = "bcrypt-hash-xyz"
					c.Status = credential.StatusActive
				})
			},
			validateResult: func(t *testing.T, _ *sql.DB, repo *database.CredentialDatabaseRepository, input *credential.Credential) {
				found, err := repo.FindByEmailProviderRole(context.Background(), input.Email, input.Provider, input.Role)
				require.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, "provider-user-id-123", found.ProviderUserID)
				assert.Equal(t, "bcrypt-hash-xyz", found.PasswordHash)
			},
		},
		{
			name:  "failure - segundo Save para mesmos email/provider/role retorna CredentialAlreadyExistsErr",
			input: func() *credential.Credential { return credentialFactory() },
			seed: func(t *testing.T, _ *sql.DB, repo *database.CredentialDatabaseRepository, input *credential.Credential) {
				existing := credentialFactory(func(c *credential.Credential) {
					c.Email = input.Email
					c.Role = input.Role
					c.Provider = input.Provider
				})
				require.NoError(t, repo.Save(context.Background(), existing))
			},
			expectError:    true,
			errMsgContains: database.CredentialAlreadyExistsErr,
		},
		{
			name:  "happy path - Save permite novo registro quando existe credential deleted com mesmo trio",
			input: func() *credential.Credential { return credentialFactory() },
			seed: func(t *testing.T, db *sql.DB, repo *database.CredentialDatabaseRepository, input *credential.Credential) {
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
		seed           func(t *testing.T, db *sql.DB, repo *database.CredentialDatabaseRepository) *credential.Credential
		query          func(seeded *credential.Credential) (email string, p provider.Provider, r role.Role)
		expectNil      bool
		validateResult func(t *testing.T, seeded, found *credential.Credential)
	}{
		{
			name: "happy path - retorna credential com todos os campos hidratados",
			seed: func(t *testing.T, _ *sql.DB, repo *database.CredentialDatabaseRepository) *credential.Credential {
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
			seed: func(t *testing.T, db *sql.DB, repo *database.CredentialDatabaseRepository) *credential.Credential {
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
			seed: func(t *testing.T, _ *sql.DB, _ *database.CredentialDatabaseRepository) *credential.Credential {
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
			seed: func(t *testing.T, db *sql.DB, repo *database.CredentialDatabaseRepository) *credential.Credential {
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
			seed: func(t *testing.T, _ *sql.DB, repo *database.CredentialDatabaseRepository) *credential.Credential {
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
