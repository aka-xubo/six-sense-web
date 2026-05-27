package server

import (
	"context"
	"reflect"
	"testing"

	"six-sense-web/backend/internal/models"
	"six-sense-web/backend/internal/services"
)

func TestGitHubReadmeCandidatesPreferReadmeOverRepositoryHTML(t *testing.T) {
	canonicalURL := "https://github.com/multica-ai/andrej-karpathy-skills"
	page := &models.Page{
		URL:          "https://github.com/multica-ai/andrej-karpathy-skills/blob/main/CLAUDE.md",
		CanonicalURL: &canonicalURL,
		Domain:       "github.com",
	}

	got := analysisContentCandidates(page)
	wantPrefix := []string{
		"https://raw.githubusercontent.com/multica-ai/andrej-karpathy-skills/HEAD/README.md",
		"https://raw.githubusercontent.com/multica-ai/andrej-karpathy-skills/HEAD/README.zh.md",
	}

	if !reflect.DeepEqual(got[:2], wantPrefix) {
		t.Fatalf("candidate prefix = %#v, want %#v", got[:2], wantPrefix)
	}
}

func TestGitHubReadmeCandidatesPreferVisitedReadmeVariant(t *testing.T) {
	canonicalURL := "https://github.com/multica-ai/andrej-karpathy-skills"
	page := &models.Page{
		URL:          "https://github.com/multica-ai/andrej-karpathy-skills/blob/main/README.zh.md",
		CanonicalURL: &canonicalURL,
		Domain:       "github.com",
	}

	got := analysisContentCandidates(page)
	want := "https://raw.githubusercontent.com/multica-ai/andrej-karpathy-skills/main/README.zh.md"

	if got[0] != want {
		t.Fatalf("first candidate = %q, want %q", got[0], want)
	}
}

func TestFetchAnalysisContentDoesNotFallbackToGitHubMetadata(t *testing.T) {
	canonicalURL := "https://github.com/multica-ai/andrej-karpathy-skills"
	page := &models.Page{
		URL:          "https://github.com/multica-ai/andrej-karpathy-skills",
		CanonicalURL: &canonicalURL,
		Title:        "andrej-karpathy-skills",
		Domain:       "github.com",
	}
	app := &Server{fetcher: services.NewPageFetcher()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	analysisURL, content, ok := app.fetchAnalysisContent(ctx, page)
	if ok {
		t.Fatalf("fetchAnalysisContent ok = true, url = %q, content = %q", analysisURL, content)
	}
	if analysisURL != "" || content != "" {
		t.Fatalf("fetchAnalysisContent returned url = %q, content = %q; want empty values", analysisURL, content)
	}
}
