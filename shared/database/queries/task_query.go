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

// TaskRepository manages MongoDB operations for Task documents.
type TaskRepository struct {
	collection *mongo.Collection
}

// NewTaskRepository initializes a new TaskRepository.
func NewTaskRepository(db *mongo.Database) *TaskRepository {
	return &TaskRepository{
		collection: db.Collection("tasks"),
	}
}

// Create inserts a new Task document into MongoDB.
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	if task.ID.IsZero() {
		task.ID = bson.NewObjectID()
	}
	if task.History == nil {
		task.History = []models.History{}
	}
	if task.Pipeline == nil {
		task.Pipeline = []string{}
	}
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// FindByID retrieves a Task document by its hex string ID.
func (r *TaskRepository) FindByID(ctx context.Context, taskID string) (*models.Task, error) {
	objID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID format: %w", err)
	}

	var task models.Task
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to find task by ID: %w", err)
	}

	return &task, nil
}

// UpdateStep updates the current step and status of a Task.
func (r *TaskRepository) UpdateStep(ctx context.Context, taskID string, step int, status string) error {
	objID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("invalid task ID format: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"current_step": step,
			"status":       status,
			"updated_at":   time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to update task step: %w", err)
	}
	return nil
}

// AppendHistory appends a new execution history item to the Task.
func (r *TaskRepository) AppendHistory(ctx context.Context, taskID string, history models.History) error {
	objID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("invalid task ID format: %w", err)
	}
	if history.Timestamp.IsZero() {
		history.Timestamp = time.Now()
	}

	// Ensure history is an array (legacy docs may have null — $push would fail).
	_, _ = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID, "history": bson.M{"$type": "null"}},
		bson.M{"$set": bson.M{"history": bson.A{}}},
	)
	_, _ = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID, "history": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"history": bson.A{}}},
	)

	update := bson.M{
		"$push": bson.M{
			"history": history,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to append history to task: %w", err)
	}
	return nil
}

// UpdateFinalStatus updates the final status and result/file_url of a Task.
func (r *TaskRepository) UpdateFinalStatus(ctx context.Context, taskID string, status string, result string) error {
	objID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("invalid task ID format: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"result":     result,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to update final task status: %w", err)
	}
	return nil
}

// UpdatePipeline updates the pipeline list of a Task.
func (r *TaskRepository) UpdatePipeline(ctx context.Context, taskID string, pipeline []string) error {
	objID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("invalid task ID format: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"pipeline":   pipeline,
			"updated_at": time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("failed to update pipeline: %w", err)
	}
	return nil
}

// UpdateStatus sets task status (and optional result).
func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID, status string) error {
	return r.UpdateFinalStatus(ctx, taskID, status, "")
}

// CompareAndSetStatus updates status only if current status matches expected.
func (r *TaskRepository) CompareAndSetStatus(ctx context.Context, taskID, fromStatus, toStatus string) (bool, error) {
	objID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return false, fmt.Errorf("invalid task ID format: %w", err)
	}
	res, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objID, "status": fromStatus},
		bson.M{"$set": bson.M{"status": toStatus, "updated_at": time.Now()}},
	)
	if err != nil {
		return false, fmt.Errorf("failed to compare-and-set status: %w", err)
	}
	return res.ModifiedCount > 0, nil
}

// FindAll retrieves all Task documents, sorted by CreatedAt descending.
func (r *TaskRepository) FindAll(ctx context.Context, limit int) ([]models.Task, error) {
	return r.FindByUserID(ctx, "", limit)
}

// FindByUserID retrieves Task documents for a user, sorted by CreatedAt descending.
// If userID is empty, returns all tasks (admin/legacy behavior).
func (r *TaskRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]models.Task, error) {
	if limit <= 0 {
		limit = 50
	}
	filter := bson.M{}
	if userID != "" {
		filter["user_id"] = userID
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err = cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks list: %w", err)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}
	return tasks, nil
}
