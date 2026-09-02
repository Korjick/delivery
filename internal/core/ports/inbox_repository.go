package ports

import "context"

type InboxRepository interface {
	TryAdd(ctx context.Context, tx Tx, messageID string) (bool, error)
}
