package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Conversation is a chat thread owned by a user.
type Conversation struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    string        `bson:"user_id" json:"userId"`
	Title     string        `bson:"title" json:"title"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updatedAt"`
}

// Message roles and kinds for chat UI.
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"

	KindChat         = "chat"
	KindTaskAck      = "task_ack"
	KindTaskProgress = "task_progress"
	KindTaskPipeline = "task_pipeline"
	KindTaskResult   = "task_result"
	KindTaskError    = "task_error"
	KindTaskCancelled = "task_cancelled"

	ApprovalAwaiting  = "awaiting"
	ApprovalApproved  = "approved"
	ApprovalCancelled = "cancelled"
)

// Message is a single chat message inside a conversation.
type Message struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ConversationID string        `bson:"conversation_id" json:"conversationId"`
	UserID         string        `bson:"user_id" json:"userId"`
	Role           string        `bson:"role" json:"role"`
	Content        string        `bson:"content" json:"content"`
	Kind           string        `bson:"kind,omitempty" json:"kind,omitempty"`
	TaskID         string        `bson:"task_id,omitempty" json:"taskId,omitempty"`
	Pipeline       []string      `bson:"pipeline,omitempty" json:"pipeline,omitempty"`
	ApprovalStatus string        `bson:"approval_status,omitempty" json:"approvalStatus,omitempty"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
}
