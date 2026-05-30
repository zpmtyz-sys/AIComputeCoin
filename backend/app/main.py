"""FastAPI application factory + lifespan + static dashboard mount.

Run (dev):   uvicorn app.main:app --reload --port 8000
Run (docker): see Dockerfile / docker-compose.yml (host port 8788).
"""

from __future__ import annotations

import asyncio
import contextlib
from pathlib import Path

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles

from app import __version__
from app.api.routes import api_router
from app.core.config import settings
from app.core.database import SessionLocal, init_db
from app.core.errors import register_exception_handlers
from app.core.logging import configure_logging, get_logger
from app.modules.ledger import service as ledger_service
from app.seed import bootstrap

log = get_logger(__name__)

WEB_DIR = Path(__file__).resolve().parents[2] / "web"


async def _emission_loop() -> None:
    """Periodically finalize emission for completed days. Idempotent."""
    while True:
        try:
            async with SessionLocal() as session:
                minted = await ledger_service.run_due_emissions(session)
                await session.commit()
                if minted and minted > 0:
                    log.info("Scheduler minted %s CC", minted)
        except asyncio.CancelledError:
            raise
        except Exception as exc:  # noqa: BLE001
            log.warning("emission loop error: %s", exc)
        await asyncio.sleep(settings.emission_tick_seconds)


@contextlib.asynccontextmanager
async def lifespan(app: FastAPI):
    configure_logging()
    log.info("Starting %s v%s (env=%s)", settings.app_name, __version__, settings.environment)
    await init_db()
    async with SessionLocal() as session:
        await bootstrap(session)
        await session.commit()

    task: asyncio.Task | None = None
    if settings.enable_emission_scheduler:
        task = asyncio.create_task(_emission_loop())

    try:
        yield
    finally:
        if task is not None:
            task.cancel()
            with contextlib.suppress(asyncio.CancelledError):
                await task


def create_app() -> FastAPI:
    app = FastAPI(
        title="ComputeCoin API",
        version=__version__,
        description=(
            "Compliant edition. Standardize delivered AI compute (CU) -> transparent token (CC) "
            "-> spot exchange. Integrates with sub2api as an external service. "
            "See repository docs/ for architecture, tokenomics, and compliance."
        ),
        lifespan=lifespan,
    )

    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.cors_origins_list,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    register_exception_handlers(app)

    @app.get("/health", tags=["meta"])
    async def health() -> dict:
        return {"status": "ok", "service": settings.app_name, "version": __version__}

    app.include_router(api_router, prefix="/api")

    # Static dashboard at root (added last so /api, /docs, /health win).
    if WEB_DIR.is_dir():
        app.mount("/", StaticFiles(directory=str(WEB_DIR), html=True), name="web")
    else:
        log.warning("web dir not found at %s; dashboard disabled", WEB_DIR)

    return app


app = create_app()
