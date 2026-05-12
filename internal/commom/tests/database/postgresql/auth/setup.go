package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	platformauth "github.com/lgustavopalmieri/healing-specialist/internal/platform/database/postgresql/auth"
)

type PostgreSQLContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	Database  string
	Username  string
	Password  string
}

func (c *PostgreSQLContainer) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

func SetupPostgreSQLContainer(t *testing.T) *PostgreSQLContainer {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("auth_testdb"),
		postgres.WithUsername("authuser"),
		postgres.WithPassword("authpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	return &PostgreSQLContainer{
		Container: postgresContainer,
		Host:      host,
		Port:      port.Port(),
		Database:  "auth_testdb",
		Username:  "authuser",
		Password:  "authpass",
	}
}

func (c *PostgreSQLContainer) CreateCleanDatabase(t *testing.T, dbName string) (*sql.DB, func()) {
	ctx := context.Background()

	adminConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port)

	adminDB, err := sql.Open("postgres", adminConnStr)
	require.NoError(t, err)
	defer adminDB.Close()

	_, err = adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err)

	testConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port, dbName)

	testDB, err := sql.Open("postgres", testConnStr)
	require.NoError(t, err)

	err = testDB.Ping()
	require.NoError(t, err)

	err = platformauth.RunMigrations(testDB)
	require.NoError(t, err)

	cleanup := func() {
		testDB.Close()
		adminDB, err := sql.Open("postgres", adminConnStr)
		if err == nil {
			adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
			adminDB.Close()
		}
	}

	return testDB, cleanup
}

func (c *PostgreSQLContainer) Terminate(t *testing.T) {
	ctx := context.Background()
	err := c.Container.Terminate(ctx)
	require.NoError(t, err)
}

type TestHelper struct {
	sharedContainer *PostgreSQLContainer
}

func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

func (h *TestHelper) RunTestMain(m *testing.M) {
	h.sharedContainer = SetupPostgreSQLContainer(&testing.T{})

	code := m.Run()

	if h.sharedContainer != nil {
		h.sharedContainer.Terminate(&testing.T{})
	}

	os.Exit(code)
}

func (h *TestHelper) RunTestMainWithoutExit(m *testing.M) {
	h.sharedContainer = SetupPostgreSQLContainer(&testing.T{})
}

func (h *TestHelper) TerminateContainer() {
	if h.sharedContainer != nil {
		h.sharedContainer.Terminate(&testing.T{})
	}
}

func (h *TestHelper) SetupTestDB(t *testing.T) (*sql.DB, func()) {
	dbName := "test_auth_" + uuid.New().String()[:8]
	return h.sharedContainer.CreateCleanDatabase(t, dbName)
}
