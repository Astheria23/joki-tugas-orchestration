package pipeline

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Astheria23/jokiOrchestrator/service/orchestrator/internal/config"
)

func TestResolveScrapeTargets_ExplicitURL(t *testing.T) {
	pr := &PipelineRunner{cfg: &config.Config{WebSearchProvider: "off", WebSearchMaxResults: 3}}
	got, viaSearch, err := pr.resolveScrapeTargets(context.Background(), "Scrape https://www.kompas.com/nasional dulu", "")
	if err != nil {
		t.Fatal(err)
	}
	if viaSearch {
		t.Fatal("explicit URL should not use search")
	}
	if len(got) != 1 || !strings.Contains(got[0], "kompas.com") {
		t.Fatalf("expected kompas url, got %#v", got)
	}
}

func TestResolveScrapeTargets_RejectsWikiSearchPage(t *testing.T) {
	pr := &PipelineRunner{cfg: &config.Config{WebSearchProvider: "off"}}
	got, _, err := pr.resolveScrapeTargets(context.Background(),
		"https://id.wikipedia.org/w/index.php?search=HTML+CSS", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("wiki search SERP must be rejected, got %#v", got)
	}
}

func TestResolveScrapeTargets_AllowsWikiArticle(t *testing.T) {
	pr := &PipelineRunner{cfg: &config.Config{WebSearchProvider: "off"}}
	got, _, err := pr.resolveScrapeTargets(context.Background(),
		"https://id.wikipedia.org/wiki/HTML", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected wiki article, got %#v", got)
	}
}

func TestResolveScrapeTargets_TopicUsesOpenSearch(t *testing.T) {
	pr := &PipelineRunner{
		cfg:    &config.Config{WebSearchProvider: "wikipedia", WebSearchMaxResults: 3},
		client: &http.Client{Timeout: 20 * time.Second},
	}
	got, viaSearch, err := pr.resolveScrapeTargets(context.Background(),
		"Cari materi step by step belajar HTML, CSS, dan JavaScript dasar", "")
	if err != nil {
		t.Fatal(err)
	}
	if !viaSearch {
		t.Fatal("expected viaSearch")
	}
	if len(got) == 0 {
		t.Fatal("expected article urls from opensearch")
	}
	for _, u := range got {
		if !isScrapeableContentURL(u) {
			t.Fatalf("non-scrapeable url returned: %s", u)
		}
		if strings.Contains(u, "index.php") || strings.Contains(u, "search=") {
			t.Fatalf("search page leaked into targets: %s", u)
		}
	}
}

func TestLooksLikeSearchSerp(t *testing.T) {
	if !looksLikeSearchSerp("Hasil Pencarian Wikipedia\nTidak ada konten yang sesuai dengan kriteria yang dicari") {
		t.Fatal("expected serp detection")
	}
	if looksLikeSearchSerp("HTML adalah bahasa markup untuk membuat struktur halaman web. CSS mengatur tampilan.") {
		t.Fatal("article text should not look like serp")
	}
}

func TestNormalizeSearchURL_FiltersSocial(t *testing.T) {
	if normalizeSearchURL("https://twitter.com/foo") != "" {
		t.Fatal("twitter should be filtered")
	}
	if normalizeSearchURL("https://kompas.com/x") == "" {
		t.Fatal("kompas should pass")
	}
}
