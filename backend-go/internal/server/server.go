package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"six-sense-web/backend/internal/config"
	"six-sense-web/backend/internal/models"
	"six-sense-web/backend/internal/services"
	"six-sense-web/backend/internal/store"
)

type Server struct {
	cfg          config.Config
	store        *store.Store
	syncService  *services.BrowserSyncService
	fetcher      *services.PageFetcher
	detector     *services.AgentDetector
	agentAdapter *services.AgentAdapter
}

func New(cfg config.Config, st *store.Store, syncService *services.BrowserSyncService, fetcher *services.PageFetcher, detector *services.AgentDetector, agentAdapter *services.AgentAdapter) *Server {
	return &Server{cfg: cfg, store: st, syncService: syncService, fetcher: fetcher, detector: detector, agentAdapter: agentAdapter}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.root)
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/api/pages", s.listPages)
	mux.HandleFunc("/api/pages/", s.getPage)
	mux.HandleFunc("/api/page-groups", s.listPageGroups)
	mux.HandleFunc("/api/domains", s.listDomains)
	mux.HandleFunc("/api/sync", s.syncHistory)
	mux.HandleFunc("/api/agents", s.listAgents)
	mux.HandleFunc("/api/analyze", s.analyze)
	mux.HandleFunc("/api/blacklist", s.blacklistRoot)
	mux.HandleFunc("/api/blacklist/", s.blacklistByPath)
	return s.cors(mux)
}

func (s *Server) cors(next http.Handler) http.Handler {
	origins := map[string]bool{}
	for _, origin := range s.cfg.CORSOrigins {
		origins[origin] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "Not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Six Sense Web API", "docs": "/docs", "health": "/health"})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	count, err := s.store.GetPagesCount(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"app":     s.cfg.AppName,
		"version": s.cfg.AppVersion,
		"database": map[string]any{
			"connected":   true,
			"path":        s.store.Path(),
			"pages_count": count,
		},
	})
}

func (s *Server) listPages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	query := r.URL.Query()
	limit := intQuery(query.Get("limit"), 50, 1, 200)
	offset := intQuery(query.Get("offset"), 0, 0, 1_000_000)
	pages, total, err := s.store.ListPages(r.Context(), query.Get("q"), limit, offset, query.Get("sort"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list pages: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.PageListResponse{Pages: pages, Total: total, HasMore: offset+limit < total})
}

func (s *Server) getPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	idText := strings.TrimPrefix(r.URL.Path, "/api/pages/")
	pageID, err := strconv.Atoi(idText)
	if err != nil {
		writeError(w, http.StatusNotFound, "Page not found")
		return
	}
	page, err := s.store.GetPage(r.Context(), pageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get page: "+err.Error())
		return
	}
	if page == nil {
		writeError(w, http.StatusNotFound, "Page not found")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (s *Server) listPageGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	query := r.URL.Query()
	limit := intQuery(query.Get("limit"), 1, 1, 7)
	groups, total, hasMore, nextCursor, err := s.store.ListPageGroups(r.Context(), query.Get("q"), query.Get("domain"), query.Get("cursor"), query.Get("date_from"), query.Get("date_to"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list page groups: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.PageGroupListResponse{Groups: groups, Total: total, HasMore: hasMore, NextCursor: nextCursor})
}

func (s *Server) listDomains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	domains, err := s.store.ListDomains(r.Context(), r.URL.Query().Get("sort"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list domains: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.DomainListResponse{Domains: domains})
}

func (s *Server) syncHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var request models.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if request.Months == 0 {
		request.Months = 2
	}
	response, err := s.syncService.SyncHistory(r.Context(), request.Months)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "Chrome history not found") {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	writeJSON(w, http.StatusOK, models.AgentsResponse{Agents: s.detector.Detect(ctx)})
}

