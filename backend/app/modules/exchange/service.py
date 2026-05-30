"""Exchange service: price-time-priority matching engine + ledger settlement.

Matching rules (CLAUDE.md invariant #5):
  * Resting orders are matched best-price first, then oldest first (time priority).
  * Trades execute at the resting (maker) price.
  * After each fill, asset deltas net to zero except explicit fees, which go to `treasury`.

Funds for resting limit orders are RESERVED implicitly: available = ledger_balance - locked,
where `locked` is derived from the user's own open orders. No separate escrow entries.

This is a correct, readable MVP engine (single-process). Concurrency hardening (row locks /
Rust engine / WS) is backlog (docs/workflow.md T4.1/T4.2/T4.4).
"""

from __future__ import annotations

from collections.abc import Sequence
from decimal import Decimal

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app import tokenomics as tk
from app.core.errors import BadRequest, Forbidden, InsufficientBalance, NotFound
from app.models.exchange import Order, Trade
from app.schemas.exchange import OrderCreate

ZERO = Decimal("0")
OPEN_STATES = ("open", "partial")


def parse_pair(pair: str) -> tuple[str, str]:
    if "/" not in pair:
        raise BadRequest(f"invalid pair '{pair}'")
    base, quote = pair.split("/", 1)
    if not base or not quote:
        raise BadRequest(f"invalid pair '{pair}'")
    return base, quote


async def _open_orders_for_user(session: AsyncSession, user_id: int) -> Sequence[Order]:
    stmt = select(Order).where(Order.user_id == user_id, Order.status.in_(OPEN_STATES))
    return (await session.execute(stmt)).scalars().all()


async def locked_amount(session: AsyncSession, user_id: int, asset: str) -> Decimal:
    """Funds reserved by the user's own resting orders for a given asset."""
    locked = ZERO
    for o in await _open_orders_for_user(session, user_id):
        base, quote = parse_pair(o.pair)
        if o.side == "sell" and base == asset:
            locked += o.remaining
        elif o.side == "buy" and quote == asset and o.price is not None:
            locked += o.remaining * o.price
    return tk.quantize(locked)


async def available(session: AsyncSession, user_id: int, asset: str) -> Decimal:
    from app.modules.ledger import service as ledger_service

    bal = await ledger_service.get_balance(session, f"user:{user_id}", asset)
    return bal - await locked_amount(session, user_id, asset)


async def _resting_opposite(session: AsyncSession, taker: Order) -> list[Order]:
    """Resting orders on the opposite side, sorted by price-time priority."""
    if taker.side == "buy":
        stmt = select(Order).where(
            Order.pair == taker.pair,
            Order.side == "sell",
            Order.status.in_(OPEN_STATES),
        )
        if taker.type == "limit":
            stmt = stmt.where(Order.price <= taker.price)
        stmt = stmt.order_by(Order.price.asc(), Order.id.asc())
    else:
        stmt = select(Order).where(
            Order.pair == taker.pair,
            Order.side == "buy",
            Order.status.in_(OPEN_STATES),
        )
        if taker.type == "limit":
            stmt = stmt.where(Order.price >= taker.price)
        stmt = stmt.order_by(Order.price.desc(), Order.id.asc())
    return list((await session.execute(stmt)).scalars().all())


def _status_after(order: Order) -> str:
    if order.filled >= order.qty:
        return "filled"
    if order.filled > 0:
        return "partial"
    return "open"


async def place_order(
    session: AsyncSession, user_id: int, data: OrderCreate
) -> tuple[Order, list[Trade]]:
    base, quote = parse_pair(data.pair)
    qty = tk.quantize(data.qty)
    price = tk.quantize(data.price) if data.price is not None else None

    # ---- pre-trade balance checks ----
    if data.side == "sell":
        avail = await available(session, user_id, base)
        if avail < qty:
            raise InsufficientBalance(f"need {qty} {base}, available {avail}")
    else:  # buy
        avail_quote = await available(session, user_id, quote)
        if data.type == "limit":
            needed = tk.quantize(qty * price * (Decimal("1") + tk.TAKER_FEE))
            if avail_quote < needed:
                raise InsufficientBalance(f"need {needed} {quote}, available {avail_quote}")
        else:  # market buy
            if avail_quote <= 0:
                raise InsufficientBalance(f"no {quote} available")

    order = Order(
        user_id=user_id,
        pair=data.pair,
        side=data.side,
        type=data.type,
        price=price,
        qty=qty,
        filled=ZERO,
        status="open",
    )
    session.add(order)
    await session.flush()

    trades = await _match(session, order, base, quote)

    order.status = _status_after(order)
    # Market orders never rest: cancel any unfilled remainder.
    if data.type == "market" and order.remaining > 0:
        order.status = "filled" if order.filled >= order.qty else "cancelled"
    await session.flush()
    return order, trades


