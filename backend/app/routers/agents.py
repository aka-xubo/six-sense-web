"""Agents API routes"""
import asyncio
from fastapi import APIRouter, HTTPException
from app.models import AgentsResponse
from app.services.agent_detector import agent_detector

router = APIRouter(prefix="/api", tags=["agents"])


@router.get("/agents", response_model=AgentsResponse)
async def get_available_agents():
    """
    Get list of available AI agents

    Returns:
        List of agents with availability and version info
    """
    try:
        agents = await asyncio.to_thread(agent_detector.detect_agents)
        return AgentsResponse(agents=agents)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to detect agents: {str(e)}")
