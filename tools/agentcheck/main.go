package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Astheria23/jokiOrchestrator/shared/agents"
	"github.com/Astheria23/jokiOrchestrator/shared/env"
	"github.com/joho/godotenv"
)

type checkResult struct {
	Key      string
	Label    string
	URL      string
	OK       bool
	Skipped  bool
	HTTPCode int
	Detail   string
	Elapsed  time.Duration
}

// Domains / host fragments treated as placeholders (not real team endpoints).
var placeholderHosts = []string{
	"example.com",
	"localhost",
	"127.0.0.1",
	"math-calc.herokuapp.com",
	"gis-spatial.com",
	"fadel-pr.com",
	"context.vercel.app",
}

func main() {
	_ = godotenv.Load()
	agents.MustLoad()

	timeout := time.Duration(env.GetInt("REQUEST_TIMEOUT_MS", 30000)) * time.Millisecond
	if timeout < 5*time.Second {
		timeout = 20 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	markdown := false
	for _, a := range os.Args[1:] {
		if a == "-markdown" || a == "--markdown" {
			markdown = true
		}
	}

	var results []checkResult
	failCount := 0
	skipCount := 0
	okCount := 0

	for _, agent := range agents.All() {
		rawURL := strings.TrimSpace(env.GetString(agent.EnvURL, ""))
		res := checkResult{Key: agent.Key, Label: agent.Label, URL: rawURL}

		if reason, skip := shouldSkipURL(rawURL, agent.EnvURL); skip {
			res.Skipped = true
			res.Detail = reason
			skipCount++
			results = append(results, res)
			continue
		}

		start := time.Now()
		ok, code, detail := pingAgent(client, agent.Key, rawURL)
		res.Elapsed = time.Since(start)
		res.OK = ok
		res.HTTPCode = code
		res.Detail = detail
		if ok {
			okCount++
		} else {
			failCount++
		}
		results = append(results, res)
	}

	if markdown {
		printMarkdown(results)
	} else {
		printTable(results)
	}

	fmt.Printf("\nSummary: %d ok / %d fail / %d skipped / %d total\n", okCount, failCount, skipCount, len(results))
	if failCount > 0 {
		os.Exit(1)
	}
}

func shouldSkipURL(raw, envName string) (string, bool) {
	if raw == "" {
		return "SKIP — URL env kosong (" + envName + ")", true
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "SKIP — URL tidak valid / placeholder", true
	}
	host := strings.ToLower(u.Hostname())
	for _, p := range placeholderHosts {
		if host == p || strings.HasSuffix(host, "."+p) {
			return "SKIP — placeholder host (" + host + ")", true
		}
	}
	if strings.Contains(host, "example") {
		return "SKIP — placeholder host (" + host + ")", true
	}
	return "", false
}

func pingAgent(client *http.Client, agentKey, agentURL string) (bool, int, string) {
	rawText := "Ini teks percobaan kontrak orchestrator (task_id, agent_type, payload.raw_text)."
	pageURL := "https://example.com"
	// web_scraper only accepts real http(s) targets — never send plain prose.
	if agentKey == "web_scraper" {
		pageURL = "https://id.wikipedia.org/wiki/Indonesia"
		rawText = pageURL
	}
	payload := map[string]any{
		"task_id":    "agentcheck-" + time.Now().Format("150405"),
		"agent_type": agentKey,
		"payload": map[string]string{
			"url":      pageURL,
			"keyword":  "",
			"raw_text": rawText,
		},
		"metadata": map[string]any{
			"sender":    "agentcheck",
			"timestamp": time.Now().Unix(),
		},
	}
	body, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, agentURL, bytes.NewReader(body))
	if err != nil {
		return false, 0, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err.Error()
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2 MiB — scraper can return large pages
	if resp.StatusCode != http.StatusOK {
		return false, resp.StatusCode, truncate(string(raw), 240)
	}

	var decoded struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
		Success *bool           `json:"success"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return false, resp.StatusCode, "response bukan JSON valid: " + truncate(string(raw), 180)
	}
	status := strings.ToLower(strings.TrimSpace(decoded.Status))
	okStatus := status == "success" || status == "ok" || (decoded.Success != nil && *decoded.Success)
	if !okStatus {
		msg := decoded.Message
		if msg == "" {
			msg = truncate(string(raw), 180)
		}
		return false, resp.StatusCode, "status=" + decoded.Status + " — " + msg
	}
	if (status == "success" || status == "ok") && len(decoded.Data) == 0 {
		return false, resp.StatusCode, "status success tapi data kosong"
	}
	return true, resp.StatusCode, "ok"
}

func printTable(results []checkResult) {
	fmt.Printf("%-22s %-6s %-4s %-8s %s\n", "AGENT", "STATE", "HTTP", "TIME", "DETAIL")
	fmt.Println(strings.Repeat("-", 100))
	for _, r := range results {
		mark := "FAIL"
		if r.Skipped {
			mark = "SKIP"
		} else if r.OK {
			mark = "OK"
		}
		fmt.Printf("%-22s %-6s %-4d %-8s %s\n", r.Key, mark, r.HTTPCode, r.Elapsed.Round(time.Millisecond), truncate(r.Detail, 60))
	}
}

func printMarkdown(results []checkResult) {
	fmt.Println("| Agent | Status | HTTP | Detail |")
	fmt.Println("| :--- | :--- | :--- | :--- |")
	for _, r := range results {
		status := "🔴 Fail"
		if r.Skipped {
			status = "⚪ Skip"
		} else if r.OK {
			status = "🟢 OK"
		}
		fmt.Printf("| `%s` | %s | %d | %s |\n", r.Key, status, r.HTTPCode, escapeMD(truncate(r.Detail, 120)))
	}
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func escapeMD(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
