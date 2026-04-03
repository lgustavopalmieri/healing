package sqs

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/lgustavopalmieri/healing-specialist/internal/commom/event"
)

const defaultWorkerPoolSize = 10

type SQSConsumer struct {
	client         *awssqs.Client
	queueURL       string
	listener       event.Listener
	topic          string
	workerPoolSize int
}

func NewSQSConsumer(client *awssqs.Client, queueURL string, topic string, listener event.Listener) *SQSConsumer {
	return &SQSConsumer{
		client:         client,
		queueURL:       queueURL,
		listener:       listener,
		topic:          topic,
		workerPoolSize: defaultWorkerPoolSize,
	}
}

func NewSQSConsumerWithWorkers(client *awssqs.Client, queueURL string, topic string, listener event.Listener, workerPoolSize int) *SQSConsumer {
	if workerPoolSize <= 0 {
		workerPoolSize = defaultWorkerPoolSize
	}
	return &SQSConsumer{
		client:         client,
		queueURL:       queueURL,
		listener:       listener,
		topic:          topic,
		workerPoolSize: workerPoolSize,
	}
}

func (c *SQSConsumer) Start(ctx context.Context) {
	log.Printf("SQS consumer started for queue: %s (topic: %s, workers: %d)", c.queueURL, c.topic, c.workerPoolSize)

	sem := make(chan struct{}, c.workerPoolSize)

	for {
		if ctx.Err() != nil {
			log.Printf("SQS consumer stopped for topic %s: context cancelled", c.topic)
			return
		}

		output, err := c.client.ReceiveMessage(ctx, &awssqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
			VisibilityTimeout:   30,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("SQS receive error on topic %s: %v", c.topic, err)
			continue
		}

		var wg sync.WaitGroup

		for _, msg := range output.Messages {
			wg.Add(1)
			sem <- struct{}{}

			go func(m sqstypes.Message) {
				defer wg.Done()
				defer func() { <-sem }()

				c.processMessage(ctx, m)
			}(msg)
		}

		wg.Wait()
	}
}

func (c *SQSConsumer) processMessage(ctx context.Context, msg sqstypes.Message) {
	evt := event.NewEvent(c.topic, []byte(*msg.Body))

	if err := c.listener.Handle(ctx, evt); err != nil {
		log.Printf("Error handling event on topic %s: %v (message will return to queue)", c.topic, err)
		return
	}

	_, deleteErr := c.client.DeleteMessage(ctx, &awssqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if deleteErr != nil {
		log.Printf("Failed to delete message on topic %s: %v", c.topic, deleteErr)
	}
}
