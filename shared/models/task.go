package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// History represents the log of a specific agent/step execution in the orchestrator pipeline.
type History struct {
	Step      int         `bson:"step" json:"step"`
	AgentKey  string      `bson:"agent_key" json:"agentKey"`
	Status    string      `bson:"status" json:"status"` // e.g., "success", "skipped", "failed"
	Input     interface{} `bson:"input" json:"input"`
	Output    interface{} `bson:"output" json:"output"`
	Error     string      `bson:"error,omitempty" json:"error,omitempty"`
	Timestamp time.Time   `bson:"timestamp" json:"timestamp"`
}

// Task represents an orchestration job/pipeline in MongoDB.
type Task struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID         string        `bson:"user_id,omitempty" json:"userId,omitempty"`
	ConversationID string        `bson:"conversation_id,omitempty" json:"conversationId,omitempty"`
	Prompt         string        `bson:"prompt" json:"prompt"`
	Pipeline       []string      `bson:"pipeline" json:"pipeline"`
	CurrentStep    int           `bson:"current_step" json:"currentStep"`
	Status         string        `bson:"status" json:"status"` // e.g., "pending", "running", "completed", "failed"
	Result         string        `bson:"result,omitempty" json:"result,omitempty"`
	History        []History     `bson:"history" json:"history"`
	CreatedAt      time.Time     `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updatedAt"`
}
