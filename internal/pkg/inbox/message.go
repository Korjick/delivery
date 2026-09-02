package inbox

import "time"

type Message struct {
	ID         string `gorm:"primaryKey;type:varchar(255)"`
	ReceivedAt time.Time
}

func (Message) TableName() string { return "inbox" }
