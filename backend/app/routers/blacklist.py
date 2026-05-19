"""Blacklist API routes"""
from typing import Optional
from fastapi import APIRouter, HTTPException
from app.database import db
from app.models import BlacklistAddPageResponse, BlacklistCreate, BlacklistDeleteResponse, BlacklistEntry, BlacklistListResponse

router = APIRouter(prefix="/api", tags=["blacklist"])


@router.get("/blacklist", response_model=BlacklistListResponse)
async def list_blacklist_entries():
    """List blacklist entries."""
    entries = await db.list_blacklist_entries()
    return BlacklistListResponse(entries=[BlacklistEntry(**entry) for entry in entries])


@router.post("/blacklist", response_model=BlacklistEntry)
async def create_blacklist_entry(request: BlacklistCreate):
    """Create a typed blacklist entry."""
    pattern = request.pattern.strip()
    if not pattern:
        raise HTTPException(status_code=400, detail="Blacklist pattern is required")

    try:
        entry = await db.create_blacklist_entry(pattern, request.type)
        return BlacklistEntry(**entry)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/blacklist/pages/{page_id}", response_model=BlacklistAddPageResponse)
async def blacklist_page(page_id: int, type: str = "url", pattern: Optional[str] = None):
    """Blacklist a page URL and hide matched historical pages."""
    page = await db.get_page(page_id)
    if not page:
        raise HTTPException(status_code=404, detail="Page not found")

    blacklist_pattern = pattern.strip() if pattern else page['url']
    if not blacklist_pattern:
        raise HTTPException(status_code=400, detail="Blacklist pattern is required")

    try:
        entry = await db.create_blacklist_entry(blacklist_pattern, type)
        hidden_pages = await db.apply_blacklist_entry(type, blacklist_pattern)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    return BlacklistAddPageResponse(
        entry=BlacklistEntry(**entry),
        hidden_pages=hidden_pages
    )


@router.delete("/blacklist/{entry_id}", response_model=BlacklistDeleteResponse)
async def delete_blacklist_entry(entry_id: int):
    """Delete a blacklist entry and refresh hidden flags."""
    deleted = await db.delete_blacklist_entry(entry_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Blacklist entry not found")

    hidden_pages = await db.rebuild_blacklist()
    return BlacklistDeleteResponse(hidden_pages=hidden_pages)
