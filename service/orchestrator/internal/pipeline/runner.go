package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/config"
	"github.com/Astheria23/jokiOrchestrator/shared/agents"
	"github.com/Astheria23/jokiOrchestrator/shared/database/queries"
	"github.com/Astheria23/jokiOrchestrator/shared/logging"
	"github.com/Astheria23/jokiOrchestrator/shared/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// PipelineRunner handles background asynchronous execution of task plans.
type PipelineRunner struct {
	cfg      *config.Config
	repo     *queries.TaskRepository
	messages *queries.MessageRepository
	convs    *queries.ConversationRepository
	router   *LLMRouter
	client   *http.Client
	// cancels maps taskID -> cancel func for in-flight executeAgents context
	cancels sync.Map
}

// NewPipelineRunner creates a new PipelineRunner instance.
func NewPipelineRunner(cfg *config.Config, db *mongo.Database) *PipelineRunner {
	return &PipelineRunner{
		cfg:      cfg,
		repo:     queries.NewTaskRepository(db),
		messages: queries.NewMessageRepository(db),
		convs:    queries.NewConversationRepository(db),
		router: NewLLMRouter(
			cfg.OpenCodeGoKey,
			cfg.LLMBaseURL,
			cfg.LLMModel,
		),
		client: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

// WSNotification represents the progress update broadcast payload.
type WSNotification struct {
	Type    string      `json:"type"` // "progress", "success", "error", "skipped"
	Message string      `json:"message"`
	Task    models.Task `json:"task"`
}

// AgentRequestPayload represents standard payload schema.
type AgentRequestPayload struct {
	URL     string `json:"url"`
	Keyword string `json:"keyword"`
	RawText string `json:"raw_text"`
}

// AgentRequestMetadata represents sender details.
type AgentRequestMetadata struct {
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
}

// AgentRequest represents the standard POST body contract.
type AgentRequest struct {
	TaskID    string               `json:"task_id"`
	AgentType string               `json:"agent_type"`
	Payload   AgentRequestPayload  `json:"payload"`
	Metadata  AgentRequestMetadata `json:"metadata"`
}

// AgentResponse defines the success return body from agent.
type AgentResponse struct {
	Status string `json:"status"`
	TaskID string `json:"task_id"`
	Data   *struct {
		Result  string `json:"result"`
		FileURL string `json:"file_url"`
	} `json:"data"`
	Message string `json:"message"`
}

var urlRegex = regexp.MustCompile(`https?://[^\s]+`)

func extractURL(prompt string) string {
	return urlRegex.FindString(prompt)
}

func getAgentOutputType(agentName string) string {
	return agents.OutputType(agentName)
}

func isAgentInputCompatible(outputType string, nextAgentName string) bool {
	return agents.InputCompatible(outputType, nextAgentName)
}

func errorsIsCanceled(err error) bool {
	return err != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

// RunTask starts planning + (legacy) waits for approval only via chat flow.
// For direct API creates without chat, it plans then auto-executes if no conversation.
func (pr *PipelineRunner) RunTask(taskID string, prompt string) {
	pr.RunTaskForConversation(taskID, prompt, "", "")
}

// RunTaskForConversation plans the pipeline and waits for user approval when conversationID is set.
func (pr *PipelineRunner) RunTaskForConversation(taskID, prompt, conversationID, userID string) {
	go pr.planPipeline(taskID, prompt, conversationID, userID)
}

// ApproveAndRun starts agent execution after user approval.
func (pr *PipelineRunner) ApproveAndRun(taskID, conversationID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ok, err := pr.repo.CompareAndSetStatus(ctx, taskID, "awaiting_approval", "running")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("task is not awaiting approval")
	}

	task, err := pr.repo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}

	_ = pr.messages.SetApprovalByTaskID(ctx, taskID, models.ApprovalApproved)
	if conversationID != "" {
		ConversationHub.Broadcast(conversationID, ChatWSPayload{
			Type: "approval",
			Message: models.Message{
				TaskID:         taskID,
				Kind:           models.KindTaskPipeline,
				ApprovalStatus: models.ApprovalApproved,
				Pipeline:       task.Pipeline,
			},
		})
	}

	pr.appendChat(ctx, conversationID, userID, taskID, models.KindTaskProgress, "Oke, gas! Mulai dikerjakan…", nil)
	go pr.executeAgents(taskID, task.Prompt, conversationID, userID, task.Pipeline)
	return nil
}

// CancelTask cancels a task waiting for approval or currently running.
func (pr *PipelineRunner) CancelTask(taskID, conversationID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ok, err := pr.repo.CompareAndSetStatus(ctx, taskID, "awaiting_approval", "cancelled")
	if err != nil {
		return err
	}
	
	wasRunning := false
	if !ok {
		ok, err = pr.repo.CompareAndSetStatus(ctx, taskID, "running", "cancelled")
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("task cannot be cancelled (not awaiting approval or running)")
		}
		wasRunning = true
	}

	_ = pr.messages.SetApprovalByTaskID(ctx, taskID, models.ApprovalCancelled)

	// Abort in-flight HTTP to the current agent immediately (if running).
	if wasRunning {
		if v, ok := pr.cancels.Load(taskID); ok {
			if cancelFn, ok := v.(context.CancelFunc); ok {
				cancelFn()
			}
		}
	}

	content := "Oke, dibatalin. Kalau mau ganti rencana, tulis aja permintaan barunya."
	if wasRunning {
		content = "Oke, prosesnya aku stop secara paksa. Mending kita batalin di awal daripada ngabisin token. 🛑 Ada revisi apa nih?"
	}

	msg := &models.Message{
		ConversationID: conversationID,
		UserID:         userID,
		Role:           models.RoleAssistant,
		Content:        content,
		Kind:           models.KindTaskCancelled,
		TaskID:         taskID,
		ApprovalStatus: models.ApprovalCancelled,
	}
	if conversationID != "" {
		_ = pr.messages.Create(ctx, msg)
		_ = pr.convs.TouchUpdatedAt(ctx, conversationID)
		ConversationHub.Broadcast(conversationID, ChatWSPayload{Type: "message", Message: *msg})
		ConversationHub.Broadcast(conversationID, ChatWSPayload{
			Type: "approval",
			Message: models.Message{
				TaskID:         taskID,
				ApprovalStatus: models.ApprovalCancelled,
				Kind:           models.KindTaskPipeline,
			},
		})
	}
	return nil
}

