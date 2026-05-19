"""FastAPI application entry point"""
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.config import settings
from app.database import db
from app.routers import sync, pages, agents, analyze, blacklist


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    # Startup
    await db.connect()
    print(f"✓ Database connected: {settings.db_path}")

    yield

    # Shutdown
    await db.disconnect()
    print("✓ Database disconnected")


app = FastAPI(
    title=settings.app_name,
    version=settings.app_version,
    debug=settings.debug,
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(sync.router)
app.include_router(pages.router)
app.include_router(agents.router)
app.include_router(analyze.router)
app.include_router(blacklist.router)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    pages_count = await db.get_pages_count()

    return {
        "status": "ok",
        "app": settings.app_name,
        "version": settings.app_version,
        "database": {
            "connected": db._connection is not None,
            "path": str(settings.db_path),
            "pages_count": pages_count
        }
    }


@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "message": "Six Sense Web API",
        "docs": "/docs",
        "health": "/health"
    }
