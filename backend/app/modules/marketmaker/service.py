"""Market maker service — transparent quoting.

Tunable, PUBLIC parameters (move to config/governance later). The MM is a normal user account
(MM_EMAIL); it has no special ledger privileges.
"""

from __future__ import annotations

from decimal import Decimal

from sqlalchemy.ext.asyncio import AsyncSession

from app import tokenomics as tk
from app.core.logging import get_logger

log = get_logger(__name__)

MM_EMAIL = "mm@computecoin.local"
DEFAULT_REFERENCE_PRICE = Decimal("1.00")  # CC priced at 1.00 USDC as a neutral starting ref
SPREAD_BPS = Decimal("50")  # 0.50% total spread
ORDER_SIZE = Decimal("500")  # CC per side
PAIR = "CC/USDC"


async def compute_quote(session: AsyncSession, pair: str = PAIR) -> dict:
    """Return a transparent bid/ask around a mid (from book, else last trade, else reference)."""
    from app.modules.exchange import service as exchange_service

    book = await exchange_service.order_book(session, pair, depth=1)
    best_bid = book["bids"][0]["price"] if book["bids"] else None
    best_ask = book["asks"][0]["price"] if book["asks"] else None

    reference_used = "book"
    if best_bid is not None and best_ask is not None:
        mid = (best_bid + best_ask) / Decimal("2")
    else:
        last = await exchange_service.last_price(session, pair)
        if last is not None:
            mid = last
            reference_used = "last_trade"
        else:
            mid = DEFAULT_REFERENCE_PRICE
            reference_used = "reference"

    half_spread = (mid * SPREAD_BPS / Decimal("10000")) / Decimal("2")
    return {
        "pair": pair,
        "mid": tk.quantize(mid),
        "bid": tk.quantize(mid - half_spread),
        "ask": tk.quantize(mid + half_spread),
        "spread_bps": SPREAD_BPS,
        "order_size": ORDER_SIZE,
        "reference_used": reference_used,
    }


async def refresh_quotes(session: AsyncSession, pair: str = PAIR) -> dict:
    """Cancel the MM's resting orders and re-post a fresh bid/ask. Best-effort.

    Used at startup seeding and can be scheduled. Skips silently if the MM lacks funds.
    """
    from app.modules.exchange import service as exchange_service
    from app.modules.identity import service as identity_service
    from app.schemas.exchange import OrderCreate

    mm = await identity_service.get_by_email(session, MM_EMAIL)
    if mm is None:
        return {"posted": False, "reason": "mm user missing"}

    # Cancel existing MM orders on this pair.
    for o in await exchange_service.list_orders(session, mm.id, open_only=True):
        if o.pair == pair:
            await exchange_service.cancel_order(session, mm.id, o.id)

    q = await compute_quote(session, pair)
    posted = []
    base, quote = exchange_service.parse_pair(pair)

    # Post bid (buy) if MM has quote, and ask (sell) if MM has base.
    try:
        if await exchange_service.available(session, mm.id, quote) >= q["bid"] * ORDER_SIZE:
            o, _ = await exchange_service.place_order(
                session,
                mm.id,
                OrderCreate(pair=pair, side="buy", type="limit", price=q["bid"], qty=ORDER_SIZE),
            )
            posted.append(o.id)
    except Exception as exc:  # noqa: BLE001
        log.warning("MM bid post skipped: %s", exc)
    try:
        if await exchange_service.available(session, mm.id, base) >= ORDER_SIZE:
            o, _ = await exchange_service.place_order(
                session,
                mm.id,
                OrderCreate(pair=pair, side="sell", type="limit", price=q["ask"], qty=ORDER_SIZE),
            )
            posted.append(o.id)
    except Exception as exc:  # noqa: BLE001
        log.warning("MM ask post skipped: %s", exc)

    return {"posted": bool(posted), "order_ids": posted, "quote": q}
