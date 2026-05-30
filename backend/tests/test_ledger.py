"""Ledger invariants: double-entry conservation, supply cap, emission distribution."""

from decimal import Decimal

import pytest

from app import tokenomics as tk
from app.core.errors import InsufficientBalance, SupplyCapExceeded
from app.models.relay import CapacitySource, DeliveryRecord
from app.modules.identity import service as identity
from app.modules.ledger import service as ledger


async def test_mint_and_conservation(session):
    await ledger.mint(
        session, dest_ref="user:1", asset="CC", amount=Decimal("100"), kind="emission", ref="t"
    )
    assert await ledger.get_balance(session, "user:1", "CC") == Decimal("100")
    assert await ledger.current_minted(session, "CC") == Decimal("100")
    # mint_source went negative by the same amount => total balances == 0
    assert await ledger.get_balance(session, ledger.MINT_SOURCE, "CC") == Decimal("-100")
    assert await ledger.conservation_ok(session, "CC")


async def test_supply_cap_enforced(session):
    with pytest.raises(SupplyCapExceeded):
        await ledger.mint(
            session,
            dest_ref="user:1",
            asset="CC",
            amount=tk.MAX_SUPPLY + Decimal("1"),
            kind="emission",
            ref="over",
        )


async def test_transfer_insufficient(session):
    with pytest.raises(InsufficientBalance):
        await ledger.transfer(
            session, from_ref="user:1", to_ref="user:2", asset="CC", amount=Decimal("5")
        )


async def test_transfer_and_conservation(session):
    await ledger.mint(
        session, dest_ref="user:1", asset="USDC", amount=Decimal("50"), kind="deposit", ref="d"
    )
    await ledger.transfer(
        session, from_ref="user:1", to_ref="user:2", asset="USDC", amount=Decimal("20")
    )
    assert await ledger.get_balance(session, "user:1", "USDC") == Decimal("30")
    assert await ledger.get_balance(session, "user:2", "USDC") == Decimal("20")
    assert await ledger.conservation_ok(session, "USDC")


async def test_seed_genesis_once(session):
    assert await ledger.seed_genesis(session) is True
    assert await ledger.seed_genesis(session) is False  # idempotent
    minted = await ledger.current_minted(session, "CC")
    # equals sum of genesis bucket amounts
    expected = sum((tk.genesis_amount(b) for b in tk.GENESIS_ALLOCATIONS), Decimal("0"))
    assert minted == expected
    assert await ledger.conservation_ok(session, "CC")
    # a known bucket has a balance
    team = await ledger.get_balance(session, "genesis:team", "CC")
    assert team == tk.genesis_amount(next(b for b in tk.GENESIS_ALLOCATIONS if b.name == "team"))


async def _make_supplier_with_cu(session, email: str, cu: Decimal, day: int):
    user = await identity.register(session, email, "password123", "supplier")
    src = CapacitySource(
        user_id=user.id,
        kind="gpu",
        name="node",
        gpu_model="H100",
        source_attestation="I own this hardware legitimately.",
        verified=True,
    )
    session.add(src)
    await session.flush()
    rec = DeliveryRecord(
        source_id=src.id,
        gpu_model="H100",
        gpu_hours=Decimal("0"),
        delivered_tokens=Decimal("0"),
        cu=cu,
        verified=True,
        period=day,
    )
    session.add(rec)
    await session.flush()
    return user


async def test_emission_distributes_by_cu(session):
    await ledger.seed_genesis(session)
    day = 5
    a = await _make_supplier_with_cu(session, "a@x.com", Decimal("300"), day)
    b = await _make_supplier_with_cu(session, "b@x.com", Decimal("100"), day)

    minted = await ledger.run_emission(session, day)
    budget = tk.daily_emission(tk.halving_index(day))
    assert minted > 0
    # A has 3x the CU of B -> ~3x the reward
    bal_a = await ledger.get_balance(session, a.account_ref, "CC")
    bal_b = await ledger.get_balance(session, b.account_ref, "CC")
    assert bal_a == tk.quantize(budget * Decimal("300") / Decimal("400"))
    assert bal_b == tk.quantize(budget * Decimal("100") / Decimal("400"))
    assert await ledger.conservation_ok(session, "CC")

    # idempotent: second run for same day mints nothing
    assert await ledger.run_emission(session, day) == Decimal("0")


async def test_emission_no_cu_mints_zero(session):
    await ledger.seed_genesis(session)
    assert await ledger.run_emission(session, 9) == Decimal("0")
