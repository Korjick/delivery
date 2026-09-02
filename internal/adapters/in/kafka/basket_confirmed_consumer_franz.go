package kafka

import (
	"context"
	"delivery/internal/core/application/usecases/commands"
	"delivery/internal/pkg/errs"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

type basketConfirmedConsumer struct {
	topic          string
	client         *kgo.Client
	messageHandler *basketConfirmedMessageHandler
}

func NewBasketConfirmedConsumer(
	brokers []string,
	group string,
	topic string,
	processBasketConfirmedHandler commands.ProcessBasketConfirmedCommandHandler,
) (BasketConfirmedConsumer, error) {
	if len(brokers) == 0 || strings.TrimSpace(brokers[0]) == "" {
		return nil, errs.NewValueIsRequired("brokers")
	}
	if strings.TrimSpace(group) == "" {
		return nil, errs.NewValueIsRequired("group")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errs.NewValueIsRequired("topic")
	}
	if processBasketConfirmedHandler == nil {
		return nil, errs.NewValueIsRequired("processBasketConfirmedHandler")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create franz-go Kafka consumer: %w", err)
	}

	return &basketConfirmedConsumer{
		topic:          topic,
		client:         client,
		messageHandler: &basketConfirmedMessageHandler{processBasketConfirmedHandler: processBasketConfirmedHandler},
	}, nil
}

func (c *basketConfirmedConsumer) Consume() error {
	ctx := context.Background()
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fetchErr := range errs {
				if errors.Is(fetchErr.Err, context.Canceled) {
					return nil
				}
				log.Printf("Kafka consumer fetch error on topic %s partition %d: %v", fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
			}
			continue
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			messageID := fmt.Sprintf("%s:%d:%d", record.Topic, record.Partition, record.Offset)
			if err := c.messageHandler.handleMessage(ctx, messageID, record.Value); err != nil {
				log.Printf("process %s message at %d:%d: %v", record.Topic, record.Partition, record.Offset, err)
				continue
			}
			if err := c.client.CommitRecords(ctx, record); err != nil {
				log.Printf("commit %s record at %d:%d: %v", record.Topic, record.Partition, record.Offset, err)
			}
		}
	}
}

func (c *basketConfirmedConsumer) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	c.client.Close()
	return nil
}
