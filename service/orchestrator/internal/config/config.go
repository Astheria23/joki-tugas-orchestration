package config

import (
	"strings"
	"time"

	"github.com/Astheria23/jokiOrchestrator/shared/agents"
	"github.com/Astheria23/jokiOrchestrator/shared/env"
)

// Config holds all the service configurations.
type Config struct {
	Port              string
	MongoURI          string
	MongoDBName       string
	MongoTimeout      time.Duration
	AppEnv            string
	OpenCodeGoKey     string
	LLMBaseURL        string
	LLMModel          string
	RequestTimeout    time.Duration
	SimulateAgents    bool
	AgentURLs         map[string]string
	JWTSecret         string
	ChatLimit         int
	WebSearchProvider string // duckduckgo | brave | off
	BraveSearchKey    string
	WebSearchMaxResults int
}

// Load reads config options from environment variables.
func Load() *Config {
	agents.MustLoad()

	normalizedMap := make(map[string]string)
	for key, envName := range agents.EnvURLKeys() {
		if envName == "" {
			continue
		}
		normalizedMap[strings.ToLower(key)] = env.GetString(envName, "")
	}

	chatLimit := env.GetInt("CHAT_LIMIT", 5)
	if chatLimit < 1 {
		chatLimit = 5
	}

	maxResults := env.GetInt("WEB_SEARCH_MAX_RESULTS", 3)
	if maxResults < 1 {
		maxResults = 1
	}
	if maxResults > 8 {
		maxResults = 8
	}

	provider := strings.ToLower(strings.TrimSpace(env.GetString("WEB_SEARCH_PROVIDER", "")))
	braveKey := env.GetString("BRAVE_SEARCH_API_KEY", "")
	if provider == "" {
		if braveKey != "" {
			provider = "brave"
		} else {
			provider = "wikipedia"
		}
	}

	return &Config{
		Port:                env.GetString("PORT", "8080"),
		MongoURI:            env.GetString("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:         env.GetString("MONGO_DB_NAME", "joki_tugas"),
		MongoTimeout:        time.Duration(env.GetInt("MONGO_TIMEOUT_SECONDS", 10)) * time.Second,
		AppEnv:              env.GetString("APP_ENV", "development"),
		OpenCodeGoKey:       env.GetString("OPENCODE_GO_API_KEY", ""),
		LLMBaseURL:          env.GetString("LLM_BASE_URL", "https://opencode.ai/zen/go/v1"),
		LLMModel:            env.GetString("LLM_MODEL", "deepseek-v4-flash"),
		RequestTimeout:      time.Duration(env.GetInt("REQUEST_TIMEOUT_MS", 30000)) * time.Millisecond,
		SimulateAgents:      env.GetBool("SIMULATE_AGENTS", false),
		AgentURLs:           normalizedMap,
		JWTSecret:           env.GetString("JWT_SECRET", "super_secret_jwt_key"),
		ChatLimit:           chatLimit,
		WebSearchProvider:   provider,
		BraveSearchKey:      braveKey,
		WebSearchMaxResults: maxResults,
	}
}
