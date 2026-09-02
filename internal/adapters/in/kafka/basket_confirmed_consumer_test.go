package kafka

import (
	"context"
	"delivery/internal/core/application/usecases/commands"
	"delivery/internal/generated/queues/basketeventspb"
	"errors"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type fakeProcessBasketConfirmedHandler struct {
	received commands.ProcessBasketConfirmedCommand
	err      error
}

func (f *fakeProcessBasketConfirmedHandler) Handle(_ context.Context, cmd commands.ProcessBasketConfirmedCommand) error {
	f.received = cmd
	return f.err
}

func TestBasketConfirmedMessageHandler_HandleMessage_Success(t *testing.T) {
	fakeHandler := &fakeProcessBasketConfirmedHandler{}
	handler := &basketConfirmedMessageHandler{processBasketConfirmedHandler: fakeHandler}

	basketID := uuid.New()
	event := &basketeventspb.BasketConfirmedIntegrationEvent{
		BasketId: basketID.String(),
		Address: &basketeventspb.Address{
			Street: "Main St",
		},
		Volume: 5,
	}

	payload, err := proto.Marshal(event)
	if err != nil {
		t.Fatalf("marshal proto: %v", err)
	}

	messageID := "basket.confirmed:0:42"
	err = handler.handleMessage(context.Background(), messageID, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fakeHandler.received.MessageID != messageID {
		t.Errorf("got MessageID %q, want %q", fakeHandler.received.MessageID, messageID)
	}
	if fakeHandler.received.OrderID != basketID {
		t.Errorf("got OrderID %s, want %s", fakeHandler.received.OrderID, basketID)
	}
	if fakeHandler.received.Street != "Main St" {
		t.Errorf("got Street %q, want %q", fakeHandler.received.Street, "Main St")
	}
	if fakeHandler.received.Volume != 5 {
		t.Errorf("got Volume %d, want %d", fakeHandler.received.Volume, 5)
	}
}

func TestBasketConfirmedMessageHandler_HandleMessage_InvalidProtobuf(t *testing.T) {
	fakeHandler := &fakeProcessBasketConfirmedHandler{}
	handler := &basketConfirmedMessageHandler{processBasketConfirmedHandler: fakeHandler}

	err := handler.handleMessage(context.Background(), "msg-1", []byte("invalid-proto"))
	if err == nil {
		t.Error("expected error for invalid protobuf, got nil")
	}
}

func TestBasketConfirmedMessageHandler_HandleMessage_InvalidBasketID(t *testing.T) {
	fakeHandler := &fakeProcessBasketConfirmedHandler{}
	handler := &basketConfirmedMessageHandler{processBasketConfirmedHandler: fakeHandler}

	event := &basketeventspb.BasketConfirmedIntegrationEvent{
		BasketId: "not-a-uuid",
		Address:  &basketeventspb.Address{Street: "Main St"},
		Volume:   1,
	}
	payload, _ := proto.Marshal(event)

	err := handler.handleMessage(context.Background(), "msg-1", payload)
	if err == nil {
		t.Error("expected error for invalid UUID basket id, got nil")
	}
}

func TestBasketConfirmedMessageHandler_HandleMessage_MissingAddress(t *testing.T) {
	fakeHandler := &fakeProcessBasketConfirmedHandler{}
	handler := &basketConfirmedMessageHandler{processBasketConfirmedHandler: fakeHandler}

	event := &basketeventspb.BasketConfirmedIntegrationEvent{
		BasketId: uuid.NewString(),
		Address:  nil,
		Volume:   1,
	}
	payload, _ := proto.Marshal(event)

	err := handler.handleMessage(context.Background(), "msg-1", payload)
	if err == nil {
		t.Error("expected error for nil address, got nil")
	}
}

func TestBasketConfirmedMessageHandler_HandleMessage_HandlerError(t *testing.T) {
	expectedErr := errors.New("database failure")
	fakeHandler := &fakeProcessBasketConfirmedHandler{err: expectedErr}
	handler := &basketConfirmedMessageHandler{processBasketConfirmedHandler: fakeHandler}

	event := &basketeventspb.BasketConfirmedIntegrationEvent{
		BasketId: uuid.NewString(),
		Address:  &basketeventspb.Address{Street: "Main St"},
		Volume:   1,
	}
	payload, _ := proto.Marshal(event)

	err := handler.handleMessage(context.Background(), "msg-1", payload)
	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}
