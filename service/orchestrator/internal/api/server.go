package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/config"
	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/handler"
	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/pipeline"
	"github.com/Astheria23/jokiOrchestrator/shared/logging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Server sets up the HTTP server engine and routing.
type Server struct {
	router *gin.Engine
	cfg    *config.Config
	db     *mongo.Database
	runner *pipeline.PipelineRunner
}

// NewServer initializes and configures a Gin server.
func NewServer(cfg *config.Config, db *mongo.Database) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	// Custom logging middleware using our shared zerolog logger
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		logging.Log.Info().
			Int("status", c.Writer.Status()).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Str("ip", c.ClientIP()).
			Dur("latency", time.Since(start)).
			Msg("request processed")
	})

	runner := pipeline.NewPipelineRunner(cfg, db)

	server := &Server{
		router: router,
		cfg:    cfg,
		db:     db,
		runner: runner,
	}

	server.routes()

	return server
}

func (s *Server) routes() {
	taskHandler := handler.NewTaskHandler(s.db, s.runner)
	authHandler := handler.NewAuthHandler(s.db, s.cfg)
	chatHandler := handler.NewChatHandler(s.db, s.cfg, s.runner)

	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// WebSocket Task Progress Endpoint
	s.router.GET("/ws/tasks/:task_id", pipeline.HandleWS)
	s.router.GET("/ws/conversations/:id", pipeline.HandleConversationWS)

	api := s.router.Group("/api")
	{
		// Unprotected Auth routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(JWTAuthMiddleware(s.cfg.JWTSecret))
		{
			protected.GET("/me", chatHandler.GetMe)

			conversations := protected.Group("/conversations")
			{
				conversations.GET("", chatHandler.ListConversations)
				conversations.POST("", chatHandler.CreateConversation)
				conversations.GET("/:id", chatHandler.GetConversation)
				conversations.POST("/:id/messages", chatHandler.SendMessage)
			}

			tasksGroup := protected.Group("/tasks")
			{
				tasksGroup.GET("", taskHandler.ListTasks)
				tasksGroup.POST("", taskHandler.CreateTask)
				tasksGroup.GET("/:id", taskHandler.GetTask)
				tasksGroup.POST("/:id/decide", taskHandler.DecideTask)
			}
		}
	}
}

// CORSMiddleware handles cross-origin resource sharing requests.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	logging.Log.Info().Str("addr", addr).Msg("Starting orchestrator server")
	return s.router.Run(addr)
}