func (pr *PipelineRunner) appendChat(ctx context.Context, conversationID, userID, taskID, kind, content string, pipeline []string) {
	pr.appendChatFull(ctx, conversationID, userID, taskID, kind, content, pipeline, "")
}

func (pr *PipelineRunner) appendChatFull(ctx context.Context, conversationID, userID, taskID, kind, content string, pipeline []string, approvalStatus string) {
	if conversationID == "" {
		return
	}
	if content == "" && len(pipeline) == 0 {
		return
	}
	msg := &models.Message{
		ConversationID: conversationID,
		UserID:         userID,
		Role:           models.RoleSystem,
		Content:        content,
		Kind:           kind,
		TaskID:         taskID,
		Pipeline:       pipeline,
		ApprovalStatus: approvalStatus,
	}
	switch kind {
	case models.KindTaskResult, models.KindTaskError, models.KindTaskCancelled:
		msg.Role = models.RoleAssistant
	case models.KindTaskPipeline:
		msg.Role = models.RoleAssistant
		if msg.Content == "" {
			msg.Content = "Ini langkah yang akan dijalankan:"
		}
	}
	if err := pr.messages.Create(ctx, msg); err != nil {
		logging.Log.Error().Err(err).Str("conversationId", conversationID).Msg("Failed to append chat message from pipeline")
		return
	}
	_ = pr.convs.TouchUpdatedAt(ctx, conversationID)
	ConversationHub.Broadcast(conversationID, ChatWSPayload{Type: "message", Message: *msg})
}

