-- Six Sense Web Database Schema

-- pages 表 (核心数据)
CREATE TABLE IF NOT EXISTS pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT UNIQUE NOT NULL,
    title TEXT,
    domain TEXT,
    visit_count INTEGER DEFAULT 0,
    last_visit_time TIMESTAMP,
    first_visit_time TIMESTAMP,
    is_blacklisted INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pages_domain ON pages(domain);
CREATE INDEX IF NOT EXISTS idx_pages_last_visit ON pages(last_visit_time DESC);
CREATE INDEX IF NOT EXISTS idx_pages_visit_count ON pages(visit_count DESC);
CREATE INDEX IF NOT EXISTS idx_pages_url ON pages(url);

-- insights 表 (AI 分析结果)
CREATE TABLE IF NOT EXISTS insights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    page_id INTEGER NOT NULL,
    summary TEXT,
    type TEXT,
    keywords TEXT,  -- JSON array: ["keyword1", "keyword2", "keyword3"]
    agent_name TEXT,  -- 使用的 AI Agent (claude/cursor/codex)
    analyzed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,  -- 7天后过期
    FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_insights_page_id ON insights(page_id);
CREATE INDEX IF NOT EXISTS idx_insights_expires_at ON insights(expires_at);

-- blacklist 表 (黑名单规则，Phase 2 使用)
CREATE TABLE IF NOT EXISTS blacklist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL DEFAULT 'url',
    pattern TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
