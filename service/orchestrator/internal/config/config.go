package config

import (
	"strings"
	"time"

	"github.com/Astheria23/jokiOrchestrator/shared/env"
)

// Config holds all the service configurations.
type Config struct {
	Port           string
	MongoURI       string
	MongoDBName    string
	MongoTimeout   time.Duration
	AppEnv         string
	OpenCodeGoKey  string
	LLMBaseURL     string
	LLMModel       string
	RequestTimeout time.Duration
	SimulateAgents bool
	AgentURLs      map[string]string
	JWTSecret      string
	ChatLimit      int
}

// Load reads config options from environment variables.
func Load() *Config {
	agentMap := map[string]string{
		"web_scraper":          env.GetString("AGENT_WEB_SCRAPER_URL", ""),
		"data_mining":          env.GetString("AGENT_DATA_MINING_URL", ""),
		"summarizer":           env.GetString("AGENT_SUMMARIZER_URL", ""),
		"outliner":             env.GetString("AGENT_OUTLINER_URL", ""),
		"translator":           env.GetString("AGENT_TRANSLATOR_URL", ""),
		"paraphrase":           env.GetString("AGENT_PARAPHRASE_URL", ""),
		"typo_checker":         env.GetString("AGENT_TYPO_CHECKER_URL", ""),
		"fact_checker":         env.GetString("AGENT_FACT_CHECKER_URL", ""),
		"literature_reviewer":  env.GetString("AGENT_LITERATURE_REVIEWER_URL", ""),
		"citation_reference":   env.GetString("AGENT_CITATION_REFERENCE_URL", ""),
		"qna_simulator":        env.GetString("AGENT_QNA_SIMULATOR_URL", ""),
		"math_calculator":      env.GetString("AGENT_MATH_CALCULATOR_URL", ""),
		"spatial_gis":          env.GetString("AGENT_SPATIAL_GIS_URL", ""),
		"requirement_analyzer": env.GetString("AGENT_REQUIREMENT_ANALYZER_URL", ""),
		"diagram_builder":      env.GetString("AGENT_DIAGRAM_BUILDER_URL", ""),
		"ppt_generator":        env.GetString("AGENT_PPT_GENERATOR_URL", ""),
		"pdf_formatter":        env.GetString("AGENT_PDF_FORMATTER_URL", ""),
		"programmer":           env.GetString("AGENT_PROGRAMMER_URL", ""),
		"pr_reviewer":          env.GetString("AGENT_PR_REVIEWER_URL", ""),
		"database_querier":     env.GetString("AGENT_DATABASE_QUERIER_URL", ""),
		"context_memory":       env.GetString("AGENT_CONTEXT_MEMORY_URL", ""),
		"supervisor":           env.GetString("AGENT_SUPERVISOR_URL", ""),
		"kesimpulan_saran":     env.GetString("AGENT_KESIMPULAN_SARAN_URL", ""),
		"prompt_generator":     env.GetString("AGENT_PROMPT_GENERATOR_URL", ""),
		"qa_bug_hunter":        env.GetString("AGENT_QA_BUG_HUNTER_URL", ""),
		"essay_writer":         env.GetString("AGENT_ESSAY_WRITER_URL", ""),
	}

	// Normalize agent keys
	normalizedMap := make(map[string]string)
	for k, v := range agentMap {
		normalizedMap[strings.ToLower(k)] = v
	}

	chatLimit := env.GetInt("CHAT_LIMIT", 5)
	if chatLimit < 1 {
		chatLimit = 5
	}

	return &Config{
		Port:           env.GetString("PORT", "8080"),
		MongoURI:       env.GetString("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:    env.GetString("MONGO_DB_NAME", "joki_tugas"),
		MongoTimeout:   time.Duration(env.GetInt("MONGO_TIMEOUT_SECONDS", 10)) * time.Second,
		AppEnv:         env.GetString("APP_ENV", "development"),
		OpenCodeGoKey:  env.GetString("OPENCODE_GO_API_KEY", ""),
		LLMBaseURL:     env.GetString("LLM_BASE_URL", "https://opencode.ai/zen/go/v1"),
		LLMModel:       env.GetString("LLM_MODEL", "deepseek-v4-flash"),
		RequestTimeout: time.Duration(env.GetInt("REQUEST_TIMEOUT_MS", 30000)) * time.Millisecond,
		SimulateAgents: env.GetBool("SIMULATE_AGENTS", false),
		AgentURLs:      normalizedMap,
		JWTSecret:      env.GetString("JWT_SECRET", "super_secret_jwt_key"),
		ChatLimit:      chatLimit,
	}
}