func (pr *PipelineRunner) planPipeline(taskID, prompt, conversationID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	logging.Log.Info().Str("taskId", taskID).Msg("Planning task pipeline")

	notify := func(typ, message string, task *models.Task) {
		if task != nil {
			GlobalHub.Broadcast(taskID, WSNotification{Type: typ, Message: message, Task: *task})
		}
		kind := models.KindTaskProgress
		if typ == "error" {
			kind = models.KindTaskError
		}
		pr.appendChat(ctx, conversationID, userID, taskID, kind, message, nil)
	}

	_ = pr.repo.UpdateStep(ctx, taskID, 0, "routing")
	task, err := pr.repo.FindByID(ctx, taskID)
	if err != nil {
		logging.Log.Error().Err(err).Str("taskId", taskID).Msg("Failed to load task from MongoDB")
		return
	}
	notify("progress", "Sedang menyusun langkahnya dulu…", task)

	pipelineSteps, err := pr.router.RoutePrompt(ctx, prompt)
	if err != nil || len(pipelineSteps) == 0 {
		if err == nil {
			err = fmt.Errorf("router returned empty step list")
		}
		logging.Log.Error().Err(err).Str("taskId", taskID).Msg("Routing failure")
		_ = pr.repo.UpdateFinalStatus(ctx, taskID, "failed", "")
		friendly := pr.router.FriendlyAgentError(ctx, "router", prompt)
		if t, dbErr := pr.repo.FindByID(ctx, taskID); dbErr == nil {
			notify("error", friendly, t)
		}
		return
	}

	logging.Log.Info().Str("taskId", taskID).Interface("pipeline", pipelineSteps).Msg("Routing complete — awaiting approval")

	if err := pr.repo.UpdatePipeline(ctx, taskID, pipelineSteps); err != nil {
		logging.Log.Error().Err(err).Str("taskId", taskID).Msg("Failed to update pipeline in database")
	}
	_ = pr.repo.UpdateFinalStatus(ctx, taskID, "awaiting_approval", "")

	pr.appendChatFull(
		ctx, conversationID, userID, taskID,
		models.KindTaskPipeline,
		"Ini langkah yang akan dijalankan. Cek dulu ya — setuju lanjut, atau batalin kalau mau ganti.",
		pipelineSteps,
		models.ApprovalAwaiting,
	)

	if t, dbErr := pr.repo.FindByID(ctx, taskID); dbErr == nil {
		GlobalHub.Broadcast(taskID, WSNotification{Type: "awaiting_approval", Message: "Menunggu persetujuan", Task: *t})
	}

	// Legacy path without chat: auto-approve and run
	if conversationID == "" {
		_ = pr.repo.UpdateFinalStatus(ctx, taskID, "running", "")
		pr.executeAgents(taskID, prompt, "", "", pipelineSteps)
	}
}

