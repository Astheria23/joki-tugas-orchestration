package queries

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Astheria23/jokiOrchestrator/shared/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// UserRepository manages MongoDB operations for User documents.
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository initializes a new UserRepository.
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// CreateUser inserts a new User document into MongoDB.
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if user.ID.IsZero() {
		user.ID = bson.NewObjectID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByUsername retrieves a User document by username.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	return &user, nil
}

// FindByID retrieves a User by hex ObjectID.
func (r *UserRepository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	var user models.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	return &user, nil
}

// ErrChatQuotaExceeded is returned when the lifetime chat limit is reached.
var ErrChatQuotaExceeded = errors.New("chat quota exceeded")

// TryConsumeChatSlot atomically increments chat_used if under limit.
// Returns the updated user on success.
func (r *UserRepository) TryConsumeChatSlot(ctx context.Context, userID string, limit int) (*models.User, error) {
	if limit < 1 {
		limit = models.DefaultChatLimit
	}
	objID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	filter := bson.M{
		"_id": objID,
		"$or": []bson.M{
			{"chat_used": bson.M{"$lt": limit}},
			{"chat_used": bson.M{"$exists": false}},
		},
	}
	update := bson.M{
		"$inc": bson.M{"chat_used": 1},
	}

	var user models.User
	err = r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrChatQuotaExceeded
		}
		return nil, fmt.Errorf("failed to consume chat slot: %w", err)
	}
	return &user, nil
}
