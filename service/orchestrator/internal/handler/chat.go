package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/config"
	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/pipeline"
	"github.com/Astheria23/jokiOrchestrator/shared/database/queries"
	"github.com/Astheria23/jokiOrchestrator/shared/logging"
	"github.com/Astheria23/jokiOrchestrator/shared/models"
	"github.com/Astheria23/jokiOrchestrator/shared/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ChatHandler serves conversation/chat endpoints.
type ChatHandler struct {
	users         *queries.UserRepository
	conversations *queries.ConversationRepository
	messages      *queries.MessageRepository
	tasks         *queries.TaskRepository
	runner        *pipeline.PipelineRunner
	router        *pipeline.LLMRouter
	cfg           *config.Config
}

// NewChatHandler creates a ChatHandler.
func NewChatHandler(db *mongo.Database, cfg *config.Config, runner *pipeline.PipelineRunner) *ChatHandler {
	return &ChatHandler{
		users:         queries.NewUserRepository(db),
		conversations: queries.NewConversationRepository(db),
		messages:      queries.NewMessageRepository(db),
		tasks:         queries.NewTaskRepository(db),
		runner:        runner,
		router: pipeline.NewLLMRouter(
			cfg.OpenCodeGoKey,
			cfg.LLMBaseURL,
			cfg.LLMModel,
		),
		cfg: cfg,
	}
}

// GetMe handles GET /api/me.
func (h *ChatHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	username, _ := c.Get("username")
	usernameStr, _ := username.(string)

	user, err := h.users.FindByID(c.Request.Context(), userIDStr)
	chatUsed := 0
	if err == nil && user != nil {
		chatUsed = user.ChatUsed
		usernameStr = user.Username
	}

	response.JSON(c, http.StatusOK, gin.H{
		"id":        userIDStr,
		"username":  usernameStr,
		"chatUsed":  chatUsed,
		"chatLimit": h.cfg.ChatLimit,
	})
}

// ListConversations handles GET /api/conversations.
func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	list, err := h.conversations.ListByUserID(c.Request.Context(), userIDStr, 50)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, list)
}

// CreateConversation handles POST /api/conversations.
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)

	var input struct {
		Title string `json:"title"`
	}
	_ = c.ShouldBindJSON(&input)

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Chat baru"
	}

	conv := &models.Conversation{
		UserID: userIDStr,
		Title:  truncateTitle(title, 60),
	}
	if err := h.conversations.Create(c.Request.Context(), conv); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, conv)
}

// GetConversation handles GET /api/conversations/:id.
func (h *ChatHandler) GetConversation(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	id := c.Param("id")

	conv, err := h.conversations.FindByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	if conv.UserID != userIDStr {
		response.Error(c, http.StatusForbidden, "access denied")
		return
	}

	msgs, err := h.messages.ListByConversation(c.Request.Context(), id, 200)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"conversation": conv,
		"messages":     msgs,
	})
}

// SendMessage handles POST /api/conversations/:id/messages.
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr, _ := userID.(string)
	convID := c.Param("id")

	var input struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		response.Error(c, http.StatusBadRequest, "pesan tidak boleh kosong")
		return
	}

	conv, err := h.conversations.FindByID(c.Request.Context(), convID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}
	if conv.UserID != userIDStr {
		response.Error(c, http.StatusForbidden, "access denied")
		return
	}

	// Hard limit: consume slot BEFORE LLM (absolute)
	updatedUser, err := h.users.TryConsumeChatSlot(c.Request.Context(), userIDStr, h.cfg.ChatLimit)
	if err != nil {
		if err == queries.ErrChatQuotaExceeded {
			response.Error(c, http.StatusTooManyRequests, fmt.Sprintf("Kuota chat habis (%d/%d)", h.cfg.ChatLimit, h.cfg.ChatLimit))
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	userMsg := &models.Message{
		ConversationID: convID,
		UserID:         userIDStr,
		Role:           models.RoleUser,
		Content:        content,
		Kind:           models.KindChat,
	}
	if err := h.messages.Create(c.Request.Context(), userMsg); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Title from first user message if still default
	if conv.Title == "Chat baru" || conv.Title == "" {
		_ = h.conversations.TouchTitle(c.Request.Context(), convID, truncateTitle(content, 60))
	} else {
		_ = h.conversations.TouchUpdatedAt(c.Request.Context(), convID)
	}

	history, _ := h.messages.ListRecentByConversation(c.Request.Context(), convID, 8)
	// exclude the just-added user message from history window for model (it's passed separately)
	var prior []models.Message
	for _, m := range history {
		if m.ID != userMsg.ID {
			prior = append(prior, m)
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	turn, err := h.router.ChatTurn(ctx, prior, content)
	if err != nil {
		logging.Log.Error().Err(err).Msg("ChatTurn failed")
		turn = &pipeline.ChatTurnResult{
			Type:  "chat",
			Reply: "Maaf, ada gangguan sebentar. Coba kirim lagi ya.",
		}
	}

	assistantKind := models.KindChat
	var taskID string

	if turn.Type == "task" {
		assistantKind = models.KindTaskAck
		task := &models.Task{
			UserID:         userIDStr,
			ConversationID: convID,
			Prompt:         turn.TaskPrompt,
			Status:         "pending",
		}
		if err := h.tasks.Create(c.Request.Context(), task); err != nil {
			logging.Log.Error().Err(err).Msg("Failed to create task from chat")
		} else {
			taskID = task.ID.Hex()
			h.runner.RunTaskForConversation(taskID, turn.TaskPrompt, convID, userIDStr)
		}
	}

	assistantMsg := &models.Message{
		ConversationID: convID,
		UserID:         userIDStr,
		Role:           models.RoleAssistant,
		Content:        turn.Reply,
		Kind:           assistantKind,
		TaskID:         taskID,
	}
	if err := h.messages.Create(c.Request.Context(), assistantMsg); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	_ = h.conversations.TouchUpdatedAt(c.Request.Context(), convID)

	response.JSON(c, http.StatusCreated, gin.H{
		"userMessage":      userMsg,
		"assistantMessage": assistantMsg,
		"chatUsed":         updatedUser.ChatUsed,
		"chatLimit":        h.cfg.ChatLimit,
		"taskId":           taskID,
	})
}

func truncateTitle(s string, max int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}