func (pr *PipelineRunner) executeAgents(taskID, prompt, conversationID, userID string, pipelineSteps []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	pr.cancels.Store(taskID, cancel)
	defer func() {
		cancel()
		pr.cancels.Delete(taskID)
	}()

	logging.Log.Info().Str("taskId", taskID).Msg("Executing approved pipeline")

	notify := func(typ, message string, task *models.Task) {
		if task != nil {
			GlobalHub.Broadcast(taskID, WSNotification{Type: typ, Message: message, Task: *task})
		}
		kind := models.KindTaskProgress
		if typ == "success" {
			kind = models.KindTaskResult
		} else if typ == "error" {
			kind = models.KindTaskError
		}
		pr.appendChat(ctx, conversationID, userID, taskID, kind, message, nil)
	}

	if len(pipelineSteps) == 0 {
		task, err := pr.repo.FindByID(ctx, taskID)
		if err == nil {
			pipelineSteps = task.Pipeline
		}
	}
	if len(pipelineSteps) == 0 {
		_ = pr.repo.UpdateFinalStatus(ctx, taskID, "failed", "")
		notify("error", "Rencana langkahnya kosong. Coba kirim ulang ya.", nil)
		return
	}

	currentPayloadText := prompt
	extractedURL := extractURL(prompt)
	currentType := "text"
	if extractedURL != "" {
		currentType = "url"
	}

	for i, agentName := range pipelineSteps {
		if ctx.Err() != nil {
			logging.Log.Info().Str("taskId", taskID).Msg("Task context cancelled — stopping execution")
			return
		}

		// Abort if task was cancelled mid-run (best-effort check)
		task, err := pr.repo.FindByID(ctx, taskID)
		if err == nil && task.Status == "cancelled" {
			logging.Log.Info().Str("taskId", taskID).Msg("Task cancelled — stopping execution")
			return
		}

		logging.Log.Info().Str("taskId", taskID).Str("agent", agentName).Msgf("Step %d of %d in progress", i+1, len(pipelineSteps))

		_ = pr.repo.UpdateStep(ctx, taskID, i+1, "running")
		task, err = pr.repo.FindByID(ctx, taskID)
		if err == nil {
			label := humanAgentLabel(agentName)
			notify("progress", fmt.Sprintf("Menjalankan %s (%d/%d)…", label, i+1, len(pipelineSteps)), task)
		}

		agentURL := pr.cfg.AgentURLs[strings.ToLower(agentName)]
		var respData *AgentResponse
		var execErr error

		if pr.cfg.SimulateAgents {
			respData, execErr = pr.simulateAgent(ctx, taskID, agentName, currentPayloadText)
		} else {
			if agentURL == "" {
				execErr = fmt.Errorf("agent endpoint URL is not configured")
			} else if strings.EqualFold(agentName, "web_scraper") {
				if extractURL(prompt) == "" && extractURL(currentPayloadText) == "" {
					notify("progress", "Sedang cari sumber di web dulu, baru di-scrape…", task)
				}
				respData, execErr = pr.scrapeOrchestrate(ctx, taskID, agentURL, prompt, currentPayloadText)
			} else {
				respData, execErr = pr.executeAgent(ctx, taskID, agentName, agentURL, extractedURL, currentPayloadText)
			}
		}

		if execErr != nil {
			if ctx.Err() != nil || errorsIsCanceled(execErr) {
				logging.Log.Info().Str("taskId", taskID).Str("agent", agentName).Msg("Step aborted by cancel")
				return
			}

			label := humanAgentLabel(agentName)
			canSkip := agents.ErrorPolicy(agentName) == agents.OnErrorSkip
			nextOK := false
			if canSkip {
				if i+1 < len(pipelineSteps) {
					nextOK = isAgentInputCompatible(currentType, pipelineSteps[i+1])
				} else {
					nextOK = true
				}
			}

			if canSkip && nextOK {
				logging.Log.Warn().Err(execErr).Str("taskId", taskID).Str("agent", agentName).Msg("Step failed — soft skip, continuing")
				historyItem := models.History{
					Step:      i + 1,
					AgentKey:  agentName,
					Status:    "skipped",
					Input:     AgentRequestPayload{URL: extractedURL, RawText: currentPayloadText},
					Error:     execErr.Error(),
					Timestamp: time.Now(),
				}
				_ = pr.repo.AppendHistory(ctx, taskID, historyItem)
				task, _ = pr.repo.FindByID(ctx, taskID)
				notify("progress", fmt.Sprintf("%s dilewati (%d/%d) — sedang bermasalah, lanjut…", label, i+1, len(pipelineSteps)), task)
				continue
			}

			logging.Log.Error().Err(execErr).Str("taskId", taskID).Str("agent", agentName).Msg("Step failed — hard stop (no further agents)")

			historyItem := models.History{
				Step:      i + 1,
				AgentKey:  agentName,
				Status:    "failed",
				Input:     AgentRequestPayload{URL: extractedURL, RawText: currentPayloadText},
				Error:     execErr.Error(),
				Timestamp: time.Now(),
			}
			_ = pr.repo.AppendHistory(ctx, taskID, historyItem)
			_ = pr.repo.UpdateFinalStatus(ctx, taskID, "failed", "")

			friendly := pr.router.FriendlyAgentError(ctx, agentName, prompt)
			task, err = pr.repo.FindByID(ctx, taskID)
			if err == nil {
				notify("error", friendly, task)
			} else {
				pr.appendChat(ctx, conversationID, userID, taskID, models.KindTaskError, friendly, nil)
			}
			return
		}

		logging.Log.Info().Str("taskId", taskID).Str("agent", agentName).Msg("Step succeeded")

		resultText := respData.Data.Result
		if respData.Data.FileURL != "" {
			resultText = respData.Data.FileURL
		}

		historyItem := models.History{
			Step:      i + 1,
			AgentKey:  agentName,
			Status:    "success",
			Input:     AgentRequestPayload{URL: extractedURL, RawText: currentPayloadText},
			Output:    respData.Data,
			Timestamp: time.Now(),
		}
		_ = pr.repo.AppendHistory(ctx, taskID, historyItem)
		currentPayloadText = resultText
		currentType = getAgentOutputType(agentName)
	}

	if ctx.Err() != nil {
		logging.Log.Info().Str("taskId", taskID).Msg("Pipeline aborted before completion (cancelled)")
		return
	}

	logging.Log.Info().Str("taskId", taskID).Msg("Pipeline execution completed successfully")
	_ = pr.repo.UpdateFinalStatus(ctx, taskID, "completed", currentPayloadText)
	task, err := pr.repo.FindByID(ctx, taskID)
	if err == nil {
		resultMsg := "Selesai! Hasilnya siap."
		if currentPayloadText != "" {
			resultMsg = currentPayloadText
		}
		notify("success", resultMsg, task)
	}
}

