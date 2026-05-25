package urlutil

import "testing"

func TestNormalizePageURLGroupsGitHubRepositoryPages(t *testing.T) {
	lastVisit := "2026-05-24 12:00:00"
	urls := []string{
		"https://github.com/multica-ai/andrej-karpathy-skills/blob/main/CLAUDE.md",
		"https://github.com/multica-ai/andrej-karpathy-skills/blob/main/README.zh.md",
		"https://github.com/multica-ai/andrej-karpathy-skills",
	}

	var firstKey string
	for _, rawURL := range urls {
		canonicalURL, canonicalKey := NormalizePageURL(rawURL, lastVisit)
		if canonicalURL != "https://github.com/multica-ai/andrej-karpathy-skills" {
			t.Fatalf("canonical URL = %q, want repo root", canonicalURL)
		}
		if firstKey == "" {
			firstKey = canonicalKey
			continue
		}
		if canonicalKey != firstKey {
			t.Fatalf("canonical key = %q, want %q", canonicalKey, firstKey)
		}
	}
	if firstKey != "github-repo:multica-ai/andrej-karpathy-skills:2026-05-24" {
		t.Fatalf("canonical key = %q, want github repo date key", firstKey)
	}
}

func TestNormalizePageURLKeepsGitHubReservedSitePagesGeneric(t *testing.T) {
	canonicalURL, canonicalKey := NormalizePageURL("https://github.com/features/actions", "2026-05-24 12:00:00")

	if canonicalURL != "https://github.com/features/actions" {
		t.Fatalf("canonical URL = %q, want generic URL", canonicalURL)
	}
	if canonicalKey != "url:https://github.com/features/actions:2026-05-24" {
		t.Fatalf("canonical key = %q, want generic key", canonicalKey)
	}
}
