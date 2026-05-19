"""Database operations for Six Sense Web"""
import aiosqlite
import json
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional
from urllib.parse import urlparse
from app.config import settings
from app.models import PageCreate, PageUpdate, InsightsCreate


class Database:
    """Database manager for Six Sense Web"""

    def __init__(self, db_path: Path = settings.db_path):
        self.db_path = db_path
        self._connection: Optional[aiosqlite.Connection] = None

    async def connect(self):
        """Connect to database"""
        self._connection = await aiosqlite.connect(str(self.db_path))
        self._connection.row_factory = aiosqlite.Row
        await self._init_schema()

    async def disconnect(self):
        """Disconnect from database"""
        if self._connection:
            await self._connection.close()
            self._connection = None

    async def _init_schema(self):
        """Initialize database schema"""
        schema_path = Path(__file__).parent / "schema.sql"
        with open(schema_path, 'r') as f:
            schema = f.read()

        await self._connection.executescript(schema)
        await self._ensure_migrations()
        await self._connection.commit()

    async def _ensure_migrations(self):
        """Apply lightweight migrations for existing local databases."""
        cursor = await self._connection.execute("PRAGMA table_info(pages)")
        columns = {row['name'] for row in await cursor.fetchall()}

        if 'is_blacklisted' not in columns:
            await self._connection.execute(
                "ALTER TABLE pages ADD COLUMN is_blacklisted INTEGER DEFAULT 0"
            )

        await self._connection.execute(
            "CREATE INDEX IF NOT EXISTS idx_pages_blacklisted ON pages(is_blacklisted)"
        )

        cursor = await self._connection.execute("PRAGMA table_info(blacklist)")
        blacklist_columns = {row['name'] for row in await cursor.fetchall()}

        if 'type' not in blacklist_columns:
            await self._connection.execute(
                "ALTER TABLE blacklist ADD COLUMN type TEXT NOT NULL DEFAULT 'url'"
            )

        await self._connection.execute(
            "CREATE UNIQUE INDEX IF NOT EXISTS idx_blacklist_type_pattern ON blacklist(type, pattern)"
        )
        await self._connection.execute(
            "DELETE FROM blacklist WHERE type = 'path' AND pattern LIKE '/%'"
        )

    # Pages operations
    async def create_page(self, page: PageCreate) -> int:
        """Create a new page"""
        cursor = await self._connection.execute(
            """
            INSERT INTO pages (url, title, domain, visit_count, last_visit_time, first_visit_time)
            VALUES (?, ?, ?, ?, ?, ?)
            """,
            (
                page.url,
                page.title,
                page.domain,
                page.visit_count,
                page.last_visit_time,
                page.first_visit_time or page.last_visit_time
            )
        )
        await self._connection.commit()
        return cursor.lastrowid

    async def get_page(self, page_id: int) -> Optional[dict]:
        """Get page by ID"""
        cursor = await self._connection.execute(
            """
            SELECT p.*,
                   CASE WHEN i.id IS NOT NULL THEN 1 ELSE 0 END as has_insights,
                   i.id as insight_id,
                   i.summary,
                   i.type,
                   i.keywords,
                   i.agent_name,
                   i.analyzed_at,
                   i.expires_at
            FROM pages p
            LEFT JOIN insights i ON p.id = i.page_id AND i.expires_at > datetime('now')
            WHERE p.id = ?
            """,
            (page_id,)
        )
        row = await cursor.fetchone()
        if not row:
            return None

        return self._row_to_page_dict(row)

    async def get_page_by_url(self, url: str) -> Optional[dict]:
        """Get page by URL"""
        cursor = await self._connection.execute(
            """
            SELECT p.*,
                   CASE WHEN i.id IS NOT NULL THEN 1 ELSE 0 END as has_insights
            FROM pages p
            LEFT JOIN insights i ON p.id = i.page_id AND i.expires_at > datetime('now')
            WHERE p.url = ?
            """,
            (url,)
        )
        row = await cursor.fetchone()
        if not row:
            return None

        return dict(row)

    async def update_page(self, page_id: int, page: PageUpdate) -> bool:
        """Update page"""
        updates = []
        values = []

        if page.title is not None:
            updates.append("title = ?")
            values.append(page.title)
        if page.visit_count is not None:
            updates.append("visit_count = ?")
            values.append(page.visit_count)
        if page.last_visit_time is not None:
            updates.append("last_visit_time = ?")
            values.append(page.last_visit_time)

        if not updates:
            return False

        updates.append("updated_at = CURRENT_TIMESTAMP")
        values.append(page_id)

        query = f"UPDATE pages SET {', '.join(updates)} WHERE id = ?"
        await self._connection.execute(query, values)
        await self._connection.commit()
        return True

    async def upsert_page(self, page: PageCreate) -> int:
        """Insert or update page"""
        existing = await self.get_page_by_url(page.url)

        if existing:
            # Update existing page
            await self.update_page(
                existing['id'],
                PageUpdate(
                    title=page.title,
                    visit_count=page.visit_count,
                    last_visit_time=page.last_visit_time
                )
            )
            return existing['id']
        else:
            # Create new page
            return await self.create_page(page)

    async def list_pages(
        self,
        query: Optional[str] = None,
        limit: int = 50,
        offset: int = 0,
        sort: str = "last_visit_time"
    ) -> tuple[list[dict], int]:
        """List pages with optional search and pagination"""

        # Build WHERE clause
        where_clauses = ["p.is_blacklisted = 0"]
        params = []

        if query:
            where_clauses.append("(p.title LIKE ? OR p.domain LIKE ? OR p.url LIKE ?)")
            search_term = f"%{query}%"
            params.extend([search_term, search_term, search_term])

        where_sql = f"WHERE {' AND '.join(where_clauses)}" if where_clauses else ""

        # Validate sort column
        valid_sorts = ["last_visit_time", "visit_count", "created_at"]
        if sort not in valid_sorts:
            sort = "last_visit_time"

        # Get total count
        count_query = f"SELECT COUNT(*) as total FROM pages p {where_sql}"
        cursor = await self._connection.execute(count_query, params)
        row = await cursor.fetchone()
        total = row['total']

        # Get pages
        query_sql = f"""
            SELECT p.*,
                   CASE WHEN i.id IS NOT NULL THEN 1 ELSE 0 END as has_insights,
                   i.id as insight_id,
                   i.summary,
                   i.type,
                   i.keywords,
                   i.agent_name,
                   i.analyzed_at,
                   i.expires_at
            FROM pages p
            LEFT JOIN insights i ON p.id = i.page_id AND i.expires_at > datetime('now')
            {where_sql}
            ORDER BY p.{sort} DESC
            LIMIT ? OFFSET ?
        """
        params.extend([limit, offset])

        cursor = await self._connection.execute(query_sql, params)
        rows = await cursor.fetchall()

        pages = [self._row_to_page_dict(row) for row in rows]

        return pages, total

    async def list_page_groups(
        self,
        query: Optional[str] = None,
        cursor: Optional[str] = None,
        limit: int = 1,
        date_from: Optional[str] = None,
        date_to: Optional[str] = None
    ) -> tuple[list[dict], int, bool, Optional[str]]:
        """List pages grouped by local date.

        Pagination happens at the date-group level, so a loaded date is always
        returned as one complete block.
        """

        where_clauses = ["p.is_blacklisted = 0"]
        params = []

        if query:
            where_clauses.append("(p.title LIKE ? OR p.domain LIKE ? OR p.url LIKE ?)")
            search_term = f"%{query}%"
            params.extend([search_term, search_term, search_term])

        if date_from:
            where_clauses.append("date(p.last_visit_time) >= ?")
            params.append(date_from)

        if date_to:
            where_clauses.append("date(p.last_visit_time) <= ?")
            params.append(date_to)

        where_sql = f"WHERE {' AND '.join(where_clauses)}" if where_clauses else ""

        count_query = f"SELECT COUNT(*) as total FROM pages p {where_sql}"
        cursor_obj = await self._connection.execute(count_query, params)
        row = await cursor_obj.fetchone()
        total = row['total']

        date_where_clauses = list(where_clauses)
        date_params = list(params)

        if cursor:
            date_where_clauses.append("date(p.last_visit_time) < ?")
            date_params.append(cursor)

        date_where_sql = f"WHERE {' AND '.join(date_where_clauses)}" if date_where_clauses else ""
        date_query = f"""
            SELECT date(p.last_visit_time) as date_key,
                   max(p.last_visit_time) as latest_visit_time
            FROM pages p
            {date_where_sql}
            GROUP BY date_key
            ORDER BY latest_visit_time DESC
            LIMIT ?
        """
        date_params.append(limit + 1)

        cursor_obj = await self._connection.execute(date_query, date_params)
        date_rows = await cursor_obj.fetchall()
        has_more = len(date_rows) > limit
        visible_date_rows = date_rows[:limit]

        if not visible_date_rows:
            return [], total, False, None

        date_keys = [row['date_key'] for row in visible_date_rows]
        placeholders = ", ".join("?" for _ in date_keys)

        pages_where_clauses = list(where_clauses)
        pages_params = list(params)
        pages_where_clauses.append(f"date(p.last_visit_time) IN ({placeholders})")
        pages_params.extend(date_keys)

        pages_where_sql = f"WHERE {' AND '.join(pages_where_clauses)}"
        pages_query = f"""
            SELECT p.*,
                   date(p.last_visit_time) as date_key,
                   CASE WHEN i.id IS NOT NULL THEN 1 ELSE 0 END as has_insights,
                   i.id as insight_id,
                   i.summary,
                   i.type,
                   i.keywords,
                   i.agent_name,
                   i.analyzed_at,
                   i.expires_at
            FROM pages p
            LEFT JOIN insights i ON p.id = i.page_id AND i.expires_at > datetime('now')
            {pages_where_sql}
            ORDER BY p.last_visit_time DESC, p.id DESC
        """

        cursor_obj = await self._connection.execute(pages_query, pages_params)
        rows = await cursor_obj.fetchall()

        groups_by_key = {
            date_key: {
                'date_key': date_key,
                'pages': []
            }
            for date_key in date_keys
        }

        for row in rows:
            groups_by_key[row['date_key']]['pages'].append(self._row_to_page_dict(row))

        groups = [groups_by_key[date_key] for date_key in date_keys]
        next_cursor = date_keys[-1] if has_more else None

        return groups, total, has_more, next_cursor

    async def get_pages_count(self) -> int:
        """Get total pages count"""
        cursor = await self._connection.execute(
            "SELECT COUNT(*) as total FROM pages WHERE is_blacklisted = 0"
        )
        row = await cursor.fetchone()
        return row['total']

    # Blacklist operations
    async def list_blacklist_entries(self) -> list[dict]:
        """List all blacklist entries"""
        cursor = await self._connection.execute(
            "SELECT * FROM blacklist ORDER BY created_at DESC"
        )
        rows = await cursor.fetchall()
        return [dict(row) for row in rows]

    async def delete_blacklist_entry(self, entry_id: int) -> bool:
        """Delete a blacklist entry by ID"""
        cursor = await self._connection.execute(
            "DELETE FROM blacklist WHERE id = ?",
            (entry_id,)
        )
        await self._connection.commit()
        return cursor.rowcount > 0

    async def create_blacklist_entry(self, pattern: str, entry_type: str = "url") -> dict:
        """Create a blacklist entry or return the existing one"""
        normalized_pattern = pattern.strip()
        normalized_type = self._normalize_blacklist_type(entry_type)
        await self._connection.execute(
            "INSERT OR IGNORE INTO blacklist (type, pattern) VALUES (?, ?)",
            (normalized_type, normalized_pattern)
        )
        await self._connection.commit()

        cursor = await self._connection.execute(
            "SELECT * FROM blacklist WHERE type = ? AND pattern = ?",
            (normalized_type, normalized_pattern)
        )
        row = await cursor.fetchone()
        return dict(row)

    async def rebuild_blacklist(self) -> int:
        """Refresh page blacklist flags and return hidden page count."""
        entries = await self.list_blacklist_entries()
        if not entries:
            await self._connection.execute("UPDATE pages SET is_blacklisted = 0 WHERE is_blacklisted != 0")
            await self._connection.commit()
            return 0

        cursor = await self._connection.execute("SELECT id, url FROM pages")
        rows = await cursor.fetchall()
        page_ids = [
            row['id']
            for row in rows
            if self._matches_blacklist_entries(row['url'], entries)
        ]

        await self._connection.execute("UPDATE pages SET is_blacklisted = 0 WHERE is_blacklisted != 0")

        if page_ids:
            placeholders = ", ".join("?" for _ in page_ids)
            await self._connection.execute(
                f"UPDATE pages SET is_blacklisted = 1 WHERE id IN ({placeholders})",
                page_ids
            )

        await self._connection.commit()

        return len(page_ids)

    async def apply_blacklist_entry(self, entry_type: str, pattern: str) -> int:
        """Apply one blacklist entry and return current hidden page count."""
        normalized_type = self._normalize_blacklist_type(entry_type)
        normalized_pattern = pattern.strip()

        if normalized_type == 'url':
            await self._connection.execute(
                "UPDATE pages SET is_blacklisted = 1 WHERE url = ?",
                (normalized_pattern,)
            )
        elif normalized_type == 'domain':
            normalized_domain = normalized_pattern.lower()
            await self._connection.execute(
                """
                UPDATE pages
                SET is_blacklisted = 1
                WHERE lower(domain) = ?
                   OR lower(domain) LIKE ?
                """,
                (normalized_domain, f"%.{normalized_domain}")
            )
        else:
            await self._connection.execute(
                "UPDATE pages SET is_blacklisted = 1 WHERE instr(url, ?) > 0",
                (normalized_pattern,)
            )

        await self._connection.commit()

        cursor = await self._connection.execute(
            "SELECT COUNT(*) as total FROM pages WHERE is_blacklisted = 1"
        )
        row = await cursor.fetchone()
        return row['total']

    async def is_url_blacklisted(self, url: str) -> bool:
        """Check whether URL matches current blacklist entries."""
        entries = await self.list_blacklist_entries()
        return self._matches_blacklist_entries(url, entries)

    def _normalize_blacklist_type(self, entry_type: str) -> str:
        """Normalize and validate blacklist type."""
        normalized_type = entry_type.strip().lower()
        if normalized_type not in {"url", "domain", "path"}:
            raise ValueError(f"Unsupported blacklist type: {entry_type}")
        return normalized_type

    def _matches_blacklist_entries(self, url: str, entries: list[dict]) -> bool:
        """Check whether URL matches any typed blacklist entry."""
        parsed = urlparse(url)
        domain = parsed.netloc.split('@')[-1].split(':')[0].lower()
        path = parsed.path or "/"
        domain_path = f"{domain}{path}"

        for entry in entries:
            entry_type = entry['type']
            pattern = entry['pattern'].strip()

            if not pattern:
                continue

            if entry_type == 'url' and url == pattern:
                return True

            if entry_type == 'domain':
                normalized_pattern = pattern.lower()
                if domain == normalized_pattern or domain.endswith(f".{normalized_pattern}"):
                    return True

            if entry_type == 'path' and pattern.lower() in domain_path:
                return True

        return False

    # Insights operations
    async def create_insights(self, insights: InsightsCreate) -> int:
        """Create insights for a page"""
        expires_at = datetime.now() + timedelta(days=7)

        cursor = await self._connection.execute(
            """
            INSERT INTO insights (page_id, summary, type, keywords, agent_name, expires_at)
            VALUES (?, ?, ?, ?, ?, ?)
            """,
            (
                insights.page_id,
                insights.summary,
                insights.type,
                json.dumps(insights.keywords),
                insights.agent_name,
                expires_at
            )
        )
        await self._connection.commit()
        return cursor.lastrowid

    async def get_insights(self, page_id: int) -> Optional[dict]:
        """Get insights for a page (if not expired)"""
        cursor = await self._connection.execute(
            """
            SELECT * FROM insights
            WHERE page_id = ? AND expires_at > datetime('now')
            ORDER BY analyzed_at DESC
            LIMIT 1
            """,
            (page_id,)
        )
        row = await cursor.fetchone()
        if not row:
            return None

        result = dict(row)
        result['keywords'] = json.loads(result['keywords'])
        return result

    async def delete_expired_insights(self) -> int:
        """Delete expired insights"""
        cursor = await self._connection.execute(
            "DELETE FROM insights WHERE expires_at <= datetime('now')"
        )
        await self._connection.commit()
        return cursor.rowcount

    # Helper methods
    def _row_to_page_dict(self, row: aiosqlite.Row) -> dict:
        """Convert row to page dict with insights"""
        page = {
            'id': row['id'],
            'url': row['url'],
            'title': row['title'],
            'domain': row['domain'],
            'visit_count': row['visit_count'],
            'last_visit_time': row['last_visit_time'],
            'first_visit_time': row['first_visit_time'],
            'created_at': row['created_at'],
            'updated_at': row['updated_at'],
            'has_insights': bool(row['has_insights'])
        }

        # Add insights if available
        if row['has_insights'] and row['insight_id']:
            page['insights'] = {
                'id': row['insight_id'],
                'page_id': row['id'],
                'summary': row['summary'],
                'type': row['type'],
                'keywords': json.loads(row['keywords']) if row['keywords'] else [],
                'agent_name': row['agent_name'],
                'analyzed_at': row['analyzed_at'],
                'expires_at': row['expires_at']
            }

        return page


# Global database instance
db = Database()


async def get_db() -> Database:
    """Get database instance (for dependency injection)"""
    return db
