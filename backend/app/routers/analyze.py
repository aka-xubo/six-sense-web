"""Analyze API routes - SSE streaming"""
from fastapi import APIRouter, HTTPException, Query
from sse_starlette.sse import EventSourceResponse
from datetime import datetime, timedelta
from app.models import InsightsCreate
from app.database import db
from app.services.agent_adapter import get_adapter
from app.services.page_fetcher import page_fetcher
from app.services.insights_parser import parse_insights
import json

router = APIRouter(prefix="/api", tags=["analyze"])


@router.get("/analyze")
async def analyze_page(
    page_id: int = Query(..., description="Page ID to analyze"),
    agent_name: str = Query(..., description="Agent name to use")
):
    """
    Analyze page content using AI agent (SSE streaming)

    Args:
        page_id: Page ID to analyze
        agent_name: Agent name to use

    Returns:
        Server-Sent Events stream
    """
    async def event_generator():
        try:
            # Get page info
            page = await db.get_page(page_id)
            if not page:
                yield {
                    "event": "analysis_error",
                    "data": json.dumps({"error": "Page not found"})
                }
                return

            # Check if insights already exist
            existing_insights = await db.get_insights(page_id)
            if existing_insights:
                yield {
                    "event": "complete",
                    "data": json.dumps(existing_insights)
                }
                return

            # Fetch page content
            yield {
                "event": "status",
                "data": json.dumps({"message": "Fetching page content..."})
            }

            content = await page_fetcher.fetch_content(page['url'])
            if not content:
                yield {
                    "event": "analysis_error",
                    "data": json.dumps({"error": "Failed to fetch page content"})
                }
                return

            # Get agent adapter
            try:
                adapter = get_adapter(agent_name)
            except ValueError as e:
                yield {
                    "event": "analysis_error",
                    "data": json.dumps({"error": str(e)})
                }
                return

            # Analyze page (stream output)
            yield {
                "event": "status",
                "data": json.dumps({"message": f"Analyzing with {agent_name}..."})
            }

            full_text = ""
            async for delta in adapter.analyze_page(page['url'], content):
                full_text += delta
                yield {
                    "event": "delta",
                    "data": json.dumps({"text": delta})
                }

            # Parse insights from full text
            insights_data = parse_insights(full_text)

            if not insights_data:
                yield {
                    "event": "analysis_error",
                    "data": json.dumps({"error": "Failed to parse insights from AI output"})
                }
                return

            # Save insights to database
            insights_create = InsightsCreate(
                page_id=page_id,
                summary=insights_data['summary'],
                type=insights_data['type'],
                keywords=insights_data['keywords'],
                agent_name=agent_name
            )

            insights_id = await db.create_insights(insights_create)

            # Get saved insights
            saved_insights = await db.get_insights(page_id)

            # Send complete event
            yield {
                "event": "complete",
                "data": json.dumps(saved_insights)
            }

        except Exception as e:
            yield {
                "event": "analysis_error",
                "data": json.dumps({"error": f"Analysis failed: {str(e)}"})
            }

    return EventSourceResponse(event_generator())
