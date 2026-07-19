package handler

import (
	"net/http"
	"time"

	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/config"
	"github.com/Astheria23/jokiOrchestrator/shared/auth"
	"github.com/Astheria23/jokiOrchestrator/shared/database/queries"
	"github.com/Astheria23/jokiOrchestrator/shared/models"
	"github.com/Astheria23/jokiOrchestrator/shared/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AuthHandler handles HTTP requests for user registration and login.
type AuthHandler struct {
	repo *queries.UserRepository
	cfg  *config.Config
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(db *mongo.Database, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		repo: queries.NewUserRepository(db),
		cfg:  cfg,
	}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user already exists
	existingUser, err := h.repo.FindByUsername(c.Request.Context(), input.Username)
	if err == nil && existingUser != nil {
		response.Error(c, http.StatusBadRequest, "username already exists")
		return
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := &models.User{
		Username: input.Username,
		Password: hashedPassword,
	}

	if err := h.repo.CreateUser(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, user)
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.repo.FindByUsername(c.Request.Context(), input.Username)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if !auth.CheckPasswordHash(input.Password, user.Password) {
		response.Error(c, http.StatusUnauthorized, "invalid username or password")
		return
	}

	// Generate JWT Token (valid for 24 hours)
	token, err := auth.GenerateToken(user.ID.Hex(), user.Username, h.cfg.JWTSecret, 24*time.Hour)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}
