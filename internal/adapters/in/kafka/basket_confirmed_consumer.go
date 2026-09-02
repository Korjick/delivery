package kafka

import (
	"context"
	"delivery/internal/core/application/usecases/commands"
	"delivery/internal/generated/queues/basketeventspb"
	"delivery/internal/pkg/errs"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type BasketConfirmedConsumer interface {
	Consume() error
	Close() error
}

type basketConfirmedMessageHandler struct {
	processBasketConfirmedHandler commands.ProcessBasketConfirmedCommandHandler
}

func (c *basketConfirmedMessageHandler) handleMessage(ctx context.Context, messageID string, payload []byte) error {
	var event basketeventspb.BasketConfirmedIntegrationEvent
	if err := proto.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode BasketConfirmedIntegrationEvent: %w", err)
	}

	orderID, err := uuid.Parse(event.GetBasketId())
	if err != nil {
		return fmt.Errorf("parse basket ID: %w", err)
	}
	if event.GetAddress() == nil {
		return errs.NewValueIsRequired("address")
	}

	return c.processBasketConfirmedHandler.Handle(ctx, commands.ProcessBasketConfirmedCommand{
		MessageID: messageID,
		OrderID:   orderID,
		Street:    event.GetAddress().GetStreet(),
		Volume:    int(event.GetVolume()),
	})
}
