package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/Astheria23/jokiOrchestrator/shared/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MessageRepository manages chat messages.
type MessageRepository struct {
	collection *mongo.Collection
}

// NewMessageRepository creates a MessageRepository.
func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{collection: db.Collection("messages")}
}

// Create inserts a message.
func (r *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	if msg.ID.IsZero() {
		msg.ID = bson.NewObjectID()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	_, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

// SetApprovalByTaskID updates approval_status on the task_pipeline message for a task.
func (r *MessageRepository) SetApprovalByTaskID(ctx context.Context, taskID, approvalStatus string) error {
	if taskID == "" {
		return nil
	}
	_, err := r.collection.UpdateOne(ctx,
		bson.M{"task_id": taskID, "kind": models.KindTaskPipeline},
		bson.M{"$set": bson.M{"approval_status": approvalStatus}},
	)
	if err != nil {
		return fmt.Errorf("failed to set approval status: %w", err)
	}
	return nil
}

// ListByConversation returns messages oldest-first.
func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID string, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 200
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"conversation_id": conversationID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer cursor.Close(ctx)

	var list []models.Message
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}
	if list == nil {
		list = []models.Message{}
	}
	return list, nil
}

// ListRecentByConversation returns the newest N messages, then reverses to chronological order.
func (r *MessageRepository) ListRecentByConversation(ctx context.Context, conversationID string, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 8
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"conversation_id": conversationID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent messages: %w", err)
	}
	defer cursor.Close(ctx)

	var newestFirst []models.Message
	if err := cursor.All(ctx, &newestFirst); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}

	// reverse to chronological
	for i, j := 0, len(newestFirst)-1; i < j; i, j = i+1, j-1 {
		newestFirst[i], newestFirst[j] = newestFirst[j], newestFirst[i]
	}
	if newestFirst == nil {
		newestFirst = []models.Message{}
	}
	return newestFirst, nil
}
