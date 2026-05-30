"""Application configuration.

All runtime knobs live here (or in tokenomics.py for protocol/economic constants).
Loaded from environment variables / .env. See .env.example.
"""

from __future__ import annotations

from functools import lru_cache

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
        case_sensitive=False,
    )

    # ---- App ----
    app_name: str = "ComputeCoin"
    environment: str = Field(default="development")  # development | staging | production
    debug: bool = True

    # Container listen address. Host port mapping is handled by docker-compose (.env APP_HOST_PORT).
    host: str = "0.0.0.0"
    port: int = 8000

    # ---- Security / JWT ----
    secret_key: str = Field(default="dev-insecure-change-me", description="JWT signing key")
    jwt_algorithm: str = "HS256"
    access_token_expire_minutes: int = 60 * 24  # 24h

    # ---- Database ----
    # Default is SQLite for zero-setup local dev & tests; docker-compose overrides with Postgres.
    database_url: str = "sqlite+aiosqlite:///./computecoin.db"
    db_echo: bool = False

    # ---- Redis ----
    redis_url: str = "redis://localhost:6379/0"

    # ---- Bootstrap admin (created on first startup if not present) ----
    admin_email: str | None = None
    admin_password: str | None = None

    # ---- sub2api integration (optional; falls back to local mock if unset) ----
    sub2api_base_url: str | None = None  # e.g. http://host.docker.internal:8080
    sub2api_admin_token: str | None = None

    # ---- CORS ----
    cors_allow_origins: str = "*"  # comma-separated, or "*"

    # ---- Feature flags / behaviour ----
    enable_emission_scheduler: bool = True
    emission_tick_seconds: int = 60 * 60  # how often the scheduler checks for due emission
    require_kyc_for_trading: bool = False  # compliance hook; off in MVP

    @property
    def cors_origins_list(self) -> list[str]:
        if self.cors_allow_origins.strip() == "*":
            return ["*"]
        return [o.strip() for o in self.cors_allow_origins.split(",") if o.strip()]

    @property
    def is_sqlite(self) -> bool:
        return self.database_url.startswith("sqlite")


@lru_cache
def get_settings() -> Settings:
    return Settings()


settings = get_settings()
