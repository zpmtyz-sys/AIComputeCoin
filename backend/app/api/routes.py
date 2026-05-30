"""Aggregate all module routers under a single API router (mounted at /api)."""

from __future__ import annotations

from fastapi import APIRouter

from app.modules.compute.router import router as compute_router
from app.modules.exchange.router import router as exchange_router
from app.modules.identity.router import router as identity_router
from app.modules.ledger.router import router as ledger_router
from app.modules.marketmaker.router import router as mm_router
from app.modules.relay.router import router as relay_router

api_router = APIRouter()
api_router.include_router(identity_router)
api_router.include_router(relay_router)
api_router.include_router(compute_router)
api_router.include_router(ledger_router)
api_router.include_router(exchange_router)
api_router.include_router(mm_router)
