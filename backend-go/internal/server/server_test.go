package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"six-sense-web/backend/internal/config"
	"six-sense-web/backend/internal/services"
	"six-sense-web/backend/internal/store"
)

func TestHealthAndListEndpoints(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "web.db")
	st, err := store.Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	detector := services.NewAgentDetector()
	adapter, err := services.NewAgentAdapter(detector, filepath.Join("..", "..", "templates", "insights_prompt_template.md"))
	if err != nil {
		t.Fatalf("agent adapter: %v", err)
	}
	app := New(
		config.Config{AppName: "Six Sense Web", AppVersion: "0.1.0", CORSOrigins: []string{"http://localhost:5173"}},
		st,
		services.NewBrowserSyncService("missing-history", st),
		services.NewPageFetcher(),
		detector,
		adapter,
	)
	handler := app.Handler()

	for _, target := range []string{"/health", "/api/pages", "/api/page-groups", "/api/domains", "/api/blacklist"} {
		request := httptest.NewRequest(http.MethodGet, target, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", target, response.Code, response.Body.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("%s returned invalid json: %v", target, err)
		}
	}
}
