package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchContentRejectsEmptyReadableBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Only title</title></head><body><script>window.app = {}</script></body></html>`))
	}))
	defer server.Close()

	fetcher := NewPageFetcher()
	content, ok := fetcher.FetchContent(context.Background(), server.URL)
	if ok {
		t.Fatalf("FetchContent ok = true, content = %q", content)
	}
	if content != "" {
		t.Fatalf("FetchContent content = %q, want empty", content)
	}
}

func TestFetchContentAcceptsReadableBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Readable</title></head><body><main>This page has enough useful readable text for analysis.</main></body></html>`))
	}))
	defer server.Close()

	fetcher := NewPageFetcher()
	content, ok := fetcher.FetchContent(context.Background(), server.URL)
	if !ok {
		t.Fatal("FetchContent ok = false, want true")
	}
	if content == "" {
		t.Fatal("FetchContent returned empty content")
	}
}
