package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DefaultChatLimit is the production default for lifetime user messages per account.
const DefaultChatLimit = 5

// User represents a system user registered in the platform.
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string        `bson:"username" json:"username"`
	Password  string        `bson:"password" json:"-"`
	ChatUsed  int           `bson:"chat_used" json:"chatUsed"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
}
