"""Async SQLAlchemy database setup.

- One engine, one async sessionmaker.
- `Base` is the declarative base for all ORM models in app/models/.
- `get_session` is the FastAPI dependency: one session (= one transaction) per request.
- `init_db` creates tables (MVP). Production uses Alembic migrations (see docs/workflow.md T3.1).

Money/quantity columns MUST use Numeric (Decimal), never Float. See CLAUDE.md invariant #6.
"""

from __future__ import annotations

from collections.abc import AsyncGenerator
from typing import Any

from sqlalchemy.ext.asyncio import (
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)
from sqlalchemy.orm import DeclarativeBase

from app.core.config import settings


class Base(DeclarativeBase):
    """Declarative base for all ORM models."""


def _engine_kwargs() -> dict[str, Any]:
    kwargs: dict[str, Any] = {"echo": settings.db_echo, "future": True}
    if settings.is_sqlite:
        # check_same_thread is irrelevant for aiosqlite, but pool settings differ.
        kwargs["connect_args"] = {}
    else:
        kwargs["pool_size"] = 10
        kwargs["max_overflow"] = 20
        kwargs["pool_pre_ping"] = True
    return kwargs


engine = create_async_engine(settings.database_url, **_engine_kwargs())

SessionLocal = async_sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,
    autoflush=False,
)


async def get_session() -> AsyncGenerator[AsyncSession, None]:
    """FastAPI dependency. Commits on success, rolls back on error."""
    async with SessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise


async def init_db() -> None:
    """Create all tables. Idempotent. MVP only — production uses Alembic."""
    # Import models so they are registered on Base.metadata before create_all.
    import app.models  # noqa: F401

    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
