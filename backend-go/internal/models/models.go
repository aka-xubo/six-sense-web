package models

type Insights struct {
	ID         int      `json:"id"`
	PageID     int      `json:"page_id"`
	Summary    string   `json:"summary"`
	Type       string   `json:"type"`
	Keywords   []string `json:"keywords"`
	UserIntent *string  `json:"user_intent,omitempty"`
	KeyPoints  []string `json:"key_points,omitempty"`
	Value      *string  `json:"value,omitempty"`
	NextAction *string  `json:"next_action,omitempty"`
	AgentName  string   `json:"agent_name"`
	AnalyzedAt string   `json:"analyzed_at"`
	ExpiresAt  string   `json:"expires_at"`
}

type Page struct {
	ID              int       `json:"id"`
	URL             string    `json:"url"`
	CanonicalURL    *string   `json:"canonical_url"`
	CanonicalKey    *string   `json:"canonical_key"`
	Title           string    `json:"title"`
	Domain          string    `json:"domain"`
	DayCount        int       `json:"day_count"`
	VisitCount      int       `json:"visit_count"`
	IsBookmarked    bool      `json:"is_bookmarked"`
	BookmarkTitle   *string   `json:"bookmark_title,omitempty"`
	BookmarkFolder  *string   `json:"bookmark_folder,omitempty"`
	BookmarkAddedAt *string   `json:"bookmark_added_at,omitempty"`
	IsGitHubStarred bool      `json:"is_github_starred"`
	LastVisitTime   string    `json:"last_visit_time"`
	FirstVisitTime  string    `json:"first_visit_time"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
	HasInsights     bool      `json:"has_insights"`
	Insights        *Insights `json:"insights,omitempty"`
}

type PageListResponse struct {
	Pages   []Page `json:"pages"`
	Total   int    `json:"total"`
	HasMore bool   `json:"has_more"`
}

type PageDateGroup struct {
	DateKey string `json:"date_key"`
	Title   string `json:"title"`
	Pages   []Page `json:"pages"`
}

type PageGroupListResponse struct {
	Groups     []PageDateGroup `json:"groups"`
	Total      int             `json:"total"`
	HasMore    bool            `json:"has_more"`
	NextCursor *string         `json:"next_cursor"`
}

type DomainSummary struct {
	Domain        string `json:"domain"`
	PageCount     int    `json:"page_count"`
	VisitCount    int    `json:"visit_count"`
	LastVisitTime string `json:"last_visit_time"`
}

type DomainListResponse struct {
	Domains []DomainSummary `json:"domains"`
}

type SyncRequest struct {
	Months int `json:"months"`
}

type SyncResponse struct {
	Status       string `json:"status"`
	NewPages     int    `json:"new_pages"`
	UpdatedPages int    `json:"updated_pages"`
	TotalPages   int    `json:"total_pages"`
	SyncTime     string `json:"sync_time"`
}

type BlacklistEntry struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Pattern   string `json:"pattern"`
	CreatedAt string `json:"created_at"`
}

type BlacklistCreate struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

type BlacklistAddPageResponse struct {
	Entry       BlacklistEntry `json:"entry"`
	HiddenPages int            `json:"hidden_pages"`
}

type BlacklistListResponse struct {
	Entries []BlacklistEntry `json:"entries"`
}

type BlacklistDeleteResponse struct {
	HiddenPages int `json:"hidden_pages"`
}

type AgentInfo struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
	Version     *string `json:"version"`
	Available   bool    `json:"available"`
}

type AgentsResponse struct {
	Agents []AgentInfo `json:"agents"`
}

type PageCreate struct {
	URL             string
	CanonicalURL    string
	CanonicalKey    string
	Title           string
	Domain          string
	DayCount        int
	VisitCount      int
	IsBookmarked    bool
	BookmarkTitle   *string
	BookmarkFolder  *string
	BookmarkAddedAt *string
	IsGitHubStarred bool
	LastVisitTime   string
	FirstVisitTime  string
}

type PageUpdate struct {
	URL             *string
	CanonicalURL    *string
	CanonicalKey    *string
	Title           *string
	DayCount        *int
	IsBookmarked    *bool
	BookmarkTitle   *string
	BookmarkFolder  *string
	BookmarkAddedAt *string
	IsGitHubStarred *bool
	LastVisitTime   *string
	FirstVisitTime  *string
}

type InsightsCreate struct {
	PageID     int
	Summary    string
	Type       string
	Keywords   []string
	UserIntent string
	KeyPoints  []string
	Value      string
	NextAction string
	AgentName  string
}
