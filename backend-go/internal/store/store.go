package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"six-sense-web/backend/internal/models"
	"six-sense-web/backend/internal/urlutil"
)

type Store struct {
	db     *sql.DB
	dbPath string
}

func Open(ctx context.Context, dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db, dbPath: dbPath}
	if err := store.initSchema(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Path() string {
	return s.dbPath
}

func (s *Store) initSchema(ctx context.Context) error {
	schema, err := readSchema()
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, string(schema)); err != nil {
		return err
	}
	return s.ensureMigrations(ctx)
}

func readSchema() ([]byte, error) {
	candidates := []string{
		"schema.sql",
		filepath.Join("backend-go", "schema.sql"),
		filepath.Join("..", "schema.sql"),
		filepath.Join("..", "..", "schema.sql"),
		filepath.Join("..", "..", "..", "schema.sql"),
	}
	for _, candidate := range candidates {
		schema, err := os.ReadFile(candidate)
		if err == nil {
			return schema, nil
		}
	}
	return nil, os.ErrNotExist
}

func (s *Store) ensureMigrations(ctx context.Context) error {
	columns, err := s.tableColumns(ctx, "pages")
	if err != nil {
		return err
	}
	migrations := map[string]string{
		"is_blacklisted":    "ALTER TABLE pages ADD COLUMN is_blacklisted INTEGER DEFAULT 0",
		"canonical_url":     "ALTER TABLE pages ADD COLUMN canonical_url TEXT",
		"canonical_key":     "ALTER TABLE pages ADD COLUMN canonical_key TEXT",
		"day_count":         "ALTER TABLE pages ADD COLUMN day_count INTEGER DEFAULT 0",
		"is_bookmarked":     "ALTER TABLE pages ADD COLUMN is_bookmarked INTEGER DEFAULT 0",
		"bookmark_title":    "ALTER TABLE pages ADD COLUMN bookmark_title TEXT",
		"bookmark_folder":   "ALTER TABLE pages ADD COLUMN bookmark_folder TEXT",
		"bookmark_added_at": "ALTER TABLE pages ADD COLUMN bookmark_added_at TEXT",
		"is_github_starred": "ALTER TABLE pages ADD COLUMN is_github_starred INTEGER DEFAULT 0",
	}
	for column, statement := range migrations {
		if !columns[column] {
			if _, err := s.db.ExecContext(ctx, statement); err != nil {
				return err
			}
			if column == "day_count" {
				_, _ = s.db.ExecContext(ctx, "UPDATE pages SET day_count = COALESCE(visit_count, 0) WHERE day_count = 0")
			}
		}
	}

	for _, statement := range []string{
		"CREATE INDEX IF NOT EXISTS idx_pages_blacklisted ON pages(is_blacklisted)",
		"CREATE INDEX IF NOT EXISTS idx_pages_canonical_key ON pages(canonical_key)",
		"CREATE INDEX IF NOT EXISTS idx_pages_day_count ON pages(day_count DESC)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_blacklist_type_pattern ON blacklist(type, pattern)",
		"DELETE FROM blacklist WHERE type = 'path' AND pattern LIKE '/%'",
	} {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	blacklistColumns, err := s.tableColumns(ctx, "blacklist")
	if err != nil {
		return err
	}
	if !blacklistColumns["type"] {
		if _, err := s.db.ExecContext(ctx, "ALTER TABLE blacklist ADD COLUMN type TEXT NOT NULL DEFAULT 'url'"); err != nil {
			return err
		}
	}

	insightColumns, err := s.tableColumns(ctx, "insights")
	if err != nil {
		return err
	}
	for column, statement := range map[string]string{
		"user_intent": "ALTER TABLE insights ADD COLUMN user_intent TEXT",
		"key_points":  "ALTER TABLE insights ADD COLUMN key_points TEXT",
		"value":       "ALTER TABLE insights ADD COLUMN value TEXT",
		"next_action": "ALTER TABLE insights ADD COLUMN next_action TEXT",
	} {
		if !insightColumns[column] {
			if _, err := s.db.ExecContext(ctx, statement); err != nil {
				return err
			}
		}
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM insights WHERE id NOT IN (SELECT MAX(id) FROM insights GROUP BY page_id)`); err != nil {
		return err
	}
	return s.backfillPageCanonicalFields(ctx)
}

func (s *Store) tableColumns(ctx context.Context, table string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	return columns, rows.Err()
}

func (s *Store) backfillPageCanonicalFields(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, "SELECT id, url, COALESCE(canonical_url, ''), COALESCE(canonical_key, ''), COALESCE(last_visit_time, '') FROM pages")
	if err != nil {
		return err
	}
	defer rows.Close()
	type rowData struct {
		id           int
		rawURL       string
		canonicalURL string
		canonicalKey string
		lastVisit    string
	}
	items := []rowData{}
	for rows.Next() {
		var item rowData
		if err := rows.Scan(&item.id, &item.rawURL, &item.canonicalURL, &item.canonicalKey, &item.lastVisit); err != nil {
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range items {
		canonicalURL, canonicalKey := urlutil.NormalizePageURL(item.rawURL, item.lastVisit)
		if canonicalKey == "" {
			continue
		}
		if item.canonicalURL != canonicalURL || item.canonicalKey != canonicalKey {
			if _, err := s.db.ExecContext(ctx, "UPDATE pages SET canonical_url = ?, canonical_key = ? WHERE id = ?", canonicalURL, canonicalKey, item.id); err != nil {
				return err
			}
		}
	}
	return s.mergeDuplicateCanonicalPages(ctx)
}

func (s *Store) mergeDuplicateCanonicalPages(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT canonical_key FROM pages WHERE canonical_key IS NOT NULL GROUP BY canonical_key HAVING COUNT(*) > 1`)
	if err != nil {
		return err
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, key := range keys {
		pageRows, err := s.db.QueryContext(ctx, "SELECT id, COALESCE(day_count, 0) FROM pages WHERE canonical_key = ? ORDER BY last_visit_time DESC, id DESC", key)
		if err != nil {
			return err
		}
		ids := []int{}
		totalDayCount := 0
		for pageRows.Next() {
			var id, dayCount int
			if err := pageRows.Scan(&id, &dayCount); err != nil {
				pageRows.Close()
				return err
			}
			ids = append(ids, id)
			totalDayCount += dayCount
		}
		pageRows.Close()
		if len(ids) <= 1 {
			continue
		}
		keeper := ids[0]
		duplicates := ids[1:]
		if _, err := s.db.ExecContext(ctx, `UPDATE pages SET day_count = ?, first_visit_time = (SELECT MIN(first_visit_time) FROM pages WHERE canonical_key = ?), is_bookmarked = (SELECT MAX(is_bookmarked) FROM pages WHERE canonical_key = ?), bookmark_title = COALESCE(bookmark_title, (SELECT bookmark_title FROM pages WHERE canonical_key = ? AND bookmark_title IS NOT NULL LIMIT 1)), bookmark_folder = COALESCE(bookmark_folder, (SELECT bookmark_folder FROM pages WHERE canonical_key = ? AND bookmark_folder IS NOT NULL LIMIT 1)), bookmark_added_at = COALESCE(bookmark_added_at, (SELECT bookmark_added_at FROM pages WHERE canonical_key = ? AND bookmark_added_at IS NOT NULL LIMIT 1)), is_github_starred = (SELECT MAX(is_github_starred) FROM pages WHERE canonical_key = ?), updated_at = CURRENT_TIMESTAMP WHERE id = ?`, totalDayCount, key, key, key, key, key, key, keeper); err != nil {
			return err
		}
		placeholder := placeholders(len(duplicates))
		args := intArgs(duplicates)
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf("UPDATE insights SET page_id = ? WHERE page_id IN (%s)", placeholder), append([]any{keeper}, args...)...); err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM pages WHERE id IN (%s)", placeholder), args...); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetPagesCount(ctx context.Context) (int, error) {
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pages WHERE is_blacklisted = 0").Scan(&total)
	return total, err
}

func (s *Store) CreatePage(ctx context.Context, page models.PageCreate) (int, error) {
	result, err := s.db.ExecContext(ctx, `INSERT INTO pages (url, canonical_url, canonical_key, title, domain, day_count, is_bookmarked, bookmark_title, bookmark_folder, bookmark_added_at, is_github_starred, last_visit_time, first_visit_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		page.URL, page.CanonicalURL, page.CanonicalKey, page.Title, page.Domain, page.DayCount, boolInt(page.IsBookmarked), page.BookmarkTitle, page.BookmarkFolder, page.BookmarkAddedAt, boolInt(page.IsGitHubStarred), page.LastVisitTime, page.FirstVisitTime)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (s *Store) UpdatePage(ctx context.Context, pageID int, page models.PageUpdate) error {
	sets := []string{}
	args := []any{}
	add := func(column string, value any) {
		sets = append(sets, column+" = ?")
		args = append(args, value)
	}
	if page.URL != nil {
		add("url", *page.URL)
	}
	if page.CanonicalURL != nil {
		add("canonical_url", *page.CanonicalURL)
	}
	if page.CanonicalKey != nil {
		add("canonical_key", *page.CanonicalKey)
	}
	if page.Title != nil {
		add("title", *page.Title)
	}
	if page.DayCount != nil {
		add("day_count", *page.DayCount)
	}
	if page.IsBookmarked != nil {
		add("is_bookmarked", boolInt(*page.IsBookmarked))
	}
	if page.BookmarkTitle != nil {
		add("bookmark_title", *page.BookmarkTitle)
	}
	if page.BookmarkFolder != nil {
		add("bookmark_folder", *page.BookmarkFolder)
	}
	if page.BookmarkAddedAt != nil {
		add("bookmark_added_at", *page.BookmarkAddedAt)
	}
	if page.IsGitHubStarred != nil {
		add("is_github_starred", boolInt(*page.IsGitHubStarred))
	}
	if page.LastVisitTime != nil {
		add("last_visit_time", *page.LastVisitTime)
	}
	if page.FirstVisitTime != nil {
		add("first_visit_time", *page.FirstVisitTime)
	}
	if len(sets) == 0 {
		return nil
	}
	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, pageID)
	_, err := s.db.ExecContext(ctx, "UPDATE pages SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
	return err
}

func (s *Store) GetPageByCanonicalKey(ctx context.Context, canonicalKey string) (*models.Page, error) {
	rows, err := s.db.QueryContext(ctx, selectPageSQL()+" WHERE p.canonical_key = ? LIMIT 1", canonicalKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		page, err := scanPage(rows)
		return &page, err
	}
	return nil, rows.Err()
}

func (s *Store) GetPage(ctx context.Context, pageID int) (*models.Page, error) {
	rows, err := s.db.QueryContext(ctx, selectPageSQL()+" WHERE p.id = ?", pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		page, err := scanPage(rows)
		return &page, err
	}
	return nil, rows.Err()
}

func (s *Store) ListPages(ctx context.Context, query string, limit int, offset int, sort string) ([]models.Page, int, error) {
	where, args := pageWhere(query, "", "", "")
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pages p "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	sortSQL := "p.last_visit_time"
	switch sort {
	case "day_count":
		sortSQL = "p.day_count"
	case "visit_count":
		sortSQL = visitCountSQL("p", false)
	case "created_at":
		sortSQL = "p.created_at"
	}
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, selectPageSQL()+" "+where+" ORDER BY "+sortSQL+" DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	pages, err := scanPages(rows)
	return pages, total, err
}

func (s *Store) ListPageGroups(ctx context.Context, query string, domain string, cursor string, dateFrom string, dateTo string, limit int) ([]models.PageDateGroup, int, bool, *string, error) {
	where, args := pageWhere(query, domain, dateFrom, dateTo)
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pages p "+where, args...).Scan(&total); err != nil {
		return nil, 0, false, nil, err
	}

	dateWhere := where
	dateArgs := append([]any{}, args...)
	if cursor != "" {
		if dateWhere == "" {
			dateWhere = "WHERE date(p.last_visit_time) < ?"
		} else {
			dateWhere += " AND date(p.last_visit_time) < ?"
		}
		dateArgs = append(dateArgs, cursor)
	}
	dateArgs = append(dateArgs, limit+1)
	dateRows, err := s.db.QueryContext(ctx, `SELECT date(p.last_visit_time) AS date_key, max(p.last_visit_time) AS latest_visit_time FROM pages p `+dateWhere+` GROUP BY date_key ORDER BY latest_visit_time DESC LIMIT ?`, dateArgs...)
	if err != nil {
		return nil, 0, false, nil, err
	}
	dateKeys := []string{}
	for dateRows.Next() {
		var dateKey, latest string
		if err := dateRows.Scan(&dateKey, &latest); err != nil {
			dateRows.Close()
			return nil, 0, false, nil, err
		}
		dateKeys = append(dateKeys, dateKey)
	}
	dateRows.Close()
	if len(dateKeys) == 0 {
		return []models.PageDateGroup{}, total, false, nil, nil
	}
	hasMore := len(dateKeys) > limit
	if hasMore {
		dateKeys = dateKeys[:limit]
	}

	groupWhere, groupArgs := pageWhere(query, domain, dateFrom, dateTo)
	groupWhere += " AND date(p.last_visit_time) IN (" + placeholders(len(dateKeys)) + ")"
	for _, key := range dateKeys {
		groupArgs = append(groupArgs, key)
	}
	rows, err := s.db.QueryContext(ctx, selectPageWithDateSQL()+" "+groupWhere+" ORDER BY p.last_visit_time DESC, p.id DESC", groupArgs...)
	if err != nil {
		return nil, 0, false, nil, err
	}
	defer rows.Close()

	groupsByKey := map[string][]models.Page{}
	for rows.Next() {
		page, dateKey, err := scanPageWithDate(rows)
		if err != nil {
			return nil, 0, false, nil, err
		}
		groupsByKey[dateKey] = append(groupsByKey[dateKey], page)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, false, nil, err
	}
	groups := make([]models.PageDateGroup, 0, len(dateKeys))
	for _, key := range dateKeys {
		groups = append(groups, models.PageDateGroup{DateKey: key, Title: FormatGroupTitle(key), Pages: groupsByKey[key]})
	}
	var nextCursor *string
	if hasMore {
		value := dateKeys[len(dateKeys)-1]
		nextCursor = &value
	}
	return groups, total, hasMore, nextCursor, nil
}

func (s *Store) ListDomains(ctx context.Context, sort string) ([]models.DomainSummary, error) {
	sortSQL := "last_visit_time DESC, domain ASC"
	if sort == "visits" {
		sortSQL = "visit_count DESC, last_visit_time DESC, domain ASC"
	}

	rows, err := s.db.QueryContext(ctx, `SELECT p.domain, COUNT(*) AS page_count, COALESCE(SUM(p.day_count), 0) AS visit_count, MAX(p.last_visit_time) AS last_visit_time FROM pages p WHERE p.is_blacklisted = 0 AND COALESCE(p.domain, '') != '' GROUP BY p.domain ORDER BY `+sortSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	domains := []models.DomainSummary{}
	for rows.Next() {
		var domain models.DomainSummary
		if err := rows.Scan(&domain.Domain, &domain.PageCount, &domain.VisitCount, &domain.LastVisitTime); err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	return domains, rows.Err()
}

func pageWhere(query string, domain string, dateFrom string, dateTo string) (string, []any) {
	clauses := []string{"p.is_blacklisted = 0"}
	args := []any{}
	if strings.TrimSpace(query) != "" {
		clauses = append(clauses, "(p.title LIKE ? OR p.domain LIKE ? OR p.url LIKE ?)")
		term := "%" + strings.TrimSpace(query) + "%"
		args = append(args, term, term, term)
	}
	if strings.TrimSpace(domain) != "" {
		clauses = append(clauses, "p.domain = ?")
		args = append(args, strings.TrimSpace(domain))
	}
	if dateFrom != "" {
		clauses = append(clauses, "date(p.last_visit_time) >= ?")
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		clauses = append(clauses, "date(p.last_visit_time) <= ?")
		args = append(args, dateTo)
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func (s *Store) ListBlacklistEntries(ctx context.Context) ([]models.BlacklistEntry, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, type, pattern, created_at FROM blacklist ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := []models.BlacklistEntry{}
	for rows.Next() {
		var entry models.BlacklistEntry
		if err := rows.Scan(&entry.ID, &entry.Type, &entry.Pattern, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *Store) CreateBlacklistEntry(ctx context.Context, pattern string, entryType string) (models.BlacklistEntry, error) {
	normalizedType, err := normalizeBlacklistType(entryType)
	if err != nil {
		return models.BlacklistEntry{}, err
	}
	normalizedPattern := strings.TrimSpace(pattern)
	if normalizedPattern == "" {
		return models.BlacklistEntry{}, errors.New("Blacklist pattern is required")
	}
	if _, err := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO blacklist (type, pattern) VALUES (?, ?)", normalizedType, normalizedPattern); err != nil {
		return models.BlacklistEntry{}, err
	}
	var entry models.BlacklistEntry
	err = s.db.QueryRowContext(ctx, "SELECT id, type, pattern, created_at FROM blacklist WHERE type = ? AND pattern = ?", normalizedType, normalizedPattern).Scan(&entry.ID, &entry.Type, &entry.Pattern, &entry.CreatedAt)
	return entry, err
}

func (s *Store) DeleteBlacklistEntry(ctx context.Context, entryID int) (bool, error) {
	result, err := s.db.ExecContext(ctx, "DELETE FROM blacklist WHERE id = ?", entryID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func (s *Store) ApplyBlacklistEntry(ctx context.Context, entryType string, pattern string) (int, error) {
	normalizedType, err := normalizeBlacklistType(entryType)
	if err != nil {
		return 0, err
	}
	normalizedPattern := strings.TrimSpace(pattern)
	switch normalizedType {
	case "url":
		_, err = s.db.ExecContext(ctx, "UPDATE pages SET is_blacklisted = 1 WHERE url = ?", normalizedPattern)
	case "domain":
		domain := strings.ToLower(normalizedPattern)
		_, err = s.db.ExecContext(ctx, "UPDATE pages SET is_blacklisted = 1 WHERE lower(domain) = ? OR lower(domain) LIKE ?", domain, "%."+domain)
	case "path":
		_, err = s.db.ExecContext(ctx, "UPDATE pages SET is_blacklisted = 1 WHERE instr(url, ?) > 0", normalizedPattern)
	}
	if err != nil {
		return 0, err
	}
	var total int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pages WHERE is_blacklisted = 1").Scan(&total)
	return total, err
}

func (s *Store) RebuildBlacklist(ctx context.Context) (int, error) {
	entries, err := s.ListBlacklistEntries(ctx)
	if err != nil {
		return 0, err
	}
	if _, err := s.db.ExecContext(ctx, "UPDATE pages SET is_blacklisted = 0 WHERE is_blacklisted != 0"); err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, nil
	}
	rows, err := s.db.QueryContext(ctx, "SELECT id, url FROM pages")
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	ids := []int{}
	for rows.Next() {
		var id int
		var rawURL string
		if err := rows.Scan(&id, &rawURL); err != nil {
			return 0, err
		}
		if MatchesBlacklistEntries(rawURL, entries) {
			ids = append(ids, id)
		}
	}
	if len(ids) > 0 {
		_, err = s.db.ExecContext(ctx, "UPDATE pages SET is_blacklisted = 1 WHERE id IN ("+placeholders(len(ids))+")", intArgs(ids)...)
	}
	return len(ids), err
}

func (s *Store) IsURLBlacklisted(ctx context.Context, rawURL string) (bool, error) {
	entries, err := s.ListBlacklistEntries(ctx)
	if err != nil {
		return false, err
	}
	return MatchesBlacklistEntries(rawURL, entries), nil
}

func MatchesBlacklistEntries(rawURL string, entries []models.BlacklistEntry) bool {
	parsed, _ := url.Parse(rawURL)
	domain := strings.ToLower(strings.Split(strings.Split(parsed.Host, "@")[len(strings.Split(parsed.Host, "@"))-1], ":")[0])
	path := parsed.Path
	if path == "" {
		path = "/"
	}
	domainPath := domain + path
	for _, entry := range entries {
		pattern := strings.TrimSpace(entry.Pattern)
		if pattern == "" {
			continue
		}
		switch entry.Type {
		case "url":
			if rawURL == pattern {
				return true
			}
		case "domain":
			normalized := strings.ToLower(pattern)
			if domain == normalized || strings.HasSuffix(domain, "."+normalized) {
				return true
			}
		case "path":
			if strings.Contains(strings.ToLower(domainPath), strings.ToLower(pattern)) {
				return true
			}
		}
	}
	return false
}

func normalizeBlacklistType(entryType string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(entryType))
	if normalized == "" {
		normalized = "url"
	}
	switch normalized {
	case "url", "domain", "path":
		return normalized, nil
	default:
		return "", fmt.Errorf("Unsupported blacklist type: %s", entryType)
	}
}

func (s *Store) CreateInsights(ctx context.Context, insights models.InsightsCreate) (int, error) {
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	keywords, _ := json.Marshal(insights.Keywords)
	keyPoints, _ := json.Marshal(insights.KeyPoints)
	var existingID int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM insights WHERE page_id = ? ORDER BY analyzed_at DESC LIMIT 1", insights.PageID).Scan(&existingID)
	if err == nil {
		_, err = s.db.ExecContext(ctx, `UPDATE insights SET summary = ?, type = ?, keywords = ?, user_intent = ?, key_points = ?, value = ?, next_action = ?, agent_name = ?, analyzed_at = CURRENT_TIMESTAMP, expires_at = ? WHERE id = ?`,
			insights.Summary, insights.Type, string(keywords), insights.UserIntent, string(keyPoints), insights.Value, insights.NextAction, insights.AgentName, expiresAt, existingID)
		if err != nil {
			return 0, err
		}
		_, err = s.db.ExecContext(ctx, "DELETE FROM insights WHERE page_id = ? AND id != ?", insights.PageID, existingID)
		return existingID, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO insights (page_id, summary, type, keywords, user_intent, key_points, value, next_action, agent_name, expires_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		insights.PageID, insights.Summary, insights.Type, string(keywords), insights.UserIntent, string(keyPoints), insights.Value, insights.NextAction, insights.AgentName, expiresAt)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (s *Store) GetInsights(ctx context.Context, pageID int) (*models.Insights, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, page_id, summary, type, keywords, user_intent, key_points, value, next_action, agent_name, analyzed_at, expires_at FROM insights WHERE page_id = ? AND expires_at > datetime('now') ORDER BY analyzed_at DESC LIMIT 1`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		insights, err := scanInsights(rows)
		return &insights, err
	}
	return nil, rows.Err()
}

func selectPageSQL() string {
	return `SELECT p.id, p.url, p.canonical_url, p.canonical_key, p.title, p.domain, p.day_count, COALESCE(p.is_bookmarked, 0), p.bookmark_title, p.bookmark_folder, p.bookmark_added_at, COALESCE(p.is_github_starred, 0), p.last_visit_time, p.first_visit_time, p.created_at, p.updated_at, CASE WHEN i.id IS NOT NULL THEN 1 ELSE 0 END as has_insights, i.id, i.summary, i.type, i.keywords, i.user_intent, i.key_points, i.value, i.next_action, i.agent_name, i.analyzed_at, i.expires_at, ` + visitCountSQL("p", true) + ` FROM pages p LEFT JOIN insights i ON p.id = i.page_id AND i.expires_at > datetime('now')`
}

func selectPageWithDateSQL() string {
	return `SELECT p.id, p.url, p.canonical_url, p.canonical_key, p.title, p.domain, p.day_count, COALESCE(p.is_bookmarked, 0), p.bookmark_title, p.bookmark_folder, p.bookmark_added_at, COALESCE(p.is_github_starred, 0), p.last_visit_time, p.first_visit_time, p.created_at, p.updated_at, CASE WHEN i.id IS NOT NULL THEN 1 ELSE 0 END as has_insights, i.id, i.summary, i.type, i.keywords, i.user_intent, i.key_points, i.value, i.next_action, i.agent_name, i.analyzed_at, i.expires_at, ` + visitCountSQL("p", true) + `, date(p.last_visit_time) as date_key FROM pages p LEFT JOIN insights i ON p.id = i.page_id AND i.expires_at > datetime('now')`
}

func visitCountSQL(alias string, includeAlias bool) string {
	suffix := ""
	if includeAlias {
		suffix = " as visit_count"
	}
	return `(SELECT COALESCE(SUM(p2.day_count), ` + alias + `.day_count) FROM pages p2 WHERE p2.is_blacklisted = 0 AND p2.canonical_key LIKE substr(` + alias + `.canonical_key, 1, length(` + alias + `.canonical_key) - 11) || ':%')` + suffix
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPage(rows rowScanner) (models.Page, error) {
	page, _, err := scanPageCore(rows, false)
	return page, err
}

func scanPageWithDate(rows rowScanner) (models.Page, string, error) {
	return scanPageCore(rows, true)
}

func scanPageCore(rows rowScanner, withDate bool) (models.Page, string, error) {
	var page models.Page
	var canonicalURL, canonicalKey, title, domain, bookmarkTitle, bookmarkFolder, bookmarkAddedAt, firstVisit sql.NullString
	var insightID sql.NullInt64
	var summary, insightType, keywords, userIntent, keyPoints, value, nextAction, agentName, analyzedAt, expiresAt sql.NullString
	var isBookmarked, isGitHubStarred, hasInsights int
	dest := []any{
		&page.ID, &page.URL, &canonicalURL, &canonicalKey, &title, &domain, &page.DayCount, &isBookmarked, &bookmarkTitle, &bookmarkFolder, &bookmarkAddedAt, &isGitHubStarred, &page.LastVisitTime, &firstVisit, &page.CreatedAt, &page.UpdatedAt,
		&hasInsights, &insightID, &summary, &insightType, &keywords, &userIntent, &keyPoints, &value, &nextAction, &agentName, &analyzedAt, &expiresAt, &page.VisitCount,
	}
	var dateKey string
	if withDate {
		dest = append(dest, &dateKey)
	}
	if err := rows.Scan(dest...); err != nil {
		return page, "", err
	}
	if canonicalURL.Valid {
		page.CanonicalURL = &canonicalURL.String
	}
	if canonicalKey.Valid {
		page.CanonicalKey = &canonicalKey.String
	}
	page.Title = title.String
	page.Domain = domain.String
	page.IsBookmarked = isBookmarked != 0
	page.IsGitHubStarred = isGitHubStarred != 0
	if bookmarkTitle.Valid {
		page.BookmarkTitle = &bookmarkTitle.String
	}
	if bookmarkFolder.Valid {
		page.BookmarkFolder = &bookmarkFolder.String
	}
	if bookmarkAddedAt.Valid {
		page.BookmarkAddedAt = &bookmarkAddedAt.String
	}
	page.FirstVisitTime = firstVisit.String
	page.HasInsights = hasInsights != 0
	if page.HasInsights && insightID.Valid {
		page.Insights = buildInsights(int(insightID.Int64), page.ID, summary, insightType, keywords, userIntent, keyPoints, value, nextAction, agentName, analyzedAt, expiresAt)
	}
	return page, dateKey, nil
}

func scanPages(rows *sql.Rows) ([]models.Page, error) {
	pages := []models.Page{}
	for rows.Next() {
		page, err := scanPage(rows)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, rows.Err()
}

func scanInsights(rows rowScanner) (models.Insights, error) {
	var id, pageID int
	var summary, insightType, keywords, userIntent, keyPoints, value, nextAction, agentName, analyzedAt, expiresAt sql.NullString
	if err := rows.Scan(&id, &pageID, &summary, &insightType, &keywords, &userIntent, &keyPoints, &value, &nextAction, &agentName, &analyzedAt, &expiresAt); err != nil {
		return models.Insights{}, err
	}
	return *buildInsights(id, pageID, summary, insightType, keywords, userIntent, keyPoints, value, nextAction, agentName, analyzedAt, expiresAt), nil
}

func buildInsights(id int, pageID int, summary, insightType, keywords, userIntent, keyPoints, value, nextAction, agentName, analyzedAt, expiresAt sql.NullString) *models.Insights {
	insights := &models.Insights{
		ID:         id,
		PageID:     pageID,
		Summary:    summary.String,
		Type:       insightType.String,
		Keywords:   decodeStringArray(keywords.String),
		KeyPoints:  decodeStringArray(keyPoints.String),
		AgentName:  agentName.String,
		AnalyzedAt: analyzedAt.String,
		ExpiresAt:  expiresAt.String,
	}
	if userIntent.Valid {
		insights.UserIntent = &userIntent.String
	}
	if value.Valid {
		insights.Value = &value.String
	}
	if nextAction.Valid {
		insights.NextAction = &nextAction.String
	}
	return insights
}

func decodeStringArray(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []string{}
	}
	return items
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

func intArgs(values []int) []any {
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value
	}
	return args
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func FormatGroupTitle(dateKey string) string {
	weekdays := []string{"一", "二", "三", "四", "五", "六", "日"}
	groupDate, err := time.ParseInLocation("2006-01-02", dateKey, time.Local)
	if err != nil {
		return dateKey
	}
	today := time.Now().In(time.Local)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	fullDate := fmt.Sprintf("%d年%d月%d日星期%s", groupDate.Year(), groupDate.Month(), groupDate.Day(), weekdays[(int(groupDate.Weekday())+6)%7])
	if groupDate.Equal(todayDate) {
		return "今天 - " + fullDate
	}
	if groupDate.Equal(todayDate.AddDate(0, 0, -1)) {
		return "昨天 - " + fullDate
	}
	return fullDate
}