// executeAgent sends POST request to agent endpoint.
func (pr *PipelineRunner) executeAgent(ctx context.Context, taskID string, agentName string, agentURL string, url string, rawText string) (*AgentResponse, error) {
	reqPayload := AgentRequest{
		TaskID:    taskID,
		AgentType: agentName,
		Payload: AgentRequestPayload{
			URL:     url,
			Keyword: "",
			RawText: rawText,
		},
		Metadata: AgentRequestMetadata{
			Sender:    "orchestrator",
			Timestamp: time.Now().Unix(),
		},
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", agentURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http agent request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := pr.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http connection to agent failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent returned HTTP status %d: %s", resp.StatusCode, string(respBytes))
	}

	var agentResp AgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&agentResp); err != nil {
		return nil, fmt.Errorf("failed to decode agent response: %w", err)
	}

	if agentResp.Status != "success" || agentResp.Data == nil {
		return nil, fmt.Errorf("agent returned error status: %s", agentResp.Message)
	}

	return &agentResp, nil
}

// simulateAgent generates simulated mock agent outputs for local testing.
func (pr *PipelineRunner) simulateAgent(ctx context.Context, taskID string, agentName string, rawText string) (*AgentResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(1200 * time.Millisecond):
	}

	task, err := pr.repo.FindByID(ctx, taskID)
	prompt := ""
	if err == nil {
		prompt = strings.ToLower(task.Prompt)
	}

	// Trigger simulation failure if prompt contains "fail <agent_name>"
	failTrigger := "fail " + strings.ToLower(agentName)
	failTriggerUnderscore := "fail_" + strings.ToLower(agentName)
	if strings.Contains(prompt, failTrigger) || strings.Contains(prompt, failTriggerUnderscore) {
		return nil, fmt.Errorf("simulated offline failure: agent node '%s' is not responding (503 Service Unavailable)", agentName)
	}

	resultText := fmt.Sprintf("[Mock Output of %s] processed successfully based on input: '%s'", agentName, truncateText(rawText, 45))
	fileURL := ""

	outType := getAgentOutputType(agentName)
	if outType == "file (pptx)" || outType == "file (pdf)" || outType == "file (svg/png)" || strings.Contains(outType, "file") {
		ext := "pdf"
		if strings.Contains(outType, "pptx") {
			ext = "pptx"
		} else if strings.Contains(outType, "png") || strings.Contains(outType, "svg") {
			ext = "png"
		}
		fileURL = fmt.Sprintf("https://jokitugas.bananaunion.web.id/public/mocks/%s-%s.%s", taskID, agentName, ext)
		resultText = fmt.Sprintf("Simulated file generated by %s.", agentName)
	} else if agentName == "summarizer" {
		resultText = "- Ringkasan materi tugas akhir AI\n- Poin utama: MAS Orchestrator berfungsi sebagai API Gateway\n- Poin kedua: Koneksi WebSocket menyalurkan progress real-time"
	} else if agentName == "outliner" {
		resultText = "1. Bab I: Pendahuluan\n   1.1 Latar Belakang MAS\n2. Bab II: Arsitektur Orkestrasi\n   2.1 Gin Router & Goroutines\n3. Bab III: Implementasi Smart Skip"
	} else if agentName == "translator" {
		resultText = "[Terjemahan Indonesia] Ini adalah hasil translasi simulasi teks input."
	}

	return &AgentResponse{
		Status: "success",
		TaskID: taskID,
		Data: &struct {
			Result  string `json:"result"`
			FileURL string `json:"file_url"`
		}{
			Result:  resultText,
			FileURL: fileURL,
		},
		Message: "Simulated execution successful",
	}, nil
}

func truncateText(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
