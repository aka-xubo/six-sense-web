"""Pages API routes"""
from typing import Optional
from fastapi import APIRouter, HTTPException, Query
from app.models import Page, PageDateGroup, PageGroupListResponse, PageListResponse
from app.database import db

router = APIRouter(prefix="/api", tags=["pages"])

WEEKDAYS = ["一", "二", "三", "四", "五", "六", "日"]


def format_group_title(date_key: str) -> str:
    """Format YYYY-MM-DD as a Chinese date group title."""
    from datetime import date, datetime, timedelta

    group_date = datetime.strptime(date_key, "%Y-%m-%d").date()
    today = date.today()
    full_date = f"{group_date.year}年{group_date.month}月{group_date.day}日星期{WEEKDAYS[group_date.weekday()]}"

    if group_date == today:
        return f"今天 - {full_date}"
    if group_date == today - timedelta(days=1):
        return f"昨天 - {full_date}"
    return full_date


@router.get("/pages", response_model=PageListResponse)
async def list_pages(
    q: Optional[str] = Query(None, description="Search query (title, domain, url)"),
    limit: int = Query(50, ge=1, le=200, description="Number of pages to return"),
    offset: int = Query(0, ge=0, description="Offset for pagination"),
    sort: str = Query("last_visit_time", description="Sort field (last_visit_time, visit_count, created_at)")
):
    """
    List pages with optional search and pagination

    Args:
        q: Search query (matches title, domain, url)
        limit: Number of pages to return (1-200)
        offset: Offset for pagination
        sort: Sort field

    Returns:
        List of pages with pagination info
    """
    try:
        pages, total = await db.list_pages(
            query=q,
            limit=limit,
            offset=offset,
            sort=sort
        )

        has_more = (offset + limit) < total

        return PageListResponse(
            pages=pages,
            total=total,
            has_more=has_more
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to list pages: {str(e)}")


@router.get("/page-groups", response_model=PageGroupListResponse)
async def list_page_groups(
    q: Optional[str] = Query(None, description="Search query (title, domain, url)"),
    cursor: Optional[str] = Query(None, description="Date cursor in YYYY-MM-DD format"),
    date_from: Optional[str] = Query(None, description="Start date in YYYY-MM-DD format"),
    date_to: Optional[str] = Query(None, description="End date in YYYY-MM-DD format"),
    limit: int = Query(1, ge=1, le=7, description="Number of date groups to return")
):
    """
    List pages grouped by date.

    Pagination is date-based rather than row-based. Each response returns whole
    date blocks so the UI never splits one date across multiple loads.
    """
    try:
        groups, total, has_more, next_cursor = await db.list_page_groups(
            query=q,
            cursor=cursor,
            limit=limit,
            date_from=date_from,
            date_to=date_to
        )

        return PageGroupListResponse(
            groups=[
                PageDateGroup(
                    date_key=group['date_key'],
                    title=format_group_title(group['date_key']),
                    pages=group['pages']
                )
                for group in groups
            ],
            total=total,
            has_more=has_more,
            next_cursor=next_cursor
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to list page groups: {str(e)}")


@router.get("/pages/{page_id}", response_model=Page)
async def get_page(page_id: int):
    """
    Get page by ID

    Args:
        page_id: Page ID

    Returns:
        Page with insights (if available)
    """
    try:
        page = await db.get_page(page_id)

        if not page:
            raise HTTPException(status_code=404, detail="Page not found")

        return page
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get page: {str(e)}")
