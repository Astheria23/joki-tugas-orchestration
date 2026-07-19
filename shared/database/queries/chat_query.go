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

// ConversationRepository manages conversation documents.
type ConversationRepository struct {
	collection *mongo.Collection
}

// NewConversationRepository creates a ConversationRepository.
func NewConversationRepository(db *mongo.Database) *ConversationRepository {
	return &ConversationRepository{collection: db.Collection("conversations")}
}

// Create inserts a conversation.
func (r *ConversationRepository) Create(ctx context.Context, conv *models.Conversation) error {
	if conv.ID.IsZero() {
		conv.ID = bson.NewObjectID()
	}
	now := time.Now()
	conv.CreatedAt = now
	conv.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, conv)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

// FindByID loads a conversation by hex ID.
func (r *ConversationRepository) FindByID(ctx context.Context, id string) (*models.Conversation, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid conversation ID: %w", err)
	}
	var conv models.Conversation
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&conv)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}
	return &conv, nil
}

// ListByUserID returns conversations for a user, newest first.
func (r *ConversationRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]models.Conversation, error) {
	if limit <= 0 {
		limit = 50
	}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer cursor.Close(ctx)

	var list []models.Conversation
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("failed to decode conversations: %w", err)
	}
	if list == nil {
		list = []models.Conversation{}
	}
	return list, nil
}

// TouchTitle updates title and updated_at.
func (r *ConversationRepository) TouchTitle(ctx context.Context, id, title string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid conversation ID: %w", err)
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"title":      title,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

// TouchUpdatedAt bumps updated_at.
func (r *ConversationRepository) TouchUpdatedAt(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid conversation ID: %w", err)
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{"updated_at": time.Now()},
	})
	return err
}
