package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Astheria23/jokiOrchestrator/shared/logging"
)

var (
	ddgResultHref = regexp.MustCompile(`(?i)href="(https?://[^"]+)"`)
	ddgUddg       = regexp.MustCompile(`uddg=([^&"]+)`)
)

const searchUserAgent = "BananacademicBot/1.0 (joki-tugas orchestrator; academic research assistant)"

// webSearch finds candidate page URLs for a topic.
// OpenCode Go / DeepSeek do NOT provide web search — this runs in the orchestrator.
//
// Providers:
//   - brave: Brave Search API (best quality; needs BRAVE_SEARCH_API_KEY)
//   - wikipedia: MediaWiki OpenSearch → real article URLs (default free)
//   - duckduckgo: HTML scrape (often blocked/SSL flaky)
//   - off: no search
func (pr *PipelineRunner) webSearch(ctx context.Context, topic string, limit int) ([]string, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, fmt.Errorf("empty search topic")
	}
	if limit <= 0 {
		limit = 3
	}
	if limit > 8 {
		limit = 8
	}

	provider := strings.ToLower(strings.TrimSpace(pr.cfg.WebSearchProvider))
	if provider == "" {
		if strings.TrimSpace(pr.cfg.BraveSearchKey) != "" {
			provider = "brave"
		} else {
			provider = "wikipedia"
		}
	}

	logging.Log.Info().Str("provider", provider).Str("topic", topic).Int("limit", limit).Msg("Web search start")

	switch provider {
	case "brave":
		return pr.webSearchBrave(ctx, topic, limit)
	case "wikipedia", "wiki":
		return pr.webSearchWikipedia(ctx, topic, limit)
	case "duckduckgo", "ddg":
		urls, err := pr.webSearchDuckDuckGo(ctx, topic, limit)
		if err != nil || len(urls) == 0 {
			logging.Log.Warn().Err(err).Msg("DuckDuckGo search empty/failed — trying Wikipedia")
			return pr.webSearchWikipedia(ctx, topic, limit)
		}
		return urls, nil
	case "off", "none", "disabled":
		return nil, nil
	default:
		return pr.webSearchWikipedia(ctx, topic, limit)
	}
}

func (pr *PipelineRunner) webSearchBrave(ctx context.Context, topic string, limit int) ([]string, error) {
	key := strings.TrimSpace(pr.cfg.BraveSearchKey)
	if key == "" {
		return nil, fmt.Errorf("BRAVE_SEARCH_API_KEY kosong")
	}
	endpoint := "https://api.search.brave.com/res/v1/web/search?q=" + url.QueryEscape(topic) + "&count=" + fmt.Sprintf("%d", limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", key)
	req.Header.Set("User-Agent", searchUserAgent)

	resp, err := pr.searchHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brave search HTTP %d: %s", resp.StatusCode, truncateRunes(string(body), 180))
	}

	var parsed struct {
		Web struct {
			Results []struct {
				URL string `json:"url"`
			} `json:"results"`
		} `json:"web"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return collectUniqueURLs(func(yield func(string)) {
		for _, r := range parsed.Web.Results {
			yield(r.URL)
		}
	}, limit), nil
}

// webSearchWikipedia uses MediaWiki OpenSearch (id then en) — free, stable, returns scrapeable article URLs.
func (pr *PipelineRunner) webSearchWikipedia(ctx context.Context, topic string, limit int) ([]string, error) {
	var all []string
	for _, lang := range []string{"id", "en"} {
		part, err := pr.wikipediaOpenSearch(ctx, lang, topic, limit)
		if err != nil {
			logging.Log.Warn().Err(err).Str("lang", lang).Msg("Wikipedia OpenSearch failed")
			continue
		}
		all = append(all, part...)
		if len(all) >= limit {
			break
		}
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("wikipedia opensearch returned no urls")
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (pr *PipelineRunner) wikipediaOpenSearch(ctx context.Context, lang, topic string, limit int) ([]string, error) {
	endpoint := fmt.Sprintf(
		"https://%s.wikipedia.org/w/api.php?action=opensearch&search=%s&limit=%d&namespace=0&format=json",
		lang,
		url.QueryEscape(topic),
		limit,
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", searchUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := pr.searchHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikipedia HTTP %d: %s", resp.StatusCode, truncateRunes(string(body), 120))
	}

	// OpenSearch JSON: [query, [titles...], [descs...], [urls...]]
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil || len(raw) < 4 {
		return nil, fmt.Errorf("wikipedia opensearch decode failed")
	}
	var urls []string
	if err := json.Unmarshal(raw[3], &urls); err != nil {
		return nil, err
	}
	return collectUniqueURLs(func(yield func(string)) {
		for _, u := range urls {
			yield(u)
		}
	}, limit), nil
}

func (pr *PipelineRunner) webSearchDuckDuckGo(ctx context.Context, topic string, limit int) ([]string, error) {
	form := url.Values{}
	form.Set("q", topic)
	form.Set("kl", "id-id")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://html.duckduckgo.com/html/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", searchUserAgent)

	resp, err := pr.searchHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo HTTP %d", resp.StatusCode)
	}

	html := string(body)
	out := make([]string, 0, limit)
	seen := map[string]struct{}{}
	add := func(raw string) {
		u := normalizeSearchURL(raw)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}

	for _, m := range ddgUddg.FindAllStringSubmatch(html, -1) {
		decoded, err := url.QueryUnescape(m[1])
		if err != nil {
			continue
		}
		add(decoded)
		if len(out) >= limit {
			return out, nil
		}
	}
	for _, m := range ddgResultHref.FindAllStringSubmatch(html, -1) {
		add(m[1])
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func collectUniqueURLs(iter func(func(string)), limit int) []string {
	out := make([]string, 0, limit)
	seen := map[string]struct{}{}
	iter(func(raw string) {
		if len(out) >= limit {
			return
		}
		u := normalizeSearchURL(raw)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	})
	return out
}

func (pr *PipelineRunner) searchHTTPClient() *http.Client {
	timeout := 20 * time.Second
	if pr.client != nil && pr.client.Timeout > 0 && pr.client.Timeout < timeout {
		timeout = pr.client.Timeout
	}
	return &http.Client{Timeout: timeout}
}

func normalizeSearchURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, ".,;)]}>\"'")
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	skipHosts := []string{
		"duckduckgo.com", "google.com", "google.co.id", "youtube.com", "youtu.be",
		"facebook.com", "instagram.com", "twitter.com", "x.com", "tiktok.com",
		"bing.com", "microsoft.com",
	}
	for _, s := range skipHosts {
		if host == s || strings.HasSuffix(host, "."+s) {
			return ""
		}
	}
	return u.String()
}
