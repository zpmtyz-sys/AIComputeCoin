"""Ledger service — the only place that mutates balances.

Invariants enforced here (CLAUDE.md):
  #1 double-entry: every _apply debits and credits equal amounts -> sum of all balances == 0.
  #2 supply cap: mint() checks total minted CC never exceeds MAX_SUPPLY.
  #3 deterministic emission: run_emission uses tokenomics.daily_emission (pure).
  #4 transparent genesis: seed_genesis uses tokenomics.GENESIS_ALLOCATIONS only.
  #6 Decimal everywhere.

Money flows out of existence via the system account `mint_source` (it goes negative).
"""

from __future__ import annotations

from datetime import UTC, datetime
from decimal import Decimal

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app import tokenomics as tk
from app.core.errors import BadRequest, InsufficientBalance, SupplyCapExceeded
from app.core.logging import get_logger
from app.models.ledger import Account, Balance, EmissionEpoch, LedgerEntry

log = get_logger(__name__)

MINT_SOURCE = "mint_source"
TREASURY = "treasury"
ZERO = Decimal("0")


# ---------------------------------------------------------------------------
# Accounts & balances
# ---------------------------------------------------------------------------
async def get_or_create_account(
    session: AsyncSession, ref: str, *, owner_type: str = "system", kind: str = "wallet"
) -> Account:
    acc = (await session.execute(select(Account).where(Account.ref == ref))).scalar_one_or_none()
    if acc is None:
        acc = Account(ref=ref, owner_type=owner_type, kind=kind)
        session.add(acc)
        await session.flush()
    return acc


async def _balance_row(session: AsyncSession, account_id: int, asset: str) -> Balance | None:
    stmt = select(Balance).where(Balance.account_id == account_id, Balance.asset == asset)
    return (await session.execute(stmt)).scalar_one_or_none()


async def _adjust_balance(
    session: AsyncSession, account_id: int, asset: str, delta: Decimal
) -> None:
    row = await _balance_row(session, account_id, asset)
    if row is None:
        row = Balance(account_id=account_id, asset=asset, amount=ZERO)
        session.add(row)
        await session.flush()
    row.amount = row.amount + delta


async def get_balance(session: AsyncSession, ref: str, asset: str) -> Decimal:
    acc = (await session.execute(select(Account).where(Account.ref == ref))).scalar_one_or_none()
    if acc is None:
        return ZERO
    row = await _balance_row(session, acc.id, asset)
    return row.amount if row else ZERO


async def balances_for(session: AsyncSession, ref: str) -> dict[str, Decimal]:
    acc = (await session.execute(select(Account).where(Account.ref == ref))).scalar_one_or_none()
    if acc is None:
        return {}
    rows = (await session.execute(select(Balance).where(Balance.account_id == acc.id))).scalars()
    return {r.asset: r.amount for r in rows if r.amount != ZERO}


# ---------------------------------------------------------------------------
# Core primitive: post a double-entry
# ---------------------------------------------------------------------------
async def _apply(
    session: AsyncSession,
    *,
    debit_ref: str,
    credit_ref: str,
    asset: str,
    amount: Decimal,
    kind: str,
    ref: str,
    memo: str | None = None,
) -> LedgerEntry:
    amount = tk.quantize(amount)
    if amount <= 0:
        raise BadRequest("ledger amount must be > 0")
    debit = await get_or_create_account(session, debit_ref)
    credit = await get_or_create_account(session, credit_ref)
    entry = LedgerEntry(
        ref=ref,
        asset=asset,
        debit_account_id=debit.id,
        credit_account_id=credit.id,
        amount=amount,
        kind=kind,
        memo=memo,
    )
    session.add(entry)
    await _adjust_balance(session, debit.id, asset, -amount)
    await _adjust_balance(session, credit.id, asset, amount)
    await session.flush()
    return entry


async def current_minted(session: AsyncSession, asset: str) -> Decimal:
    """Total minted = -balance(mint_source, asset)."""
    return -(await get_balance(session, MINT_SOURCE, asset))


async def mint(
    session: AsyncSession,
    *,
    dest_ref: str,
    asset: str,
    amount: Decimal,
    kind: str,
    ref: str,
    memo: str | None = None,
) -> LedgerEntry:
    amount = tk.quantize(amount)
    if asset == tk.SYMBOL:
        minted = await current_minted(session, asset)
        if minted + amount > tk.MAX_SUPPLY:
            raise SupplyCapExceeded(
                f"mint of {amount} would exceed MAX_SUPPLY "
                f"(minted={minted}, cap={tk.MAX_SUPPLY})"
            )
    return await _apply(
        session,
        debit_ref=MINT_SOURCE,
        credit_ref=dest_ref,
        asset=asset,
        amount=amount,
        kind=kind,
        ref=ref,
        memo=memo,
    )


