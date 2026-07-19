package main

import (
	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/api"
	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/config"
	"github.com/Astheria23/jokiOrchestrator/shared/database"
	"github.com/Astheria23/jokiOrchestrator/shared/logging"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present (local/dev). Missing file is fine in production containers.
	_ = godotenv.Load()

	// 1. Load Configurations
	cfg := config.Load()

	logging.Log.Info().
		Str("env", cfg.AppEnv).
		Str("llmModel", cfg.LLMModel).
		Str("llmBaseURL", cfg.LLMBaseURL).
		Bool("hasOpenCodeKey", cfg.OpenCodeGoKey != "").
		Int("chatLimit", cfg.ChatLimit).
		Msg("Initializing Joki Tasks Orchestrator service...")

	// 2. Connect to MongoDB
	client, err := database.ConnectMongo(cfg.MongoURI, cfg.MongoTimeout)
	if err != nil {
		logging.Log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer func() {
		if err := database.DisconnectMongo(); err != nil {
			logging.Log.Error().Err(err).Msg("Error disconnecting from MongoDB")
		}
	}()

	db := client.Database(cfg.MongoDBName)
	logging.Log.Info().Str("db", cfg.MongoDBName).Msg("Successfully connected to MongoDB")

	// 3. Initialize and Start API Server
	server := api.NewServer(cfg, db)
	if err := server.Start(); err != nil {
		logging.Log.Fatal().Err(err).Msg("Server terminated unexpectedly")
	}
}