func (s *Server) blacklistRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		entries, err := s.store.ListBlacklistEntries(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, models.BlacklistListResponse{Entries: entries})
	case http.MethodPost:
		var request models.BlacklistCreate
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		entry, err := s.store.CreateBlacklistEntry(r.Context(), request.Pattern, request.Type)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, entry)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) blacklistByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/blacklist/")
	if strings.HasPrefix(path, "pages/") {
		s.blacklistPage(w, r, strings.TrimPrefix(path, "pages/"))
		return
	}
	if r.Method == http.MethodDelete {
		entryID, err := strconv.Atoi(path)
		if err != nil {
			writeError(w, http.StatusNotFound, "Blacklist entry not found")
			return
		}
		deleted, err := s.store.DeleteBlacklistEntry(r.Context(), entryID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !deleted {
			writeError(w, http.StatusNotFound, "Blacklist entry not found")
			return
		}
		hidden, err := s.store.RebuildBlacklist(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, models.BlacklistDeleteResponse{HiddenPages: hidden})
		return
	}
	writeError(w, http.StatusNotFound, "Not found")
}

func (s *Server) blacklistPage(w http.ResponseWriter, r *http.Request, idText string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	pageID, err := strconv.Atoi(idText)
	if err != nil {
		writeError(w, http.StatusNotFound, "Page not found")
		return
	}
	page, err := s.store.GetPage(r.Context(), pageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if page == nil {
		writeError(w, http.StatusNotFound, "Page not found")
		return
	}
	entryType := r.URL.Query().Get("type")
	if entryType == "" {
		entryType = "url"
	}
	pattern := strings.TrimSpace(r.URL.Query().Get("pattern"))
	if pattern == "" {
		pattern = page.URL
	}
	entry, err := s.store.CreateBlacklistEntry(r.Context(), pattern, entryType)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hidden, err := s.store.ApplyBlacklistEntry(r.Context(), entryType, pattern)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, models.BlacklistAddPageResponse{Entry: entry, HiddenPages: hidden})
}

