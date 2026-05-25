package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"six-sense-web/backend/internal/config"
	"six-sense-web/backend/internal/server"
	"six-sense-web/backend/internal/services"
	"six-sense-web/backend/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	st, err := store.Open(ctx, cfg.DBPath)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	defer st.Close()

	detector := services.NewAgentDetector()
	adapter, err := services.NewAgentAdapter(detector, filepath.Join("..", "backend", "templates", "insights_prompt_template.md"))
	if err != nil {
		adapter, err = services.NewAgentAdapter(detector, filepath.Join("templates", "insights_prompt_template.md"))
		if err != nil {
			log.Fatalf("agent adapter init failed: %v", err)
		}
	}

	app := server.New(
		cfg,
		st,
		services.NewBrowserSyncService(cfg.ChromeHistoryPath, st),
		services.NewPageFetcher(),
		detector,
		adapter,
	)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("Six Sense Web Go API listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, app.Handler()))
}
