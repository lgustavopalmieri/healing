package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tclocalstack "github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type LocalStackContainer struct {
	Container testcontainers.Container
	Endpoint  string
}

func SetupLocalStackContainer(t *testing.T) *LocalStackContainer {
	ctx := context.Background()

	container, err := tclocalstack.Run(ctx,
		"localstack/localstack:3.8",
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "sqs",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/_localstack/health").
				WithPort("4566/tcp").
				WithStatusCodeMatcher(func(status int) bool {
					return status == 200
				}).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "4566")
	require.NoError(t, err)

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	return &LocalStackContainer{
		Container: container,
		Endpoint:  endpoint,
	}
}

func (c *LocalStackContainer) Terminate(t *testing.T) {
	ctx := context.Background()
	err := c.Container.Terminate(ctx)
	require.NoError(t, err)
}

func (c *LocalStackContainer) CreateSQSClient(t *testing.T) *awssqs.Client {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
	)
	require.NoError(t, err)

	endpoint := c.Endpoint
	client := awssqs.NewFromConfig(awsCfg, func(o *awssqs.Options) {
		o.BaseEndpoint = &endpoint
	})
	return client
}

func (c *LocalStackContainer) CreateProducer(t *testing.T, queueURLs map[string]string) *platformsqs.SQSProducer {
	client := c.CreateSQSClient(t)
	return platformsqs.NewSQSProducer(client, queueURLs)
}

func (c *LocalStackContainer) EnsureQueues(t *testing.T, prefix string) (map[string]string, *awssqs.Client) {
	client := c.CreateSQSClient(t)
	urls, err := platformsqs.EnsureQueues(context.Background(), client, prefix, platformsqs.DefaultQueueDefinitions())
	require.NoError(t, err)
	return urls, client
}

func (c *LocalStackContainer) ProduceEvent(t *testing.T, client *awssqs.Client, queueURL string, payload any) {
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	_, err = client.SendMessage(context.Background(), &awssqs.SendMessageInput{
		QueueUrl:               aws.String(queueURL),
		MessageBody:            aws.String(string(data)),
		MessageGroupId:         aws.String("test-group"),
		MessageDeduplicationId: aws.String(fmt.Sprintf("dedup-%d", time.Now().UnixNano())),
	})
	require.NoError(t, err)
}

func (c *LocalStackContainer) ConsumeEvent(t *testing.T, client *awssqs.Client, queueURL string, timeout time.Duration) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		if ctx.Err() != nil {
			t.Fatalf("timeout waiting for message on queue %s", queueURL)
			return nil
		}

		output, err := client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			if ctx.Err() != nil {
				t.Fatalf("timeout waiting for message on queue %s", queueURL)
				return nil
			}
			continue
		}

		if len(output.Messages) > 0 {
			_, _ = client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: output.Messages[0].ReceiptHandle,
			})
			return []byte(*output.Messages[0].Body)
		}
	}
}

type TestHelper struct {
	SharedContainer *LocalStackContainer
	SQSClient       *awssqs.Client
	QueueURLs       map[string]string
}

func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

func (h *TestHelper) RunTestMain(m *testing.M) {
	h.SharedContainer = SetupLocalStackContainer(&testing.T{})

	urls, client := h.SharedContainer.EnsureQueues(&testing.T{}, "specialist")
	h.SQSClient = client
	h.QueueURLs = urls

	code := m.Run()

	if h.SharedContainer != nil {
		h.SharedContainer.Terminate(&testing.T{})
	}

	os.Exit(code)
}

func (h *TestHelper) CreateConsumer(t *testing.T, queueURL string, topic string, listener event.Listener) *platformsqs.SQSConsumer {
	return platformsqs.NewSQSConsumer(h.SQSClient, queueURL, topic, listener)
}