func (s *Server) analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}
	query := r.URL.Query()
	pageID, err := strconv.Atoi(query.Get("page_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "page_id is required")
		return
	}
	agentName := query.Get("agent_name")
	force := query.Get("force") == "true"

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	send := func(event string, payload any) {
		data, _ := json.Marshal(payload)
		fmt.Fprintf(w, "event: %s\n", event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	page, err := s.store.GetPage(r.Context(), pageID)
	if err != nil || page == nil {
		send("analysis_error", map[string]string{"error": "Page not found"})
		return
	}
	if !force {
		existing, err := s.store.GetInsights(r.Context(), pageID)
		if err == nil && existing != nil {
			send("complete", existing)
			return
		}
	}
	send("status", map[string]string{"message": "Fetching page content..."})
	analysisURL, content, ok := s.fetchAnalysisContent(r.Context(), page)
	if !ok {
		send("analysis_error", map[string]string{"error": "Failed to fetch page content"})
		return
	}
	send("status", map[string]string{"message": "Analyzing with " + agentName + "..."})
	fullText, err := s.agentAdapter.AnalyzePage(r.Context(), agentName, analysisURL, content, func(delta string) {
		send("delta", map[string]string{"text": delta})
	})
	if err != nil {
		send("analysis_error", map[string]string{"error": "Analysis failed: " + err.Error()})
		return
	}
	parsed, ok := services.ParseInsights(fullText)
	if !ok {
		send("analysis_error", map[string]string{"error": "Failed to parse insights from AI output"})
		return
	}
	if _, err := s.store.CreateInsights(r.Context(), models.InsightsCreate{
		PageID: pageID, Summary: parsed.Summary, Type: parsed.Type, Keywords: parsed.Keywords,
		UserIntent: parsed.UserIntent, KeyPoints: parsed.KeyPoints, Value: parsed.Value, NextAction: parsed.NextAction, AgentName: agentName,
	}); err != nil {
		send("analysis_error", map[string]string{"error": "Analysis failed: " + err.Error()})
		return
	}
	saved, err := s.store.GetInsights(r.Context(), pageID)
	if err != nil {
		send("analysis_error", map[string]string{"error": "Analysis failed: " + err.Error()})
		return
	}
	send("complete", saved)
}

func (s *Server) fetchAnalysisContent(ctx context.Context, page *models.Page) (string, string, bool) {
	for _, candidate := range analysisContentCandidates(page) {
		content, ok := s.fetcher.FetchContent(ctx, candidate)
		if ok {
			return candidate, content, true
		}
	}
	if isGitHubPage(page) {
		return pageDisplayURL(page), buildGitHubFallbackContent(page), true
	}
	return "", "", false
}

func analysisContentCandidates(page *models.Page) []string {
	if isGitHubPage(page) {
		return githubReadmeCandidates(page)
	}
	candidates := []string{page.URL}
	if page.CanonicalURL != nil && *page.CanonicalURL != "" && *page.CanonicalURL != page.URL {
		candidates = append(candidates, *page.CanonicalURL)
	}
	return candidates
}

func isGitHubPage(page *models.Page) bool {
	return strings.EqualFold(page.Domain, "github.com") || strings.Contains(strings.ToLower(pageDisplayURL(page)), "github.com/")
}

func pageDisplayURL(page *models.Page) string {
	if page.CanonicalURL != nil && *page.CanonicalURL != "" {
		return *page.CanonicalURL
	}
	return page.URL
}

func githubReadmeCandidates(page *models.Page) []string {
	candidates := []string{}
	seen := map[string]bool{}
	add := func(candidate string) {
		if candidate != "" && !seen[candidate] {
			candidates = append(candidates, candidate)
			seen[candidate] = true
		}
	}

	if rawReadme := githubBlobReadmeRawURL(page.URL); rawReadme != "" {
		add(rawReadme)
	}
	owner, repo, ok := githubRepoParts(pageDisplayURL(page))
	if !ok {
		owner, repo, ok = githubRepoParts(page.URL)
	}
	if ok {
		for _, filename := range []string{"README.md", "README.zh.md", "README.rst", "README.txt", "README"} {
			add(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/HEAD/%s", owner, repo, filename))
		}
	}
	add(page.URL)
	if page.CanonicalURL != nil && *page.CanonicalURL != "" && *page.CanonicalURL != page.URL {
		add(*page.CanonicalURL)
	}
	return candidates
}

func githubBlobReadmeRawURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	segments := cleanPathSegments(parsed.Path)
	if len(segments) < 5 || !strings.EqualFold(parsed.Hostname(), "github.com") {
		return ""
	}
	if !strings.EqualFold(segments[2], "blob") || !isReadmeFile(segments[len(segments)-1]) {
		return ""
	}
	return fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/%s",
		segments[0],
		segments[1],
		segments[3],
		strings.Join(segments[4:], "/"),
	)
}

func githubRepoParts(rawURL string) (string, string, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", false
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" && host != "www.github.com" {
		return "", "", false
	}
	segments := cleanPathSegments(parsed.Path)
	if len(segments) < 2 {
		return "", "", false
	}
	owner := segments[0]
	repo := strings.TrimSuffix(segments[1], ".git")
	if owner == "" || repo == "" {
		return "", "", false
	}
	return owner, repo, true
}

func cleanPathSegments(path string) []string {
	rawSegments := strings.Split(path, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	return segments
}

func isReadmeFile(filename string) bool {
	return strings.HasPrefix(strings.ToLower(filename), "readme")
}

func buildGitHubFallbackContent(page *models.Page) string {
	parts := []string{
		"URL: " + pageDisplayURL(page),
		"Original URL: " + page.URL,
		"Title: " + page.Title,
		"Domain: " + page.Domain,
		fmt.Sprintf("Visits today: %d", page.DayCount),
		fmt.Sprintf("Total visits: %d", page.VisitCount),
		"First visit time: " + page.FirstVisitTime,
		"Last visit time: " + page.LastVisitTime,
		"Content note: GitHub page content could not be fetched by the local backend. Analyze this record using the repository URL, page title, and browsing metadata.",
	}
	return strings.Join(parts, "\n")
}

func intQuery(raw string, fallback int, minValue int, maxValue int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	if detail == "" {
		detail = http.StatusText(status)
	}
	writeJSON(w, status, map[string]string{"detail": detail})
}

var _ = errors.New
