package sns_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tclocalstack "github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
	platformsns "github.com/lgustavopalmieri/healing-specialist/internal/platform/sns"
	platformsqs "github.com/lgustavopalmieri/healing-specialist/internal/platform/sqs"
)

type localStackFixture struct {
	endpoint  string
	snsClient *awssns.Client
	sqsClient *awssqs.Client
	topicARNs map[string]string
	queueURLs map[string]string
	terminate func()
}

func setupLocalStack(t *testing.T) *localStackFixture {
	t.Helper()

	ctx := context.Background()
	container, err := tclocalstack.Run(ctx,
		"localstack/localstack:3.8",
		testcontainers.WithEnv(map[string]string{"SERVICES": "sqs,sns"}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/_localstack/health").
				WithPort("4566/tcp").
				WithStatusCodeMatcher(func(status int) bool { return status == 200 }).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "4566")
	require.NoError(t, err)
	endpoint := "http://" + host + ":" + port.Port()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
	)
	require.NoError(t, err)

	sqsClient := awssqs.NewFromConfig(awsCfg, func(o *awssqs.Options) { o.BaseEndpoint = &endpoint })
	snsClient := awssns.NewFromConfig(awsCfg, func(o *awssns.Options) { o.BaseEndpoint = &endpoint })

	queueURLs, err := platformsqs.EnsureConsumerQueues(ctx, sqsClient, "test", platformsqs.DefaultConsumerQueueDefinitions())
	require.NoError(t, err)

	topicARNs, err := platformsns.EnsureTopics(ctx, snsClient, "test", platformsns.DefaultTopicDefinitions())
	require.NoError(t, err)

	subs := make([]platformsns.SubscriptionDefinition, 0)
	for _, def := range platformsqs.DefaultConsumerQueueDefinitions() {
		if def.SubscribesToEvent == "" {
			continue
		}
		subs = append(subs, platformsns.SubscriptionDefinition{
			EventName:    def.SubscribesToEvent,
			ConsumerName: def.ConsumerName,
			QueueURL:     queueURLs[def.ConsumerName],
		})
	}
	require.NoError(t, platformsns.EnsureSubscriptions(ctx, snsClient, sqsClient, topicARNs, subs))

	return &localStackFixture{
		endpoint:  endpoint,
		snsClient: snsClient,
		sqsClient: sqsClient,
		topicARNs: topicARNs,
		queueURLs: queueURLs,
		terminate: func() {
			termCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = container.Terminate(termCtx)
		},
	}
}

func waitForMessage(t *testing.T, client *awssqs.Client, queueURL string, timeout time.Duration) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		if ctx.Err() != nil {
			t.Fatalf("timeout waiting for message on queue %s", queueURL)
		}
		out, err := client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			continue
		}
		if len(out.Messages) > 0 {
			_, _ = client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: out.Messages[0].ReceiptHandle,
			})
			return *out.Messages[0].Body
		}
	}
}

func TestSNSProducer_Dispatch_FanoutToSubscribedConsumers(t *testing.T) {
	tests := []struct {
		name              string
		eventName         string
		payload           map[string]any
		expectedConsumers []string
	}{
		{
			name:      "happy path - specialist.created entregue para validate_license E register_credential",
			eventName: "specialist.created",
			payload: map[string]any{
				"id":    "specialist-1",
				"email": "specialist@healing.com",
			},
			expectedConsumers: []string{
				"specialist-validate-license",
				"specialist-register-credential",
			},
		},
		{
			name:      "happy path - specialist.updated entregue apenas para update_data_repos",
			eventName: "specialist.updated",
			payload: map[string]any{
				"id":    "specialist-2",
				"email": "specialist2@healing.com",
			},
			expectedConsumers: []string{
				"specialist-update-data-repos",
			},
		},
	}

	fixture := setupLocalStack(t)
	defer fixture.terminate()

	producer := platformsns.NewSNSProducer(fixture.snsClient, fixture.topicARNs)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := producer.Dispatch(context.Background(), event.NewEvent(tt.eventName, tt.payload))
			require.NoError(t, err)

			for _, consumer := range tt.expectedConsumers {
				queueURL := fixture.queueURLs[consumer]
				require.NotEmpty(t, queueURL)

				body := waitForMessage(t, fixture.sqsClient, queueURL, 15*time.Second)
				require.NotEmpty(t, body)

				var received map[string]any
				require.NoError(t, json.Unmarshal([]byte(body), &received))
				assert.Equal(t, tt.payload["id"], received["id"])
				assert.Equal(t, tt.payload["email"], received["email"])
			}
		})
	}
}

func TestSNSProducer_Dispatch_UnknownEventName(t *testing.T) {
	tests := []struct {
		name        string
		eventName   string
		expectedErr string
	}{
		{
			name:        "failure - evento sem topic configurado retorna erro",
			eventName:   "unknown.event",
			expectedErr: "no topic ARN configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := platformsns.NewSNSProducer(nil, map[string]string{})
			err := producer.Dispatch(context.Background(), event.NewEvent(tt.eventName, nil))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
