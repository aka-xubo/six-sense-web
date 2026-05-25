package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAgentDetectorUsesConfiguredPathAndCaches(t *testing.T) {
	dir := t.TempDir()
	codexPath := filepath.Join(dir, "codex")
	if err := os.WriteFile(codexPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write codex fixture: %v", err)
	}
	t.Setenv("CODEX_PATH", codexPath)

	detector := NewAgentDetector()
	agents := detector.Detect(context.Background())
	if len(agents) == 0 {
		t.Fatal("expected agents")
	}

	resolved, err := detector.Resolve(context.Background(), "codex")
	if err != nil {
		t.Fatalf("resolve codex: %v", err)
	}
	if resolved != codexPath {
		t.Fatalf("resolved path = %q, want %q", resolved, codexPath)
	}

	t.Setenv("CODEX_PATH", "")
	resolvedAgain, err := detector.Resolve(context.Background(), "codex")
	if err != nil {
		t.Fatalf("resolve cached codex: %v", err)
	}
	if resolvedAgain != codexPath {
		t.Fatalf("cached path = %q, want %q", resolvedAgain, codexPath)
	}
}
