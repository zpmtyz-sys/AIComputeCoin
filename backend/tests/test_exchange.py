"""Exchange matching engine + settlement tests."""

from decimal import Decimal

import pytest

from app import tokenomics as tk
from app.core.errors import InsufficientBalance
from app.modules.exchange import service as exchange
from app.modules.identity import service as identity
from app.modules.ledger import service as ledger
from app.schemas.exchange import OrderCreate


async def _fund(session, ref: str, asset: str, amount: str):
    await ledger.mint(
        session, dest_ref=ref, asset=asset, amount=Decimal(amount), kind="deposit", ref="seed"
    )


async def test_full_match_and_fees(session):
    seller = await identity.register(session, "seller@x.com", "password123", "user")
    buyer = await identity.register(session, "buyer@x.com", "password123", "user")
    await _fund(session, seller.account_ref, "CC", "1000")
    await _fund(session, buyer.account_ref, "USDC", "1000")

    # Maker sells 100 CC @ 1.0
    await exchange.place_order(
        session,
        seller.id,
        OrderCreate(
            pair="CC/USDC", side="sell", type="limit", price=Decimal("1.0"), qty=Decimal("100")
        ),
    )
    # Taker buys 100 CC @ 1.0 -> full fill at maker price
    order, trades = await exchange.place_order(
        session,
        buyer.id,
        OrderCreate(
            pair="CC/USDC", side="buy", type="limit", price=Decimal("1.0"), qty=Decimal("100")
        ),
    )

    assert order.status == "filled"
    assert len(trades) == 1

    # Balances
    assert await ledger.get_balance(session, buyer.account_ref, "CC") == Decimal("100")
    assert await ledger.get_balance(session, seller.account_ref, "CC") == Decimal("900")

    # notional = 100; buyer is taker (0.10%), seller is maker (0.02%)
    buyer_fee = tk.quantize(Decimal("100") * tk.TAKER_FEE)  # 0.10
    seller_fee = tk.quantize(Decimal("100") * tk.MAKER_FEE)  # 0.02
    assert (
        await ledger.get_balance(session, buyer.account_ref, "USDC")
        == Decimal("1000") - Decimal("100") - buyer_fee
    )
    assert (
        await ledger.get_balance(session, seller.account_ref, "USDC") == Decimal("100") - seller_fee
    )
    assert await ledger.get_balance(session, ledger.TREASURY, "USDC") == buyer_fee + seller_fee

    assert await ledger.conservation_ok(session, "CC")
    assert await ledger.conservation_ok(session, "USDC")


async def test_insufficient_balance_rejected(session):
    u = await identity.register(session, "poor@x.com", "password123", "user")
    with pytest.raises(InsufficientBalance):
        await exchange.place_order(
            session,
            u.id,
            OrderCreate(
                pair="CC/USDC", side="sell", type="limit", price=Decimal("1.0"), qty=Decimal("10")
            ),
        )


async def test_price_time_priority(session):
    # Two sellers rest at the same price; the earlier order fills first.
    s1 = await identity.register(session, "s1@x.com", "password123", "user")
    s2 = await identity.register(session, "s2@x.com", "password123", "user")
    buyer = await identity.register(session, "b@x.com", "password123", "user")
    await _fund(session, s1.account_ref, "CC", "100")
    await _fund(session, s2.account_ref, "CC", "100")
    await _fund(session, buyer.account_ref, "USDC", "1000")

    o1, _ = await exchange.place_order(
        session,
        s1.id,
        OrderCreate(
            pair="CC/USDC", side="sell", type="limit", price=Decimal("2.0"), qty=Decimal("50")
        ),
    )
    o2, _ = await exchange.place_order(
        session,
        s2.id,
        OrderCreate(
            pair="CC/USDC", side="sell", type="limit", price=Decimal("2.0"), qty=Decimal("50")
        ),
    )
    # Buyer takes 50 -> should hit o1 (older) fully, o2 untouched.
    _, trades = await exchange.place_order(
        session,
        buyer.id,
        OrderCreate(
            pair="CC/USDC", side="buy", type="limit", price=Decimal("2.0"), qty=Decimal("50")
        ),
    )
    assert len(trades) == 1
    assert trades[0].maker_order_id == o1.id
    assert await ledger.get_balance(session, buyer.account_ref, "CC") == Decimal("50")


async def test_best_price_priority(session):
    # Lower-priced ask should fill before a higher-priced one for a buy.
    s1 = await identity.register(session, "low@x.com", "password123", "user")
    s2 = await identity.register(session, "high@x.com", "password123", "user")
    buyer = await identity.register(session, "bb@x.com", "password123", "user")
    await _fund(session, s1.account_ref, "CC", "100")
    await _fund(session, s2.account_ref, "CC", "100")
    await _fund(session, buyer.account_ref, "USDC", "1000")

    await exchange.place_order(
        session,
        s2.id,
        OrderCreate(
            pair="CC/USDC", side="sell", type="limit", price=Decimal("3.0"), qty=Decimal("10")
        ),
    )
    low, _ = await exchange.place_order(
        session,
        s1.id,
        OrderCreate(
            pair="CC/USDC", side="sell", type="limit", price=Decimal("2.0"), qty=Decimal("10")
        ),
    )
    _, trades = await exchange.place_order(
        session,
        buyer.id,
        OrderCreate(
            pair="CC/USDC", side="buy", type="limit", price=Decimal("3.0"), qty=Decimal("10")
        ),
    )
    assert len(trades) == 1
    assert trades[0].maker_order_id == low.id
    assert trades[0].price == Decimal("2.0")


async def test_market_buy_against_book(session):
    seller = await identity.register(session, "ms@x.com", "password123", "user")
    buyer = await identity.register(session, "mb@x.com", "password123", "user")
    await _fund(session, seller.account_ref, "CC", "100")
    await _fund(session, buyer.account_ref, "USDC", "1000")

    await exchange.place_order(
        session,
        seller.id,
        OrderCreate(
            pair="CC/USDC", side="sell", type="limit", price=Decimal("2.0"), qty=Decimal("30")
        ),
    )
    order, trades = await exchange.place_order(
        session, buyer.id, OrderCreate(pair="CC/USDC", side="buy", type="market", qty=Decimal("30"))
    )
    assert order.status == "filled"
    assert await ledger.get_balance(session, buyer.account_ref, "CC") == Decimal("30")
    assert await ledger.conservation_ok(session, "USDC")
    assert await ledger.conservation_ok(session, "CC")
