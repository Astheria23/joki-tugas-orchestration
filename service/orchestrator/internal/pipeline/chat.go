package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Astheria23/jokiOrchestrator/shared/logging"
	"github.com/Astheria23/jokiOrchestrator/shared/models"
)

// ChatTurnResult is the structured LLM response for a chat turn.
type ChatTurnResult struct {
	Type       string `json:"type"` // "chat" | "task"
	Reply      string `json:"reply"`
	TaskPrompt string `json:"taskPrompt,omitempty"`
}

const chatSystemPrompt = `Kamu asisten Bananacademic untuk mahasiswa. Bahasa Indonesia, santai, jelas, tanpa jargon teknis.

Balas HANYA JSON valid (tanpa markdown):
{"type":"chat","reply":"..."} 
atau
{"type":"task","reply":"...","taskPrompt":"..."}

Pakai "task" hanya jika user minta KERJAKAN tugas konkret (ringkas, terjemah, PPT, PDF, scrape URL, buat kode, dll). taskPrompt = instruksi lengkap untuk dikerjakan.
Selain itu pakai "chat" (sapaan, tanya, klarifikasi).
Reply singkat dan membantu.`

// ChatTurn classifies and replies to a user message with short history.
func (r *LLMRouter) ChatTurn(ctx context.Context, history []models.Message, userText string) (*ChatTurnResult, error) {
	if r.apiKey == "" {
		return heuristicChatTurn(userText), nil
	}

	messages := []ChatMessage{
		{Role: "system", Content: chatSystemPrompt},
	}

	for _, m := range history {
		role := m.Role
		if role != "user" && role != "assistant" {
			continue
		}
		content := truncateRunes(m.Content, 600)
		if content == "" {
			continue
		}
		messages = append(messages, ChatMessage{Role: role, Content: content})
	}
	messages = append(messages, ChatMessage{Role: "user", Content: truncateRunes(userText, 2000)})

	reqBody := ChatRequest{
		Model:          r.model,
		Messages:       messages,
		ResponseFormat: map[string]string{"type": "json_object"},
		Temperature:    0.3,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal chat turn: %w", err)
	}

	endpoint := r.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.client.Do(req)
	if err != nil {
		logging.Log.Error().Err(err).Msg("ChatTurn LLM request failed, using heuristic")
		return heuristicChatTurn(userText), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		logging.Log.Error().Int("status", resp.StatusCode).Str("body", string(b)).Msg("ChatTurn LLM error, using heuristic")
		return heuristicChatTurn(userText), nil
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}
	if len(chatResp.Choices) == 0 {
		return heuristicChatTurn(userText), nil
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result ChatTurnResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		logging.Log.Warn().Err(err).Str("content", content).Msg("ChatTurn JSON parse failed")
		return &ChatTurnResult{Type: "chat", Reply: content}, nil
	}

	result.Type = strings.ToLower(strings.TrimSpace(result.Type))
	if result.Type != "chat" && result.Type != "task" {
		result.Type = "chat"
	}
	if result.Reply == "" {
		result.Reply = "Oke, aku dengar."
	}
	if result.Type == "task" && strings.TrimSpace(result.TaskPrompt) == "" {
		result.TaskPrompt = userText
	}

	return &result, nil
}

func heuristicChatTurn(userText string) *ChatTurnResult {
	lower := strings.ToLower(userText)
	taskHints := []string{
		"kerjakan", "ringkas", "rangkum", "terjemah", "ppt", "presentasi", "slide",
		"pdf", "scrape", "scrap", "http://", "https://", "buatkan", "tuliskan kode",
		"parafrase", "sitasi", "outline",
	}
	for _, h := range taskHints {
		if strings.Contains(lower, h) {
			return &ChatTurnResult{
				Type:       "task",
				Reply:      "Siap, aku kerjakan dulu ya. Nanti hasilnya muncul di sini.",
				TaskPrompt: userText,
			}
		}
	}
	return &ChatTurnResult{
		Type:  "chat",
		Reply: "Hai! Ceritain aja tugas kuliah yang mau dibantu — misalnya ringkas materi, bikin slide, atau terjemahin teks.",
	}
}

func truncateRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}

var agentDisplayNames = map[string]string{
	"web_scraper":          "pengambil konten web",
	"data_mining":          "analisis data",
	"summarizer":           "perangkum",
	"outliner":             "pembuat kerangka",
	"translator":           "penerjemah",
	"paraphrase":           "parafrase",
	"typo_checker":         "pemeriksa ejaan",
	"fact_checker":         "pemeriksa fakta",
	"literature_reviewer":  "review literatur",
	"citation_reference":   "pembuat sitasi",
	"qna_simulator":        "simulator tanya-jawab",
	"math_calculator":      "penyelesai soal matematika",
	"spatial_gis":          "pemroses data peta",
	"requirement_analyzer": "analisis kebutuhan",
	"diagram_builder":      "pembuat diagram",
	"ppt_generator":        "pembuat presentasi / PPT",
	"pdf_formatter":        "pembuat PDF",
	"programmer":           "penulis kode",
	"pr_reviewer":          "pemeriksa kode",
	"database_querier":     "query database",
	"context_memory":       "penyimpan konteks",
	"supervisor":           "pemeriksa hasil akhir",
}

func humanAgentLabel(agentKey string) string {
	if name, ok := agentDisplayNames[strings.ToLower(agentKey)]; ok {
		return name
	}
	return strings.ReplaceAll(agentKey, "_", " ")
}

// FriendlyAgentError asks the LLM for a short, human apology when an agent fails.
// Falls back to a static Indonesian message if the LLM is unavailable.
func (r *LLMRouter) FriendlyAgentError(ctx context.Context, agentKey, userTask string) string {
	label := humanAgentLabel(agentKey)
	fallback := fmt.Sprintf(
		"Mohon maaf nih, engine buat %s lagi error / belum siap. Coba lagi nanti ya!",
		label,
	)

	if r.apiKey == "" {
		return fallback
	}

	system := `Kamu asisten Bananacademic. Bahasa Indonesia santai. Tanpa jargon teknis (jangan sebut agent, API, HTTP, pipeline).
Tulis 1-2 kalimat singkat: minta maaf karena fitur/mesin untuk pekerjaan tertentu sedang bermasalah, sarankan coba lagi nanti.
Balas HANYA teks biasa, bukan JSON.`

	user := fmt.Sprintf(
		"Fitur yang gagal: %s (%s).\nPermintaan user: %s",
		label,
		agentKey,
		truncateRunes(userTask, 400),
	)

	reqBody := ChatRequest{
		Model: r.model,
		Messages: []ChatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.4,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fallback
	}

	endpoint := r.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fallback
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	// Short dedicated client so we don't block the pipeline too long on apology generation
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logging.Log.Warn().Err(err).Msg("FriendlyAgentError LLM failed, using fallback copy")
		return fallback
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fallback
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil || len(chatResp.Choices) == 0 {
		return fallback
	}

	text := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	if text == "" {
		return fallback
	}
	return text
}

