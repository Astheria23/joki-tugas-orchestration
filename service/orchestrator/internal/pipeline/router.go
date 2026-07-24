package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Astheria23/jokiOrchestrator/shared/agents"
	"github.com/Astheria23/jokiOrchestrator/shared/logging"
)

// LLMRouter plans agent pipelines via an OpenAI-compatible chat completions API.
// Default target: OpenCode Go (DeepSeek V4 Flash).
type LLMRouter struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewLLMRouter instantiates a new LLMRouter.
func NewLLMRouter(apiKey, baseURL, model string) *LLMRouter {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://opencode.ai/zen/go/v1"
	}
	if model == "" {
		model = "deepseek-v4-flash"
	}
	return &LLMRouter{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// ChatMessage represents a single message in OpenAI format.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest defines the schema for the chat completions request.
type ChatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	ResponseFormat interface{}   `json:"response_format,omitempty"`
	Temperature    float64       `json:"temperature,omitempty"`
}

// ChatResponse defines the schema for the chat completions response.
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// RoutePrompt parses the prompt into a type-safe sequential array of agent names.
func (r *LLMRouter) RoutePrompt(ctx context.Context, prompt string) ([]string, error) {
	// 1. Intercept Pre-defined Templates (Fast-path without LLM)
	promptLower := strings.ToLower(strings.TrimSpace(prompt))
	if strings.HasPrefix(promptLower, "/template ") {
		parts := strings.SplitN(promptLower, " ", 2)
		if len(parts) == 2 {
			tmpl := strings.TrimSpace(parts[1])
			logging.Log.Info().Str("template", tmpl).Msg("Triggering hardcoded pipeline template")
			switch tmpl {
			case "joki_makalah":
				return []string{"web_scraper", "data_mining", "literature_reviewer", "essay_writer", "citation_reference", "typo_checker", "pdf_formatter"}, nil
			case "joki_koding":
				return []string{"prompt_generator", "programmer", "qa_bug_hunter", "supervisor"}, nil
			case "joki_presentasi":
				return []string{"web_scraper", "summarizer", "outliner", "ppt_generator"}, nil
			case "review_tugas":
				return []string{"typo_checker", "fact_checker", "supervisor", "kesimpulan_saran"}, nil
			case "analisis_data":
				return []string{"data_mining", "database_querier", "diagram_builder"}, nil
			}
		}
	}

	// 2. Normal LLM Routing if no template matched
	if r.apiKey == "" {
		logging.Log.Warn().Msg("OPENCODE_GO_API_KEY is not configured. Falling back to local heuristic routing.")
		return r.fallbackRoute(prompt), nil
	}

	systemPrompt := agents.RouterSystemPrompt()

	reqBody := ChatRequest{
		Model: r.model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0.2,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	endpoint := r.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		logging.Log.Error().Err(err).Str("endpoint", endpoint).Msg("LLM router connection failed, using fallback routing.")
		return r.fallbackRoute(prompt), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		logging.Log.Error().
			Int("status", resp.StatusCode).
			Str("endpoint", endpoint).
			Str("model", r.model).
			Str("body", string(respBytes)).
			Msg("LLM router returned error status, using fallback.")
		return r.fallbackRoute(prompt), nil
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response payload: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, errors.New("received empty choice set from llm router")
	}

	content := chatResp.Choices[0].Message.Content

	var pipelineResult struct {
		Pipeline []string `json:"pipeline"`
	}

	if err := json.Unmarshal([]byte(content), &pipelineResult); err != nil {
		contentClean := strings.TrimPrefix(strings.TrimSpace(content), "```json")
		contentClean = strings.TrimSuffix(contentClean, "```")
		contentClean = strings.TrimSpace(contentClean)
		if err := json.Unmarshal([]byte(contentClean), &pipelineResult); err != nil {
			return nil, fmt.Errorf("failed to parse router pipeline JSON: %w, content: %s", err, content)
		}
	}

	if len(pipelineResult.Pipeline) == 0 {
		logging.Log.Warn().Msg("LLM returned empty pipeline, using fallback routing.")
		return r.fallbackRoute(prompt), nil
	}

	logging.Log.Info().
		Str("model", r.model).
		Interface("pipeline", pipelineResult.Pipeline).
		Msg("LLM router resolved pipeline")

	return pipelineResult.Pipeline, nil
}

func (r *LLMRouter) fallbackRoute(prompt string) []string {
	promptLower := strings.ToLower(prompt)
	var pipeline []string

	if strings.Contains(promptLower, "scrap") || strings.Contains(promptLower, "web") || strings.Contains(promptLower, "http") {
		pipeline = append(pipeline, "web_scraper")
	}
	if strings.Contains(promptLower, "mine") || strings.Contains(promptLower, "extract") {
		pipeline = append(pipeline, "data_mining")
	}
	if strings.Contains(promptLower, "summarize") || strings.Contains(promptLower, "rangkum") || strings.Contains(promptLower, "ringkas") || len(pipeline) > 0 {
		pipeline = append(pipeline, "summarizer")
	}
	if strings.Contains(promptLower, "outline") || strings.Contains(promptLower, "struktur") {
		pipeline = append(pipeline, "outliner")
	}
	if strings.Contains(promptLower, "translate") || strings.Contains(promptLower, "terjemah") {
		pipeline = append(pipeline, "translator")
	}
	if strings.Contains(promptLower, "ppt") || strings.Contains(promptLower, "presentasi") || strings.Contains(promptLower, "slide") {
		pipeline = append(pipeline, "ppt_generator")
	}
	if strings.Contains(promptLower, "pdf") || strings.Contains(promptLower, "cetak") {
		pipeline = append(pipeline, "pdf_formatter")
	}
	if strings.Contains(promptLower, "code") || strings.Contains(promptLower, "program") {
		pipeline = append(pipeline, "programmer")
	}

	if len(pipeline) == 0 {
		pipeline = []string{"summarizer"}
	}

	logging.Log.Info().Str("prompt", prompt).Interface("pipeline", pipeline).Msg("Local fallback routing resolved")
	return pipeline
}
