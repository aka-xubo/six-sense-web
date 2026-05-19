"""Configuration settings for Six Sense Web"""
from pathlib import Path
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings"""

    # Application
    app_name: str = "Six Sense Web"
    app_version: str = "0.1.0"
    debug: bool = True

    # Server
    host: str = "127.0.0.1"
    port: int = 8000

    # Database
    data_dir: Path = Path.home() / ".six-sense"
    db_path: Path = data_dir / "web.db"

    # Chrome History
    chrome_history_path: Path = Path.home() / "Library/Application Support/Google/Chrome/Default/History"

    # CORS
    cors_origins: list[str] = ["http://localhost:5173", "http://127.0.0.1:5173"]

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()

# Ensure data directory exists
settings.data_dir.mkdir(parents=True, exist_ok=True)
