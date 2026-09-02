package inboxrepo

import (
	"context"
	"delivery/internal/core/ports"
	"delivery/internal/pkg/errs"
	"delivery/internal/pkg/inbox"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

type Repository struct{}

func NewRepository() ports.InboxRepository { return &Repository{} }

func (r *Repository) TryAdd(ctx context.Context, tx ports.Tx, messageID string) (bool, error) {
	if strings.TrimSpace(messageID) == "" {
		return false, errs.NewValueIsRequired("messageID")
	}
	message := inbox.Message{ID: messageID, ReceivedAt: time.Now().UTC()}
	result := tx.DB().WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&message)
	return result.RowsAffected == 1, result.Error
}
