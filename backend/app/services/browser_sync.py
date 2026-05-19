"""Browser history sync service - reuses logic from six_sense.py"""
import os
import sqlite3
import shutil
from datetime import datetime, timedelta
from pathlib import Path
from urllib.parse import urlparse
from typing import Optional
from app.config import settings
from app.database import db
from app.models import PageCreate, PageUpdate


class BrowserSyncService:
    """Service for syncing browser history to database"""

    def __init__(self):
        self.chrome_history = settings.chrome_history_path
        self.temp_history = Path("/tmp/History-six-sense-web")

    async def sync_history(self, months: int = 2) -> dict:
        """
        Sync browser history to database

        Args:
            months: Number of months to sync (1-24)

        Returns:
            dict with sync statistics
        """
        start_time = datetime.now()

        # Query browser history
        results = self._query_browser_history(months)

        # Process and save to database
        new_pages = 0
        updated_pages = 0

        for url, title, visit_count, last_visit_time in results:
            # Filter invalid pages
            if self._should_filter(url) or await db.is_url_blacklisted(url):
                continue

            # Extract domain
            domain = self._extract_domain(url)

            # Create page object
            page = PageCreate(
                url=url,
                title=title or url,
                domain=domain,
                visit_count=visit_count,
                last_visit_time=datetime.fromisoformat(last_visit_time)
            )

            # Check if page exists
            existing = await db.get_page_by_url(url)

            if existing:
                # Update existing page
                await db.update_page(
                    existing['id'],
                    PageUpdate(
                        title=page.title,
                        visit_count=page.visit_count,
                        last_visit_time=page.last_visit_time
                    )
                )
                updated_pages += 1
            else:
                # Create new page
                await db.create_page(page)
                new_pages += 1

        total_pages = await db.get_pages_count()

        return {
            'status': 'success',
            'new_pages': new_pages,
            'updated_pages': updated_pages,
            'total_pages': total_pages,
            'sync_time': start_time
        }

    def _query_browser_history(self, months: int) -> list:
        """
        Query Chrome browser history (copied from six_sense.py)

        Returns:
            list of (url, title, visit_count, last_visit_time)
        """
        # Check if Chrome history exists
        if not self.chrome_history.exists():
            raise FileNotFoundError(f"Chrome history not found at {self.chrome_history}")

        # Copy database to avoid lock issues
        try:
            shutil.copy2(self.chrome_history, self.temp_history)
        except Exception as e:
            raise RuntimeError(f"Error copying Chrome history: {e}")

        # Query the copy
        conn = sqlite3.connect(str(self.temp_history))
        cursor = conn.cursor()

        try:
            # Query last N months
            query = """
                SELECT url, title, visit_count,
                       datetime(last_visit_time/1000000-11644473600, 'unixepoch') as last_visit_time
                FROM urls
                WHERE last_visit_time > (strftime('%s', 'now', '-{} months') + 11644473600) * 1000000
                ORDER BY visit_count DESC
            """.format(months)

            cursor.execute(query)
            results = cursor.fetchall()
        finally:
            conn.close()
            # Clean up temp file
            if self.temp_history.exists():
                self.temp_history.unlink()

        return results

    def _extract_domain(self, url: str) -> str:
        """Extract domain from URL"""
        try:
            parsed = urlparse(url)
            return parsed.netloc or parsed.path.split('/')[0]
        except:
            return ""

    def _should_filter(self, url: str) -> bool:
        """
        Check if URL should be filtered out
        (copied from six_sense.py)
        """
        if self._is_homepage(url):
            return True
        if self._is_local_file(url):
            return True
        if self._is_ip_address(url):
            return True
        return False

    def _is_homepage(self, url: str) -> bool:
        """Check if URL is a homepage (path is empty or just '/')"""
        try:
            parsed = urlparse(url)
            path = parsed.path.strip()
            return path == '' or path == '/'
        except:
            return False

    def _is_local_file(self, url: str) -> bool:
        """Check if URL is a local file (file://)"""
        try:
            parsed = urlparse(url)
            return parsed.scheme == 'file'
        except:
            return False

    def _is_ip_address(self, url: str) -> bool:
        """Check if URL uses IP address instead of domain"""
        try:
            parsed = urlparse(url)
            domain = parsed.netloc.split(':')[0]  # Remove port if present

            # Check if it's an IPv4 address
            import re
            ipv4_pattern = r'^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$'
            if re.match(ipv4_pattern, domain):
                return True

            # Check if it's an IPv6 address
            if ':' in domain and '[' in parsed.netloc:
                return True

            # Check for localhost
            if domain in ['localhost', '127.0.0.1', '::1']:
                return True

            return False
        except:
            return False


# Global service instance
browser_sync_service = BrowserSyncService()
