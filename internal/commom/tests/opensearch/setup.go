package opensearch

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	opensearchgo "github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcopensearch "github.com/testcontainers/testcontainers-go/modules/opensearch"
	"github.com/testcontainers/testcontainers-go/wait"
)

type OpenSearchContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

func (c *OpenSearchContainer) Address() string {
	return fmt.Sprintf("http://%s:%s", c.Host, c.Port)
}

func SetupOpenSearchContainer(t *testing.T) *OpenSearchContainer {
	ctx := context.Background()

	container, err := tcopensearch.Run(ctx,
		"opensearchproject/opensearch:2.19.0",
		testcontainers.WithEnv(map[string]string{
			"discovery.type":          "single-node",
			"DISABLE_SECURITY_PLUGIN": "true",
			"OPENSEARCH_JAVA_OPTS":    "-Xms512m -Xmx512m",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/_cluster/health").
				WithPort("9200/tcp").
				WithStatusCodeMatcher(func(status int) bool {
					return status == 200
				}).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "9200")
	require.NoError(t, err)

	return &OpenSearchContainer{
		Container: container,
		Host:      host,
		Port:      port.Port(),
	}
}

func (c *OpenSearchContainer) CreateClient(t *testing.T) *opensearchapi.Client {
	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearchgo.Config{
			Addresses: []string{c.Address()},
		},
	})
	require.NoError(t, err)

	return client
}

func (c *OpenSearchContainer) CreateCleanIndex(t *testing.T, indexName string, createIndexFunc func(context.Context, *opensearchapi.Client, string) error) (*opensearchapi.Client, func()) {
	ctx := context.Background()

	client := c.CreateClient(t)

	err := createIndexFunc(ctx, client, indexName)
	require.NoError(t, err)

	cleanup := func() {
		client.Indices.Delete(ctx, opensearchapi.IndicesDeleteReq{
			Indices: []string{indexName},
		})
	}

	return client, cleanup
}

func (c *OpenSearchContainer) Terminate(t *testing.T) {
	ctx := context.Background()
	err := c.Container.Terminate(ctx)
	require.NoError(t, err)
}

type TestHelper struct {
	sharedContainer *OpenSearchContainer
}

func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

func (h *TestHelper) RunTestMain(m *testing.M) {
	h.sharedContainer = SetupOpenSearchContainer(&testing.T{})

	code := m.Run()

	if h.sharedContainer != nil {
		h.sharedContainer.Terminate(&testing.T{})
	}

	os.Exit(code)
}

func (h *TestHelper) SetupTestIndex(t *testing.T, createIndexFunc func(context.Context, *opensearchapi.Client, string) error) (*opensearchapi.Client, string, func()) {
	indexName := "test_specialists_" + uuid.New().String()[:8]
	client, cleanup := h.sharedContainer.CreateCleanIndex(t, indexName, createIndexFunc)
	return client, indexName, cleanup
}
