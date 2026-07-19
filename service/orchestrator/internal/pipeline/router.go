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
	if r.apiKey == "" {
		logging.Log.Warn().Msg("OPENCODE_GO_API_KEY is not configured. Falling back to local heuristic routing.")
		return r.fallbackRoute(prompt), nil
	}

	systemPrompt := `You are the routing engine for the Joki Tugas Multi-Agent Orchestrator. 
Your task is to analyze the user's request and plan a sequential pipeline of agents to fulfill it.

Available agents are:
1. web_scraper (input: url -> output: text) - Clean text content from a URL.
2. data_mining (input: text|url -> output: text) - Extract key patterns from raw text or url.
3. summarizer (input: text -> output: text) - Summarize long text.
4. outliner (input: text -> output: outline) - Create outline structures.
5. translator (input: text -> output: text) - Translate text to other languages.
6. paraphrase (input: text -> output: text) - Rewrite text to remove plagiarism.
7. typo_checker (input: text -> output: text) - Correct typography/grammar.
8. fact_checker (input: text -> output: text) - Verify info facts.
9. literature_reviewer (input: text -> output: text) - Build literature reviews.
10. citation_reference (input: text -> output: text) - Build APA/IEEE citation records.
11. qna_simulator (input: text -> output: text) - Build Q&A lists from content.
12. math_calculator (input: text -> output: text) - Solve math equations.
13. spatial_gis (input: text -> output: text) - Process GIS/coordinates.
14. requirement_analyzer (input: text -> output: text) - Create SRS document.
15. diagram_builder (input: text -> output: file) - Build diagrams (flowcharts, SVG/PNG).
16. ppt_generator (input: text|outline -> output: file) - Build PowerPoint slide files.
17. pdf_formatter (input: text -> output: file) - Render text to PDF format.
18. programmer (input: text -> output: code) - Write computer code.
19. pr_reviewer (input: code -> output: text) - Review Pull Requests/code.
20. database_querier (input: text -> output: text) - Execute database queries.
21. context_memory (input: text -> output: text) - Retrieve/save conversation context memory.
22. supervisor (input: text|code -> output: text) - Quality control check of outputs.

Rules:
- Respond ONLY with a valid JSON object containing a "pipeline" key which is an array of strings representing the agent names in sequential order of execution.
- Example JSON response:
{
  "pipeline": ["web_scraper", "summarizer", "ppt_generator"]
}
- Do not include markdown formatting or extra text outside the JSON object.
- The pipeline should be a realistic, type-compatible sequence.`

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
