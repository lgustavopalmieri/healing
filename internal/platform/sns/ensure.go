package sns

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type TopicDefinition struct {
	EventName string
	Suffix    string
}

type SubscriptionDefinition struct {
	EventName    string
	ConsumerName string
	QueueURL     string
}

func DefaultTopicDefinitions() []TopicDefinition {
	return []TopicDefinition{
		{EventName: "specialist.created", Suffix: "specialist-created"},
		{EventName: "specialist.updated", Suffix: "specialist-updated"},
		{EventName: "auth.credential.pending", Suffix: "auth-credential-pending"},
		{EventName: "auth.credential.activated", Suffix: "auth-credential-activated"},
	}
}

func topicName(prefix, suffix string) string {
	return fmt.Sprintf("%s-%s.fifo", prefix, suffix)
}

func EnsureTopics(ctx context.Context, client *awssns.Client, prefix string, defs []TopicDefinition) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	arns := make(map[string]string)

	for _, def := range defs {
		name := topicName(prefix, def.Suffix)
		arn, err := createFIFOTopic(ctx, client, name)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure topic %s: %w", name, err)
		}
		log.Printf("SNS topic ready: %s", name)
		arns[def.EventName] = arn
	}

	return arns, nil
}

func EnsureSubscriptions(ctx context.Context, snsClient *awssns.Client, sqsClient *awssqs.Client, topicARNs map[string]string, subs []SubscriptionDefinition) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, sub := range subs {
		topicARN, ok := topicARNs[sub.EventName]
		if !ok {
			return fmt.Errorf("no topic ARN for event %q required by consumer %q", sub.EventName, sub.ConsumerName)
		}
		queueARN, err := getQueueArn(ctx, sqsClient, sub.QueueURL)
		if err != nil {
			return fmt.Errorf("get queue arn for consumer %q: %w", sub.ConsumerName, err)
		}
		if err := setQueueSNSPolicy(ctx, sqsClient, sub.QueueURL, queueARN, topicARN); err != nil {
			return fmt.Errorf("set queue policy for consumer %q: %w", sub.ConsumerName, err)
		}
		if err := subscribeQueueToTopic(ctx, snsClient, topicARN, queueARN); err != nil {
			return fmt.Errorf("subscribe consumer %q to topic %s: %w", sub.ConsumerName, sub.EventName, err)
		}
		log.Printf("SNS subscription ready: %s -> %s", sub.EventName, sub.ConsumerName)
	}

	return nil
}

func createFIFOTopic(ctx context.Context, client *awssns.Client, name string) (string, error) {
	out, err := client.CreateTopic(ctx, &awssns.CreateTopicInput{
		Name: aws.String(name),
		Attributes: map[string]string{
			"FifoTopic":                 "true",
			"ContentBasedDeduplication": "true",
		},
	})
	if err != nil {
		return "", err
	}
	return *out.TopicArn, nil
}

func getQueueArn(ctx context.Context, client *awssqs.Client, queueURL string) (string, error) {
	out, err := client.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn},
	})
	if err != nil {
		return "", err
	}
	arn, ok := out.Attributes[string(sqstypes.QueueAttributeNameQueueArn)]
	if !ok {
		return "", fmt.Errorf("queue ARN attribute not found")
	}
	return arn, nil
}

func setQueueSNSPolicy(ctx context.Context, client *awssqs.Client, queueURL, queueARN, topicARN string) error {
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":    "Allow",
				"Principal": map[string]string{"Service": "sns.amazonaws.com"},
				"Action":    "sqs:SendMessage",
				"Resource":  queueARN,
				"Condition": map[string]any{
					"ArnEquals": map[string]string{"aws:SourceArn": topicARN},
				},
			},
		},
	}
	raw, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	_, err = client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl: aws.String(queueURL),
		Attributes: map[string]string{
			string(sqstypes.QueueAttributeNamePolicy): string(raw),
		},
	})
	return err
}

func subscribeQueueToTopic(ctx context.Context, client *awssns.Client, topicARN, queueARN string) error {
	_, err := client.Subscribe(ctx, &awssns.SubscribeInput{
		TopicArn:              aws.String(topicARN),
		Protocol:              aws.String("sqs"),
		Endpoint:              aws.String(queueARN),
		Attributes:            map[string]string{"RawMessageDelivery": "true"},
		ReturnSubscriptionArn: true,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	return nil
}
