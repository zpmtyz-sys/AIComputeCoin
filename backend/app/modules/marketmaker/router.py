"""Market maker HTTP routes (thin). Quotes are public for transparency."""

from __future__ import annotations

from fastapi import APIRouter, Query

from app.api.deps import SessionDep
from app.modules.marketmaker import service as mm_service

router = APIRouter(prefix="/mm", tags=["marketmaker"])


@router.get("/quotes")
async def quotes(session: SessionDep, pair: str = Query(default="CC/USDC")) -> dict:
    """Transparent market-maker quote (mid/bid/ask) and public parameters."""
    q = await mm_service.compute_quote(session, pair)
    return {
        "pair": q["pair"],
        "mid": format(q["mid"], "f"),
        "bid": format(q["bid"], "f"),
        "ask": format(q["ask"], "f"),
        "spread_bps": format(q["spread_bps"], "f"),
        "order_size": format(q["order_size"], "f"),
        "reference_used": q["reference_used"],
    }
