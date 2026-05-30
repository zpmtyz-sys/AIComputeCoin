"""Exchange HTTP routes (thin)."""

from __future__ import annotations

from fastapi import APIRouter, Depends, Query

from app.api.deps import CurrentUser, SessionDep, require_kyc
from app.modules.exchange import service as exchange_service
from app.schemas.exchange import (
    BookLevel,
    OrderBookOut,
    OrderCreate,
    OrderOut,
    TradeOut,
)

router = APIRouter(prefix="/exchange", tags=["exchange"])


@router.get("/orderbook", response_model=OrderBookOut)
async def orderbook(
    session: SessionDep,
    pair: str = Query(default="CC/USDC"),
    depth: int = Query(default=20, ge=1, le=100),
) -> OrderBookOut:
    book = await exchange_service.order_book(session, pair, depth)
    return OrderBookOut(
        pair=book["pair"],
        bids=[BookLevel(**lv) for lv in book["bids"]],
        asks=[BookLevel(**lv) for lv in book["asks"]],
    )


@router.get("/trades", response_model=list[TradeOut])
async def trades(
    session: SessionDep,
    pair: str = Query(default="CC/USDC"),
    limit: int = Query(default=50, ge=1, le=200),
) -> list[TradeOut]:
    rows = await exchange_service.recent_trades(session, pair, limit)
    return [TradeOut.model_validate(t) for t in rows]


@router.post("/orders", response_model=OrderOut, status_code=201)
async def place_order(
    body: OrderCreate,
    user: CurrentUser,
    session: SessionDep,
    _kyc=Depends(require_kyc),  # compliance hook (no-op unless REQUIRE_KYC_FOR_TRADING=true)
) -> OrderOut:
    order, _ = await exchange_service.place_order(session, user.id, body)
    return OrderOut.model_validate(order)


@router.get("/orders", response_model=list[OrderOut])
async def my_orders(
    user: CurrentUser,
    session: SessionDep,
    open_only: bool = Query(default=False),
) -> list[OrderOut]:
    rows = await exchange_service.list_orders(session, user.id, open_only=open_only)
    return [OrderOut.model_validate(o) for o in rows]


@router.delete("/orders/{order_id}", response_model=OrderOut)
async def cancel_order(order_id: int, user: CurrentUser, session: SessionDep) -> OrderOut:
    order = await exchange_service.cancel_order(session, user.id, order_id)
    return OrderOut.model_validate(order)