async def transfer(
    session: AsyncSession,
    *,
    from_ref: str,
    to_ref: str,
    asset: str,
    amount: Decimal,
    kind: str = "transfer",
    ref: str | None = None,
    memo: str | None = None,
    allow_negative: bool = False,
) -> LedgerEntry:
    amount = tk.quantize(amount)
    if not allow_negative:
        bal = await get_balance(session, from_ref, asset)
        if bal < amount:
            raise InsufficientBalance(f"{from_ref} has {bal} {asset}, needs {amount}")
    return await _apply(
        session,
        debit_ref=from_ref,
        credit_ref=to_ref,
        asset=asset,
        amount=amount,
        kind=kind,
        ref=ref or f"{kind}:{from_ref}->{to_ref}",
        memo=memo,
    )


# ---------------------------------------------------------------------------
# Genesis & emission
# ---------------------------------------------------------------------------
async def seed_genesis(session: AsyncSession) -> bool:
    """Mint the transparent genesis allocation once. Returns True if it ran."""
    existing = (
        await session.execute(select(LedgerEntry).where(LedgerEntry.kind == "genesis").limit(1))
    ).scalar_one_or_none()
    if existing is not None:
        return False
    for bucket in tk.GENESIS_ALLOCATIONS:
        amt = tk.genesis_amount(bucket)
        await mint(
            session,
            dest_ref=f"genesis:{bucket.name}",
            asset=tk.SYMBOL,
            amount=amt,
            kind="genesis",
            ref=f"genesis:{bucket.name}",
            memo=bucket.note,
        )
    log.info(
        "Genesis seeded: %s CC across %d buckets", tk.GENESIS_SUPPLY, len(tk.GENESIS_ALLOCATIONS)
    )
    return True


async def genesis_time(session: AsyncSession) -> datetime | None:
    acc = (
        await session.execute(select(Account).where(Account.ref == MINT_SOURCE))
    ).scalar_one_or_none()
    if acc is None:
        return None
    ts = acc.created_at
    if ts.tzinfo is None:
        ts = ts.replace(tzinfo=UTC)
    return ts


async def current_day_index(session: AsyncSession) -> int:
    gt = await genesis_time(session)
    if gt is None:
        return 0
    delta = datetime.now(UTC) - gt
    return max(0, delta.days)


async def run_emission(session: AsyncSession, day_index: int) -> Decimal:
    """Mint a single day's emission, distributed by verified CU. Idempotent per day_index.

    Returns the amount actually minted (may be 0 if no verified compute that day).
    """
    already = (
        await session.execute(select(EmissionEpoch).where(EmissionEpoch.day_index == day_index))
    ).scalar_one_or_none()
    if already is not None:
        return ZERO

    period = tk.halving_index(day_index)
    budget = tk.daily_emission(period)

    # Clamp to remaining supply.
    minted = await current_minted(session, tk.SYMBOL)
    remaining = tk.MAX_SUPPLY - minted
    if remaining <= 0:
        budget = ZERO
    elif budget > remaining:
        budget = tk.quantize(remaining)

    # Lazy import to avoid cycle (ledger -> compute -> relay models).
    from app.modules.compute import service as compute_service

    cu_by_supplier: dict[int, Decimal] = await compute_service.verified_cu_by_supplier(
        session, day_index
    )
    total_cu = sum(cu_by_supplier.values(), ZERO)

    total_minted = ZERO
    if budget > 0 and total_cu > 0:
        for user_id, cu in cu_by_supplier.items():
            share = tk.quantize(budget * cu / total_cu)
            if share <= 0:
                continue
            await mint(
                session,
                dest_ref=f"user:{user_id}",
                asset=tk.SYMBOL,
                amount=share,
                kind="emission",
                ref=f"emission:day:{day_index}",
                memo=f"day {day_index} CU={cu}",
            )
            total_minted += share

    session.add(
        EmissionEpoch(
            day_index=day_index,
            halving_period=period,
            minted=total_minted,
            processed_at=datetime.now(UTC),
        )
    )
    await session.flush()
    if total_minted > 0:
        log.info("Emission day=%d period=%d minted=%s CC", day_index, period, total_minted)
    return total_minted


async def run_due_emissions(session: AsyncSession, max_backfill_days: int = 400) -> Decimal:
    """Process emission for all COMPLETED days that have not been processed yet.

    The current (incomplete) day is intentionally not emitted until it completes, so a day's
    distribution reflects the full day's verified CU. Idempotent.
    """
    current = await current_day_index(session)
    if current <= 0:
        return ZERO
    processed = set((await session.execute(select(EmissionEpoch.day_index))).scalars().all())
    start = max(0, current - max_backfill_days)
    total = ZERO
    for day in range(start, current):
        if day not in processed:
            total += await run_emission(session, day)
    return total


async def supply_info(session: AsyncSession) -> dict:
    minted = await current_minted(session, tk.SYMBOL)
    return {
        "symbol": tk.SYMBOL,
        "decimals": tk.DECIMALS,
        "max_supply": tk.MAX_SUPPLY,
        "genesis_supply": tk.GENESIS_SUPPLY,
        "minted": minted,
        "remaining_to_mint": tk.MAX_SUPPLY - minted,
    }


async def conservation_ok(session: AsyncSession, asset: str) -> bool:
    """Sum of ALL balances for an asset must be exactly 0 (invariant #1)."""
    rows = (await session.execute(select(Balance.amount).where(Balance.asset == asset))).scalars()
    return sum(rows, ZERO) == ZERO
