package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"six-sense-web/backend/internal/models"
	"six-sense-web/backend/internal/store"
	"six-sense-web/backend/internal/urlutil"
)

type BrowserSyncService struct {
	chromeHistoryPath   string
	chromeBookmarksPath string
	tempHistoryPath     string
	store               *store.Store
}

func NewBrowserSyncService(chromeHistoryPath string, store *store.Store) *BrowserSyncService {
	return &BrowserSyncService{
		chromeHistoryPath:   chromeHistoryPath,
		chromeBookmarksPath: filepath.Join(filepath.Dir(chromeHistoryPath), "Bookmarks"),
		tempHistoryPath:     filepath.Join(os.TempDir(), "History-six-sense-web"),
		store:               store,
	}
}

func (s *BrowserSyncService) SyncHistory(ctx context.Context, months int) (models.SyncResponse, error) {
	if months < 1 || months > 24 {
		return models.SyncResponse{}, errors.New("months must be between 1 and 24")
	}
	start := time.Now()
	rows, err := s.queryBrowserHistory(ctx, months)
	if err != nil {
		return models.SyncResponse{}, err
	}
	bookmarks, err := s.queryBookmarks(ctx)
	if err != nil {
		return models.SyncResponse{}, err
	}
	starredRepos, starChecked := queryGitHubStarredRepos(ctx)
	pages, err := s.buildPages(ctx, rows, bookmarks, starredRepos, starChecked)
	if err != nil {
		return models.SyncResponse{}, err
	}
	newPages := 0
	updatedPages := 0
	for _, page := range pages {
		existing, err := s.store.GetPageByCanonicalKey(ctx, page.CanonicalKey)
		if err != nil {
			return models.SyncResponse{}, err
		}
		if existing != nil {
			title := page.Title
			dayCount := page.DayCount
			rawURL := page.URL
			canonicalURL := page.CanonicalURL
			canonicalKey := page.CanonicalKey
			lastVisit := page.LastVisitTime
			firstVisit := page.FirstVisitTime
			isBookmarked := page.IsBookmarked
			bookmarkTitle := bookmarkStringValue(page.BookmarkTitle)
			bookmarkFolder := bookmarkStringValue(page.BookmarkFolder)
			bookmarkAddedAt := bookmarkStringValue(page.BookmarkAddedAt)
			update := models.PageUpdate{
				URL: &rawURL, CanonicalURL: &canonicalURL, CanonicalKey: &canonicalKey, Title: &title,
				DayCount: &dayCount, IsBookmarked: &isBookmarked, BookmarkTitle: &bookmarkTitle, BookmarkFolder: &bookmarkFolder, BookmarkAddedAt: &bookmarkAddedAt,
				LastVisitTime: &lastVisit, FirstVisitTime: &firstVisit,
			}
			if starChecked {
				isGitHubStarred := page.IsGitHubStarred
				update.IsGitHubStarred = &isGitHubStarred
			}
			if err := s.store.UpdatePage(ctx, existing.ID, update); err != nil {
				return models.SyncResponse{}, err
			}
			updatedPages++
		} else {
			if _, err := s.store.CreatePage(ctx, page); err != nil {
				return models.SyncResponse{}, err
			}
			newPages++
		}
	}
	totalPages, err := s.store.GetPagesCount(ctx)
	if err != nil {
		return models.SyncResponse{}, err
	}
	return models.SyncResponse{Status: "success", NewPages: newPages, UpdatedPages: updatedPages, TotalPages: totalPages, SyncTime: start.Format("2006-01-02T15:04:05.999999")}, nil
}

type historyRow struct {
	URL            string
	Title          string
	VisitCount     int
	FirstVisitTime string
	LastVisitTime  string
}

type bookmarkInfo struct {
	URL     string
	Title   string
	Folder  string
	AddedAt string
}

type chromeBookmarkFile struct {
	Roots map[string]chromeBookmarkNode `json:"roots"`
}

type chromeBookmarkNode struct {
	Type      string               `json:"type"`
	Name      string               `json:"name"`
	URL       string               `json:"url"`
	DateAdded string               `json:"date_added"`
	Children  []chromeBookmarkNode `json:"children"`
}

