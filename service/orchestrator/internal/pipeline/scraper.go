package pipeline

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Astheria23/jokiOrchestrator/shared/logging"
)

var (
	multiURLRegex = regexp.MustCompile(`https?://[^\s<>"')\]]+`)
	templateLine  = regexp.MustCompile(`(?i)^/template\s+\S+\s*`)
)

// scrapeOrchestrate prepares URL-only payloads for web_scraper and can fan-out
// across multiple URLs, merging text results for the next pipeline step.
//
// Flow when user gives a topic (no URL):
//  1. web search (Wikipedia OpenSearch / Brave) → article URLs only
//  2. scrape each candidate via web_scraper agent
//  3. skip SERP / empty pages and try the next URL
//
// OpenCode Go / DeepSeek does NOT provide web search — search lives in the orchestrator.
func (pr *PipelineRunner) scrapeOrchestrate(
	ctx context.Context,
	taskID string,
	agentEndpoint string,
	prompt string,
	currentText string,
) (*AgentResponse, error) {
	targets, searchUsed, err := pr.resolveScrapeTargets(ctx, prompt, currentText)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf(
			"web_scraper butuh URL artikel http(s) yang valid. Tambahin link sumber, atau perjelas topiknya",
		)
	}

	logging.Log.Info().
		Str("taskId", taskID).
		Int("urls", len(targets)).
		Bool("viaSearch", searchUsed).
		Strs("targets", targets).
		Msg("Scraper orchestration: resolved target URL(s)")

	var chunks []string
	var last *AgentResponse
	var lastErr error

	for i, target := range targets {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !isScrapeableContentURL(target) {
			logging.Log.Warn().Str("url", target).Msg("Scraper orchestration: skip non-article / SERP url")
			continue
		}

		logging.Log.Info().
			Str("taskId", taskID).
			Str("url", target).
			Msgf("Scraper orchestration: fetching %d/%d", i+1, len(targets))

		resp, err := pr.executeAgent(ctx, taskID, "web_scraper", agentEndpoint, target, target)
		if err != nil {
			lastErr = err
			logging.Log.Warn().Err(err).Str("url", target).Msg("Scraper orchestration: scrape failed, try next")
			continue
		}

		text := ""
		if resp.Data != nil {
			text = strings.TrimSpace(resp.Data.Result)
			if text == "" && resp.Data.FileURL != "" {
				text = resp.Data.FileURL
			}
		}
		if looksLikeSearchSerp(text) {
			lastErr = fmt.Errorf("konten dari %s terlihat seperti halaman pencarian, bukan artikel", target)
			logging.Log.Warn().Str("url", target).Msg("Scraper orchestration: SERP-like content rejected")
			continue
		}
		if len([]rune(text)) < 80 {
			lastErr = fmt.Errorf("konten dari %s terlalu pendek", target)
			logging.Log.Warn().Str("url", target).Int("len", len([]rune(text))).Msg("Scraper orchestration: content too short")
			continue
		}

		last = resp
		if len(targets) > 1 {
			chunks = append(chunks, fmt.Sprintf("=== Sumber: %s ===\n%s", target, text))
		} else {
			chunks = append(chunks, text)
		}
	}

	if len(chunks) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("web_scraper gagal dapat konten artikel: %w", lastErr)
		}
		return nil, fmt.Errorf("web_scraper tidak menghasilkan konten artikel yang layak dari URL target")
	}

	merged := strings.Join(chunks, "\n\n")
	if last == nil {
		last = &AgentResponse{
			Status: "success",
			TaskID: taskID,
			Data: &struct {
				Result  string `json:"result"`
				FileURL string `json:"file_url"`
			}{},
		}
	}
	if last.Data == nil {
		last.Data = &struct {
			Result  string `json:"result"`
			FileURL string `json:"file_url"`
		}{}
	}
	last.Status = "success"
	last.Data.Result = merged
	last.Data.FileURL = ""
	return last, nil
}

