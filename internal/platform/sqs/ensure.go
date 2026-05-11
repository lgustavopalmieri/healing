package sqs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type ConsumerQueueDefinition struct {
	ConsumerName      string
	SubscribesToEvent string
	MaxReceiveCount   int
	VisibilityTimeout int
	RetentionPeriod   int
}

func DefaultConsumerQueueDefinitions() []ConsumerQueueDefinition {
	return []ConsumerQueueDefinition{
		{
			ConsumerName:      "specialist-validate-license",
			SubscribesToEvent: "specialist.created",
			MaxReceiveCount:   3,
			VisibilityTimeout: 30,
			RetentionPeriod:   1209600,
		},
		{
			ConsumerName:      "specialist-register-credential",
			SubscribesToEvent: "specialist.created",
			MaxReceiveCount:   3,
			VisibilityTimeout: 30,
			RetentionPeriod:   1209600,
		},
		{
			ConsumerName:      "specialist-update-data-repos",
			SubscribesToEvent: "specialist.updated",
			MaxReceiveCount:   3,
			VisibilityTimeout: 30,
			RetentionPeriod:   1209600,
		},
		{
			ConsumerName:      "specialist-send-welcome-email",
			SubscribesToEvent: "specialist.created",
			MaxReceiveCount:   3,
			VisibilityTimeout: 30,
			RetentionPeriod:   1209600,
		},
	}
}

func queueName(prefix, consumerName string) string {
	return fmt.Sprintf("%s-%s.fifo", prefix, consumerName)
}

func dlqName(prefix, consumerName string) string {
	return fmt.Sprintf("%s-%s-dlq.fifo", prefix, consumerName)
}

func EnsureConsumerQueues(ctx context.Context, client *sqs.Client, prefix string, definitions []ConsumerQueueDefinition) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	urls := make(map[string]string)

	for _, def := range definitions {
		dlq := dlqName(prefix, def.ConsumerName)
		dlqURL, err := createFIFOQueue(ctx, client, dlq, def.VisibilityTimeout, def.RetentionPeriod)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure DLQ %s: %w", dlq, err)
		}
		log.Printf("DLQ ready: %s", dlq)

		dlqArn, err := getQueueArn(ctx, client, dlqURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get DLQ ARN for %s: %w", dlq, err)
		}

		main := queueName(prefix, def.ConsumerName)
		redrivePolicy := fmt.Sprintf(`{"deadLetterTargetArn":"%s","maxReceiveCount":"%d"}`, dlqArn, def.MaxReceiveCount)

		mainURL, err := createFIFOQueueWithRedrive(ctx, client, main, def.VisibilityTimeout, def.RetentionPeriod, redrivePolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure queue %s: %w", main, err)
		}
		log.Printf("Queue ready: %s", main)

		urls[def.ConsumerName] = mainURL
	}

	return urls, nil
}

func createFIFOQueue(ctx context.Context, client *sqs.Client, name string, visibilityTimeout, retentionPeriod int) (string, error) {
	out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Attributes: map[string]string{
			string(types.QueueAttributeNameFifoQueue):                 "true",
			string(types.QueueAttributeNameContentBasedDeduplication): "true",
			string(types.QueueAttributeNameVisibilityTimeout):         fmt.Sprintf("%d", visibilityTimeout),
			string(types.QueueAttributeNameMessageRetentionPeriod):    fmt.Sprintf("%d", retentionPeriod),
		},
	})
	if err != nil {
		return "", err
	}
	return *out.QueueUrl, nil
}

func createFIFOQueueWithRedrive(ctx context.Context, client *sqs.Client, name string, visibilityTimeout, retentionPeriod int, redrivePolicy string) (string, error) {
	out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Attributes: map[string]string{
			string(types.QueueAttributeNameFifoQueue):                 "true",
			string(types.QueueAttributeNameContentBasedDeduplication): "true",
			string(types.QueueAttributeNameVisibilityTimeout):         fmt.Sprintf("%d", visibilityTimeout),
			string(types.QueueAttributeNameMessageRetentionPeriod):    fmt.Sprintf("%d", retentionPeriod),
			string(types.QueueAttributeNameRedrivePolicy):             redrivePolicy,
		},
	})
	if err != nil {
		return "", err
	}
	return *out.QueueUrl, nil
}

func getQueueArn(ctx context.Context, client *sqs.Client, queueURL string) (string, error) {
	out, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	})
	if err != nil {
		return "", err
	}

	arn, ok := out.Attributes[string(types.QueueAttributeNameQueueArn)]
	if !ok {
		return "", fmt.Errorf("QueueArn attribute not found")
	}
	return arn, nil
}

func ResolveQueueURL(ctx context.Context, client *sqs.Client, prefix, consumerName string) (string, error) {
	name := queueName(prefix, consumerName)
	out, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve queue URL for %s: %w", name, err)
	}
	return *out.QueueUrl, nil
}

func EventSuffix(eventName string) string {
	parts := strings.Split(eventName, ".")
	if len(parts) < 2 {
		return eventName
	}
	return parts[len(parts)-1]
}
