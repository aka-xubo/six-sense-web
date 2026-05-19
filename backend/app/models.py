"""Pydantic models for request/response validation"""
from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class PageBase(BaseModel):
    """Base page model"""
    url: str
    title: Optional[str] = None
    domain: str
    visit_count: int = 0
    last_visit_time: datetime
    first_visit_time: Optional[datetime] = None


class PageCreate(PageBase):
    """Model for creating a page"""
    pass


class PageUpdate(BaseModel):
    """Model for updating a page"""
    title: Optional[str] = None
    visit_count: Optional[int] = None
    last_visit_time: Optional[datetime] = None


class InsightsBase(BaseModel):
    """Base insights model"""
    summary: str
    type: str
    keywords: list[str] = Field(default_factory=list, max_length=3)
    agent_name: str


class InsightsCreate(InsightsBase):
    """Model for creating insights"""
    page_id: int


class Insights(InsightsBase):
    """Insights response model"""
    id: int
    page_id: int
    analyzed_at: datetime
    expires_at: datetime

    class Config:
        from_attributes = True


class Page(PageBase):
    """Page response model"""
    id: int
    created_at: datetime
    updated_at: datetime
    has_insights: bool = False
    insights: Optional[Insights] = None

    class Config:
        from_attributes = True


class PageListResponse(BaseModel):
    """Page list response"""
    pages: list[Page]
    total: int
    has_more: bool


class PageDateGroup(BaseModel):
    """Pages grouped by a natural date"""
    date_key: str
    title: str
    pages: list[Page]


class PageGroupListResponse(BaseModel):
    """Page groups response"""
    groups: list[PageDateGroup]
    total: int
    has_more: bool
    next_cursor: Optional[str] = None


class SyncRequest(BaseModel):
    """Sync request model"""
    months: int = Field(default=2, ge=1, le=24)


class SyncResponse(BaseModel):
    """Sync response model"""
    status: str
    new_pages: int
    updated_pages: int
    total_pages: int
    sync_time: datetime


class AnalyzeRequest(BaseModel):
    """Analyze request model"""
    page_id: int
    agent_name: str


class BlacklistEntry(BaseModel):
    """Blacklist entry"""
    id: int
    type: str = "url"
    pattern: str
    created_at: datetime


class BlacklistCreate(BaseModel):
    """Create blacklist entry request"""
    type: str = "url"
    pattern: str


class BlacklistAddPageResponse(BaseModel):
    """Response after blacklisting a page"""
    entry: BlacklistEntry
    hidden_pages: int


class BlacklistListResponse(BaseModel):
    """Blacklist list response"""
    entries: list[BlacklistEntry]


class BlacklistDeleteResponse(BaseModel):
    """Response after deleting a blacklist entry"""
    hidden_pages: int


class AgentInfo(BaseModel):
    """Agent information model"""
    name: str
    display_name: str
    version: Optional[str] = None
    available: bool


class AgentsResponse(BaseModel):
    """Agents list response"""
    agents: list[AgentInfo]
