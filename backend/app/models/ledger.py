"""Ledger models — transparent, double-entry, multi-asset.

Accounting model (CLAUDE.md invariant #1):
  * Every LedgerEntry debits one account and credits another by the SAME amount and asset.
  * Therefore the sum of ALL balances per asset is ALWAYS exactly 0.
  * `mint_source` is the system account that goes negative as tokens are minted; total minted
    of an asset = -balance(mint_source, asset). For CC this must never exceed MAX_SUPPLY (#2).

`Balance` is a denormalized per-(account, asset) running total, updated transactionally on
every entry, so balance reads are O(1) and conservation checks are cheap.
"""

from __future__ import annotations

from datetime import UTC, datetime
from decimal import Decimal

from sqlalchemy import DateTime, ForeignKey, Integer, String, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column

from app.core.database import Base
from app.models.types import DecimalText


def _utcnow() -> datetime:
    return datetime.now(UTC)


class Account(Base):
    __tablename__ = "accounts"

    id: Mapped[int] = mapped_column(primary_key=True)
    # Stable reference, e.g. "user:1", "treasury", "mint_source", "mm", "genesis:team".
    ref: Mapped[str] = mapped_column(String(128), unique=True, index=True)
    # "system" | "user"
    owner_type: Mapped[str] = mapped_column(String(32), default="system")
    kind: Mapped[str] = mapped_column(String(32), default="wallet")
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)


class Balance(Base):
    __tablename__ = "balances"
    __table_args__ = (UniqueConstraint("account_id", "asset", name="uq_balance_account_asset"),)

    id: Mapped[int] = mapped_column(primary_key=True)
    account_id: Mapped[int] = mapped_column(ForeignKey("accounts.id"), index=True)
    asset: Mapped[str] = mapped_column(String(16), index=True)
    amount: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))


class LedgerEntry(Base):
    __tablename__ = "ledger_entries"

    id: Mapped[int] = mapped_column(primary_key=True)
    # grouping / idempotency reference, e.g. "genesis:team", "emission:day:5", "trade:42".
    ref: Mapped[str] = mapped_column(String(128), index=True)
    asset: Mapped[str] = mapped_column(String(16), index=True)
    debit_account_id: Mapped[int] = mapped_column(ForeignKey("accounts.id"), index=True)
    credit_account_id: Mapped[int] = mapped_column(ForeignKey("accounts.id"), index=True)
    amount: Mapped[Decimal] = mapped_column(DecimalText())
    # genesis | emission | transfer | trade | fee | deposit | withdrawal | burn
    kind: Mapped[str] = mapped_column(String(32), index=True)
    memo: Mapped[str | None] = mapped_column(String(255), nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)


class EmissionEpoch(Base):
    """Records that a given day index has been processed (idempotent emission)."""

    __tablename__ = "emission_epochs"

    day_index: Mapped[int] = mapped_column(Integer, primary_key=True)
    halving_period: Mapped[int] = mapped_column(Integer)
    minted: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))
    processed_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)
