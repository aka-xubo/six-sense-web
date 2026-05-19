"""Sync API routes"""
from fastapi import APIRouter, HTTPException
from app.models import SyncRequest, SyncResponse
from app.services.browser_sync import browser_sync_service

router = APIRouter(prefix="/api", tags=["sync"])


@router.post("/sync", response_model=SyncResponse)
async def sync_browser_history(request: SyncRequest):
    """
    Sync browser history to database

    Args:
        request: Sync request with months parameter

    Returns:
        Sync statistics
    """
    try:
        result = await browser_sync_service.sync_history(request.months)
        return SyncResponse(**result)
    except FileNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except RuntimeError as e:
        raise HTTPException(status_code=500, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Sync failed: {str(e)}")