async def _match(session: AsyncSession, taker: Order, base: str, quote: str) -> list[Trade]:
    from app.modules.ledger import service as ledger_service

    trades: list[Trade] = []
    makers = await _resting_opposite(session, taker)

    for maker in makers:
        if taker.remaining <= 0:
            break
        if maker.user_id == taker.user_id:
            continue  # basic self-trade prevention
        trade_price = maker.price  # execute at resting price
        if trade_price is None or trade_price <= 0:
            continue

        trade_qty = min(taker.remaining, maker.remaining)

        # Determine buyer/seller.
        if taker.side == "buy":
            buyer_id, seller_id = taker.user_id, maker.user_id
            buyer_is_taker = True
        else:
            buyer_id, seller_id = maker.user_id, taker.user_id
            buyer_is_taker = False

        # For a market buy, cap trade_qty by how much quote the buyer can still afford.
        if taker.type == "market" and taker.side == "buy":
            avail_quote = await available(session, buyer_id, quote)
            max_notional = avail_quote / (Decimal("1") + tk.TAKER_FEE)
            affordable_qty = tk.quantize(max_notional / trade_price)
            trade_qty = min(trade_qty, affordable_qty)
            if trade_qty <= 0:
                break

        trade_qty = tk.quantize(trade_qty)
        notional = tk.quantize(trade_qty * trade_price)
        if notional <= 0:
            continue

        taker_fee = tk.quantize(notional * tk.TAKER_FEE)
        maker_fee = tk.quantize(notional * tk.MAKER_FEE)
        buyer_fee = taker_fee if buyer_is_taker else maker_fee
        seller_fee = maker_fee if buyer_is_taker else taker_fee

        buyer_ref = f"user:{buyer_id}"
        seller_ref = f"user:{seller_id}"
        ref = f"trade:{taker.id}:{maker.id}"

        # Settlement: base from seller->buyer, quote from buyer->seller, fees->treasury.
        await ledger_service.transfer(
            session,
            from_ref=seller_ref,
            to_ref=buyer_ref,
            asset=base,
            amount=trade_qty,
            kind="trade",
            ref=ref,
        )
        await ledger_service.transfer(
            session,
            from_ref=buyer_ref,
            to_ref=seller_ref,
            asset=quote,
            amount=notional,
            kind="trade",
            ref=ref,
        )
        if buyer_fee > 0:
            await ledger_service.transfer(
                session,
                from_ref=buyer_ref,
                to_ref=ledger_service.TREASURY,
                asset=quote,
                amount=buyer_fee,
                kind="fee",
                ref=ref,
            )
        if seller_fee > 0:
            await ledger_service.transfer(
                session,
                from_ref=seller_ref,
                to_ref=ledger_service.TREASURY,
                asset=quote,
                amount=seller_fee,
                kind="fee",
                ref=ref,
            )

        # Update both orders.
        taker.filled = tk.quantize(taker.filled + trade_qty)
        maker.filled = tk.quantize(maker.filled + trade_qty)
        maker.status = _status_after(maker)

        trade = Trade(
            pair=taker.pair,
            price=trade_price,
            qty=trade_qty,
            taker_order_id=taker.id,
            maker_order_id=maker.id,
            taker_user_id=taker.user_id,
            maker_user_id=maker.user_id,
            taker_fee=taker_fee,
            maker_fee=maker_fee,
        )
        session.add(trade)
        trades.append(trade)

    await session.flush()
    return trades


async def cancel_order(session: AsyncSession, user_id: int, order_id: int) -> Order:
    order = (await session.execute(select(Order).where(Order.id == order_id))).scalar_one_or_none()
    if order is None:
        raise NotFound("order not found")
    if order.user_id != user_id:
        raise Forbidden("not your order")
    if order.status in OPEN_STATES:
        order.status = "cancelled"
        await session.flush()
    return order


async def list_orders(
    session: AsyncSession, user_id: int, open_only: bool = False, limit: int = 100
) -> Sequence[Order]:
    stmt = select(Order).where(Order.user_id == user_id)
    if open_only:
        stmt = stmt.where(Order.status.in_(OPEN_STATES))
    stmt = stmt.order_by(Order.id.desc()).limit(limit)
    return (await session.execute(stmt)).scalars().all()


async def order_book(session: AsyncSession, pair: str, depth: int = 20) -> dict:
    stmt = select(Order).where(Order.pair == pair, Order.status.in_(OPEN_STATES))
    orders = (await session.execute(stmt)).scalars().all()

    bids: dict[Decimal, Decimal] = {}
    asks: dict[Decimal, Decimal] = {}
    for o in orders:
        if o.price is None:
            continue
        book = bids if o.side == "buy" else asks
        book[o.price] = book.get(o.price, ZERO) + o.remaining

    bid_levels = [
        {"price": p, "qty": q} for p, q in sorted(bids.items(), key=lambda x: x[0], reverse=True)
    ][:depth]
    ask_levels = [{"price": p, "qty": q} for p, q in sorted(asks.items(), key=lambda x: x[0])][
        :depth
    ]
    return {"pair": pair, "bids": bid_levels, "asks": ask_levels}


async def recent_trades(session: AsyncSession, pair: str, limit: int = 50) -> Sequence[Trade]:
    stmt = select(Trade).where(Trade.pair == pair).order_by(Trade.id.desc()).limit(limit)
    return (await session.execute(stmt)).scalars().all()


async def last_price(session: AsyncSession, pair: str) -> Decimal | None:
    stmt = select(Trade.price).where(Trade.pair == pair).order_by(Trade.id.desc()).limit(1)
    return (await session.execute(stmt)).scalars().first()
