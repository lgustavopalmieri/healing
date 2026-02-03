package elasticsearch

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcelasticsearch "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ElasticsearchContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

func (c *ElasticsearchContainer) Address() string {
	return fmt.Sprintf("http://%s:%s", c.Host, c.Port)
}

func SetupElasticsearchContainer(t *testing.T) *ElasticsearchContainer {
	ctx := context.Background()

	elasticsearchContainer, err := tcelasticsearch.Run(ctx,
		"docker.elastic.co/elasticsearch/elasticsearch:8.17.0",
		tcelasticsearch.WithPassword(""),
		testcontainers.WithEnv(map[string]string{
			"discovery.type":                  "single-node",
			"xpack.security.enabled":          "false",
			"xpack.security.http.ssl.enabled": "false",
			"ES_JAVA_OPTS":                    "-Xms512m -Xmx512m",
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

	host, err := elasticsearchContainer.Host(ctx)
	require.NoError(t, err)

	port, err := elasticsearchContainer.MappedPort(ctx, "9200")
	require.NoError(t, err)

	return &ElasticsearchContainer{
		Container: elasticsearchContainer,
		Host:      host,
		Port:      port.Port(),
	}
}

func (c *ElasticsearchContainer) CreateClient(t *testing.T) *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{c.Address()},
	}

	client, err := elasticsearch.NewClient(cfg)
	require.NoError(t, err)

	return client
}

func (c *ElasticsearchContainer) CreateCleanIndex(t *testing.T, indexName string, createIndexFunc func(context.Context, *elasticsearch.Client, string) error) (*elasticsearch.Client, func()) {
	ctx := context.Background()

	client := c.CreateClient(t)

	err := createIndexFunc(ctx, client, indexName)
	require.NoError(t, err)

	cleanup := func() {
		client.Indices.Delete([]string{indexName})
	}

	return client, cleanup
}

func (c *ElasticsearchContainer) Terminate(t *testing.T) {
	ctx := context.Background()
	err := c.Container.Terminate(ctx)
	require.NoError(t, err)
}

type TestHelper struct {
	sharedContainer *ElasticsearchContainer
}

func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

func (h *TestHelper) RunTestMain(m *testing.M) {
	h.sharedContainer = SetupElasticsearchContainer(&testing.T{})

	code := m.Run()

	if h.sharedContainer != nil {
		h.sharedContainer.Terminate(&testing.T{})
	}

	os.Exit(code)
}

func (h *TestHelper) SetupTestIndex(t *testing.T, createIndexFunc func(context.Context, *elasticsearch.Client, string) error) (*elasticsearch.Client, string, func()) {
	indexName := "test_specialists_" + uuid.New().String()[:8]
	client, cleanup := h.sharedContainer.CreateCleanIndex(t, indexName, createIndexFunc)
	return client, indexName, cleanup
}
