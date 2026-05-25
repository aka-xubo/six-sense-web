-- Six Sense Web Database Schema

CREATE TABLE IF NOT EXISTS pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL,
    canonical_url TEXT,
    canonical_key TEXT,
    title TEXT,
    domain TEXT,
    day_count INTEGER DEFAULT 0,
    is_bookmarked INTEGER DEFAULT 0,
    bookmark_title TEXT,
    bookmark_folder TEXT,
    bookmark_added_at TIMESTAMP,
    is_github_starred INTEGER DEFAULT 0,
    last_visit_time TIMESTAMP,
    first_visit_time TIMESTAMP,
    is_blacklisted INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pages_domain ON pages(domain);
CREATE INDEX IF NOT EXISTS idx_pages_last_visit ON pages(last_visit_time DESC);
CREATE INDEX IF NOT EXISTS idx_pages_url ON pages(url);

CREATE TABLE IF NOT EXISTS insights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    page_id INTEGER NOT NULL,
    summary TEXT,
    type TEXT,
    keywords TEXT,
    user_intent TEXT,
    key_points TEXT,
    value TEXT,
    next_action TEXT,
    agent_name TEXT,
    analyzed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_insights_page_id ON insights(page_id);
CREATE INDEX IF NOT EXISTS idx_insights_expires_at ON insights(expires_at);

CREATE TABLE IF NOT EXISTS blacklist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL DEFAULT 'url',
    pattern TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
