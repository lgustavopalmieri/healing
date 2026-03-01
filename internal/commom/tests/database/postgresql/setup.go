package postgresql

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

	"github.com/lgustavopalmieri/healing-specialist/internal/platform/database/postgresql"
)

// PostgreSQLContainer wraps the testcontainers postgres container
type PostgreSQLContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	Database  string
	Username  string
	Password  string
}

// ConnectionString returns the database connection string
func (c *PostgreSQLContainer) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// SetupPostgreSQLContainer creates and starts a PostgreSQL container
func SetupPostgreSQLContainer(t *testing.T) *PostgreSQLContainer {
	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	return &PostgreSQLContainer{
		Container: postgresContainer,
		Host:      host,
		Port:      port.Port(),
		Database:  "testdb",
		Username:  "testuser",
		Password:  "testpass",
	}
}

// CreateCleanDatabase creates a new clean database for each test
func (c *PostgreSQLContainer) CreateCleanDatabase(t *testing.T, dbName string) (*sql.DB, func()) {
	ctx := context.Background()

	// Connect to the default database to create a new one
	adminConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port)

	adminDB, err := sql.Open("postgres", adminConnStr)
	require.NoError(t, err)
	defer adminDB.Close()

	// Create the test database
	_, err = adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err)

	// Connect to the new database
	testConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username, c.Password, c.Host, c.Port, dbName)

	testDB, err := sql.Open("postgres", testConnStr)
	require.NoError(t, err)

	// Test connection
	err = testDB.Ping()
	require.NoError(t, err)

	// Run migrations
	err = postgresql.RunMigrations(testDB)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		testDB.Close()
		// Drop the test database
		adminDB, err := sql.Open("postgres", adminConnStr)
		if err == nil {
			adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
			adminDB.Close()
		}
	}

	return testDB, cleanup
}

// Terminate stops and removes the container
func (c *PostgreSQLContainer) Terminate(t *testing.T) {
	ctx := context.Background()
	err := c.Container.Terminate(ctx)
	require.NoError(t, err)
}

// TestHelper provides a reusable test setup pattern
type TestHelper struct {
	sharedContainer *PostgreSQLContainer
}

// NewTestHelper creates a new test helper instance
func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

// RunTestMain provides a reusable TestMain implementation
func (h *TestHelper) RunTestMain(m *testing.M) {
	// Setup shared container once for all tests
	h.sharedContainer = SetupPostgreSQLContainer(&testing.T{})

	// Run tests
	code := m.Run()

	// Cleanup
	if h.sharedContainer != nil {
		h.sharedContainer.Terminate(&testing.T{})
	}

	// Exit with the test result code
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

// SetupTestDB creates a clean database for a single test
func (h *TestHelper) SetupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a unique database name for this test
	dbName := "test_" + uuid.New().String()[:8]
	return h.sharedContainer.CreateCleanDatabase(t, dbName)
}