// resolveScrapeTargets picks concrete http(s) article URLs for the scraper.
// Priority: explicit URLs → web search (article URLs) → simplified Wikipedia OpenSearch retries.
// Never returns Wikipedia /search index pages as scrape targets.
func (pr *PipelineRunner) resolveScrapeTargets(ctx context.Context, prompt, currentText string) ([]string, bool, error) {
	seen := map[string]struct{}{}
	var out []string
	add := func(u string) {
		u = normalizeSearchURL(u)
		if u == "" || !isScrapeableContentURL(u) {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}

	for _, m := range multiURLRegex.FindAllString(prompt, -1) {
		add(m)
	}
	for _, m := range multiURLRegex.FindAllString(currentText, -1) {
		add(m)
	}
	if len(out) > 0 {
		return out, false, nil
	}

	topic := extractScrapeTopic(prompt, currentText)
	if topic == "" {
		return nil, false, nil
	}

	limit := 3
	if pr.cfg != nil && pr.cfg.WebSearchMaxResults > 0 {
		limit = pr.cfg.WebSearchMaxResults
	}

	queries := searchQueryVariants(topic)
	for _, q := range queries {
		found, err := pr.webSearch(ctx, q, limit)
		if err != nil {
			logging.Log.Warn().Err(err).Str("query", q).Msg("Web search failed for query variant")
			continue
		}
		for _, u := range found {
			add(u)
		}
		// Keep searching variants to gather multiple article sources (HTML + CSS + JS, etc.)
		if len(out) >= limit {
			break
		}
	}

	if len(out) > 0 {
		if len(out) > limit {
			out = out[:limit]
		}
		return out, true, nil
	}

	logging.Log.Warn().Str("topic", topic).Msg("Web search returned no scrapeable article URLs")
	return nil, false, nil
}

func searchQueryVariants(topic string) []string {
	topic = strings.TrimSpace(topic)
	variants := []string{topic}

	// Prefer compact tech keywords when present (better for Wikipedia OpenSearch).
	lower := strings.ToLower(topic)
	var keys []string
	for _, k := range []string{"html", "css", "javascript", "js", "python", "golang", "react", "vue", "sql"} {
		if strings.Contains(lower, k) {
			label := k
			switch k {
			case "js":
				label = "JavaScript"
			case "html":
				label = "HTML"
			case "css":
				label = "CSS"
			case "sql":
				label = "SQL"
			default:
				label = strings.ToUpper(k[:1]) + k[1:]
			}
			keys = append(keys, label)
		}
	}
	// Single-keyword queries work best on Wikipedia OpenSearch (multi-word often returns []).
	for _, k := range keys {
		variants = append(variants, k)
	}
	if len(keys) > 1 {
		variants = append(variants, strings.Join(keys, " "))
	}

	// Shorten long prompts
	runes := []rune(topic)
	if len(runes) > 60 {
		variants = append(variants, string(runes[:60]))
	}

	seen := map[string]struct{}{}
	var out []string
	for _, v := range variants {
		v = strings.Join(strings.Fields(v), " ")
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// isScrapeableContentURL rejects search SERPs / index pages that produce garbage PPT content.
func isScrapeableContentURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	path := strings.ToLower(u.Path)
	q := u.Query()

	if strings.Contains(host, "wikipedia.org") {
		// Only allow article pages: /wiki/Title (not /w/index.php?search=...)
		if strings.Contains(path, "/w/index.php") || q.Get("search") != "" {
			return false
		}
		if !strings.Contains(path, "/wiki/") {
			return false
		}
		// Special non-article namespaces
		bad := []string{"/wiki/special:", "/wiki/file:", "/wiki/berkas:", "/wiki/template:", "/wiki/bantuan:"}
		for _, b := range bad {
			if strings.Contains(path, b) {
				return false
			}
		}
	}

	if strings.Contains(path, "/search") && q.Get("q") != "" {
		return false
	}
	return true
}

func looksLikeSearchSerp(text string) bool {
	lower := strings.ToLower(text)
	signals := 0
	for _, s := range []string{
		"hasil pencarian",
		"tidak ada konten yang sesuai",
		"halaman pencarian",
		"search results",
		"did you mean",
		"dari 43",
		"20 butir",
		"tuju ke",
	} {
		if strings.Contains(lower, s) {
			signals++
		}
	}
	return signals >= 2 || strings.Contains(lower, "tidak ada konten yang sesuai dengan kriteria")
}

func extractScrapeTopic(prompt, currentText string) string {
	candidates := []string{prompt, currentText}
	for _, raw := range candidates {
		lines := strings.Split(raw, "\n")
		var kept []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if templateLine.MatchString(line) {
				continue
			}
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				inner := strings.TrimSpace(line[1 : len(line)-1])
				if strings.Contains(strings.ToLower(inner), "paste") ||
					strings.Contains(strings.ToLower(inner), "jelaskan") {
					continue
				}
				line = inner
			}
			kept = append(kept, line)
		}
		topic := strings.TrimSpace(strings.Join(kept, " "))
		topic = multiURLRegex.ReplaceAllString(topic, " ")
		topic = strings.Join(strings.Fields(topic), " ")
		if len([]rune(topic)) >= 3 {
			runes := []rune(topic)
			if len(runes) > 160 {
				topic = string(runes[:160])
			}
			return topic
		}
	}
	return ""
}
