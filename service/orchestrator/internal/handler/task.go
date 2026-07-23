package handler

import (
	"net/http"
	"strings"

	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/pipeline"
	"github.com/Astheria23/jokiOrchestrator/shared/database/queries"
	"github.com/Astheria23/jokiOrchestrator/shared/models"
	"github.com/Astheria23/jokiOrchestrator/shared/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TaskHandler handles HTTP and WebSocket requests for Tasks.
type TaskHandler struct {
	repo   *queries.TaskRepository
	runner *pipeline.PipelineRunner
}

// NewTaskHandler creates a new TaskHandler instance.
func NewTaskHandler(db *mongo.Database, runner *pipeline.PipelineRunner) *TaskHandler {
	return &TaskHandler{
		repo:   queries.NewTaskRepository(db),
		runner: runner,
	}
}

// CreateTask handles POST /api/tasks.
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var input struct {
		Prompt   string   `json:"prompt" binding:"required"`
		Pipeline []string `json:"pipeline"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	task := &models.Task{
		UserID:   userIDStr,
		Prompt:   input.Prompt,
		Pipeline: input.Pipeline,
		Status:   "pending",
	}

	if err := h.repo.Create(c.Request.Context(), task); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Trigger asynchronous runner execution
	h.runner.RunTask(task.ID.Hex(), task.Prompt)

	response.JSON(c, http.StatusCreated, task)
}

// GetTask handles GET /api/tasks/:id.
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "task ID is required")
		return
	}

	task, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	if task.UserID != "" && task.UserID != userIDStr {
		response.Error(c, http.StatusForbidden, "access denied")
		return
	}

	response.JSON(c, http.StatusOK, task)
}

// ListTasks handles GET /api/tasks.
func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	tasks, err := h.repo.FindByUserID(c.Request.Context(), userIDStr, 50)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, tasks)
}

// DecideTask handles POST /api/tasks/:id/decide { "action": "approve" | "cancel" }.
func (h *TaskHandler) DecideTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "task ID is required")
		return
	}

	var input struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	task, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	if task.UserID != "" && task.UserID != userIDStr {
		response.Error(c, http.StatusForbidden, "access denied")
		return
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	switch action {
	case "approve", "gas":
		if task.Status != "awaiting_approval" {
			response.Error(c, http.StatusConflict, "tugas ini sudah tidak menunggu persetujuan")
			return
		}
		if err := h.runner.ApproveAndRun(id, task.ConversationID, userIDStr); err != nil {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.JSON(c, http.StatusOK, gin.H{"status": "running", "taskId": id})
	case "cancel", "batal":
		// Allow cancel while awaiting approval OR mid-run (Stop button).
		if task.Status != "awaiting_approval" || task.Status == "running" {
			if err := h.runner.CancelTask(id, task.ConversationID, userIDStr); err != nil {
				response.Error(c, http.StatusConflict, err.Error())
				return
			}
			response.JSON(c, http.StatusOK, gin.H{"status": "cancelled", "taskId": id})
			return
		}
		response.Error(c, http.StatusConflict, "tugas ini sudah tidak bisa dibatalkan")
	default:
		response.Error(c, http.StatusBadRequest, "action harus approve atau cancel")
	}
}