func (s *BrowserSyncService) queryBrowserHistory(ctx context.Context, months int) ([]historyRow, error) {
	if _, err := os.Stat(s.chromeHistoryPath); err != nil {
		return nil, fmt.Errorf("Chrome history not found at %s", s.chromeHistoryPath)
	}
	input, err := os.ReadFile(s.chromeHistoryPath)
	if err != nil {
		return nil, fmt.Errorf("Error copying Chrome history: %w", err)
	}
	if err := os.WriteFile(s.tempHistoryPath, input, 0o600); err != nil {
		return nil, fmt.Errorf("Error copying Chrome history: %w", err)
	}
	defer os.Remove(s.tempHistoryPath)

	db, err := sql.Open("sqlite3", s.tempHistoryPath+"?_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf(`
		SELECT u.url,
		       COALESCE(u.title, ''),
		       COUNT(*) as visit_count,
		       datetime(MIN(v.visit_time)/1000000-11644473600, 'unixepoch', 'localtime') as first_visit_time,
		       datetime(MAX(v.visit_time)/1000000-11644473600, 'unixepoch', 'localtime') as last_visit_time
		FROM visits v
		JOIN urls u ON u.id = v.url
		WHERE v.visit_time > (strftime('%%s', 'now', '-%d months') + 11644473600) * 1000000
		GROUP BY u.url,
		         date(v.visit_time/1000000-11644473600, 'unixepoch', 'localtime')
		ORDER BY MAX(v.visit_time) DESC
	`, months)
	sqlRows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()
	rows := []historyRow{}
	for sqlRows.Next() {
		var row historyRow
		if err := sqlRows.Scan(&row.URL, &row.Title, &row.VisitCount, &row.FirstVisitTime, &row.LastVisitTime); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, sqlRows.Err()
}

func (s *BrowserSyncService) queryBookmarks(ctx context.Context) (map[string]bookmarkInfo, error) {
	if _, err := os.Stat(s.chromeBookmarksPath); err != nil {
		return map[string]bookmarkInfo{}, nil
	}
	input, err := os.ReadFile(s.chromeBookmarksPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading Chrome bookmarks: %w", err)
	}
	var file chromeBookmarkFile
	if err := json.Unmarshal(input, &file); err != nil {
		return nil, fmt.Errorf("Error parsing Chrome bookmarks: %w", err)
	}
	bookmarks := map[string]bookmarkInfo{}
	for rootName, root := range file.Roots {
		if rootName == "sync_transaction_version" {
			continue
		}
		collectBookmarkURLs(root, nil, bookmarks)
	}
	return bookmarks, ctx.Err()
}

func collectBookmarkURLs(node chromeBookmarkNode, folderPath []string, bookmarks map[string]bookmarkInfo) {
	if node.Type == "url" && strings.TrimSpace(node.URL) != "" {
		info := bookmarkInfo{
			URL:     node.URL,
			Title:   node.Name,
			Folder:  bookmarkFolder(folderPath),
			AddedAt: chromeBookmarkTime(node.DateAdded),
		}
		addBookmarkKey(bookmarks, node.URL, info)
		canonicalURL, _ := urlutil.NormalizePageURL(node.URL, "")
		addBookmarkKey(bookmarks, canonicalURL, info)
		return
	}
	nextPath := folderPath
	if node.Type == "folder" && node.Name != "" {
		nextPath = append(folderPath, node.Name)
	}
	for _, child := range node.Children {
		collectBookmarkURLs(child, nextPath, bookmarks)
	}
}

func addBookmarkKey(bookmarks map[string]bookmarkInfo, key string, info bookmarkInfo) {
	if key == "" {
		return
	}
	if _, exists := bookmarks[key]; !exists {
		bookmarks[key] = info
	}
}

func bookmarkFolder(path []string) string {
	parts := []string{}
	for _, part := range path {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, " / ")
}

func chromeBookmarkTime(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	microseconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return ""
	}
	seconds := microseconds/1000000 - 11644473600
	return time.Unix(seconds, 0).Local().Format("2006-01-02T15:04:05")
}

func queryGitHubStarredRepos(ctx context.Context) (map[string]bool, bool) {
	command := findGHCommand()
	if command == "" {
		return map[string]bool{}, false
	}
	apiCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(apiCtx, command, "api", "--paginate", "/user/starred", "--jq", ".[].full_name")
	output, err := cmd.Output()
	if err != nil {
		return map[string]bool{}, false
	}
	repos := map[string]bool{}
	for _, line := range strings.Split(string(output), "\n") {
		repo := strings.ToLower(strings.TrimSpace(line))
		if repo != "" {
			repos[repo] = true
		}
	}
	return repos, true
}

func findGHCommand() string {
	candidates := []string{}
	if configured := strings.TrimSpace(os.Getenv("GH_PATH")); configured != "" {
		candidates = append(candidates, configured)
	}
	if path, err := exec.LookPath("gh"); err == nil {
		candidates = append(candidates, path)
	}
	candidates = append(candidates, "/opt/homebrew/bin/gh", "/usr/local/bin/gh")
	for _, candidate := range candidates {
		if isExecutableFile(candidate) {
			return candidate
		}
	}
	return ""
}

func githubRepoKeyFromCanonicalKey(canonicalKey string) string {
	if !strings.HasPrefix(canonicalKey, "github-repo:") {
		return ""
	}
	rest := strings.TrimPrefix(canonicalKey, "github-repo:")
	lastColon := strings.LastIndex(rest, ":")
	if lastColon <= 0 {
		return ""
	}
	return strings.ToLower(rest[:lastColon])
}

func (s *BrowserSyncService) buildPages(ctx context.Context, rows []historyRow, bookmarks map[string]bookmarkInfo, starredRepos map[string]bool, starChecked bool) ([]models.PageCreate, error) {
	pagesByKey := map[string]models.PageCreate{}
	for _, row := range rows {
		if shouldFilterURL(row.URL) {
			continue
		}
		blacklisted, err := s.store.IsURLBlacklisted(ctx, row.URL)
		if err != nil {
			return nil, err
		}
		if blacklisted {
			continue
		}
		canonicalURL, canonicalKey := urlutil.NormalizePageURL(row.URL, row.LastVisitTime)
		page := models.PageCreate{
			URL: row.URL, CanonicalURL: canonicalURL, CanonicalKey: canonicalKey,
			Title: fallbackTitle(row.Title, row.URL), Domain: extractDomain(row.URL),
			DayCount: row.VisitCount, VisitCount: row.VisitCount,
			LastVisitTime: row.LastVisitTime, FirstVisitTime: row.FirstVisitTime,
		}
		applyBookmarkInfo(&page, bookmarks)
		if starChecked {
			page.IsGitHubStarred = starredRepos[githubRepoKeyFromCanonicalKey(canonicalKey)]
		}
		existing, ok := pagesByKey[canonicalKey]
		if !ok {
			pagesByKey[canonicalKey] = page
			continue
		}
		existing.DayCount += row.VisitCount
		existing.VisitCount += row.VisitCount
		if page.IsBookmarked {
			existing.IsBookmarked = true
			existing.BookmarkTitle = page.BookmarkTitle
			existing.BookmarkFolder = page.BookmarkFolder
			existing.BookmarkAddedAt = page.BookmarkAddedAt
		}
		if page.IsGitHubStarred {
			existing.IsGitHubStarred = true
		}
		if page.FirstVisitTime < existing.FirstVisitTime || existing.FirstVisitTime == "" {
			existing.FirstVisitTime = page.FirstVisitTime
		}
		if shouldReplaceTitle(existing.Title, page.Title) {
			existing.Title = page.Title
		}
		if page.LastVisitTime >= existing.LastVisitTime {
			existing.URL = page.URL
			existing.CanonicalURL = page.CanonicalURL
			existing.Domain = page.Domain
			existing.LastVisitTime = page.LastVisitTime
		}
		pagesByKey[canonicalKey] = existing
	}
	pages := make([]models.PageCreate, 0, len(pagesByKey))
	for _, page := range pagesByKey {
		pages = append(pages, page)
	}
	return pages, nil
}

func applyBookmarkInfo(page *models.PageCreate, bookmarks map[string]bookmarkInfo) {
	info, ok := bookmarks[page.URL]
	if !ok {
		info, ok = bookmarks[page.CanonicalURL]
	}
	if !ok {
		return
	}
	page.IsBookmarked = true
	page.BookmarkTitle = bookmarkStringPointer(info.Title)
	page.BookmarkFolder = bookmarkStringPointer(info.Folder)
	page.BookmarkAddedAt = bookmarkStringPointer(info.AddedAt)
}

func bookmarkStringPointer(value string) *string {
	return &value
}

func bookmarkStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func fallbackTitle(title string, rawURL string) string {
	if strings.TrimSpace(title) == "" {
		return rawURL
	}
	return title
}

func shouldReplaceTitle(current string, next string) bool {
	if next == "" {
		return false
	}
	if current == "" {
		return true
	}
	generic := map[string]bool{"微信公众平台": true, "微信公众平台安全验证": true}
	if generic[next] && !generic[current] {
		return false
	}
	if generic[current] && !generic[next] {
		return true
	}
	return true
}

func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	if parsed.Host != "" {
		return parsed.Host
	}
	parts := strings.Split(parsed.Path, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func shouldFilterURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if parsed.Scheme == "file" {
		return true
	}
	host := strings.Trim(parsed.Hostname(), "[]")
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return true
	}
	return regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}$`).MatchString(host)
}
