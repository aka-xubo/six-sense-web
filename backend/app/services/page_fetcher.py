"""Page content fetcher service"""
import asyncio
import aiohttp
from typing import Optional


class PageFetcher:
    """Fetch page content from URL"""

    def __init__(self, timeout: int = 10):
        self.timeout = aiohttp.ClientTimeout(total=timeout)

    async def fetch_content(self, url: str) -> Optional[str]:
        """
        Fetch page content from URL

        Args:
            url: Page URL

        Returns:
            Page content (HTML) or None if failed
        """
        try:
            async with aiohttp.ClientSession(timeout=self.timeout) as session:
                async with session.get(url) as response:
                    if response.status == 200:
                        content = await response.text()
                        # Limit content size to avoid overwhelming the AI
                        max_length = 50000  # ~50KB
                        if len(content) > max_length:
                            content = content[:max_length]
                        return content
                    else:
                        return None
        except Exception as e:
            print(f"Failed to fetch {url}: {e}")
            return None


# Global fetcher instance
page_fetcher = PageFetcher()
