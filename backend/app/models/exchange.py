"""Exchange models — spot orders and trades.

Settlement happens in the ledger. Reserved/locked funds for resting limit orders are DERIVED
from open orders (see exchange.service.locked_amount), not stored as escrow entries.
"""

from __future__ import annotations

from datetime import UTC, datetime
from decimal import Decimal

from sqlalchemy import DateTime, ForeignKey, Integer, String
from sqlalchemy.orm import Mapped, mapped_column

from app.core.database import Base
from app.models.types import DecimalText


def _utcnow() -> datetime:
    return datetime.now(UTC)


class Order(Base):
    __tablename__ = "orders"

    id: Mapped[int] = mapped_column(primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id"), index=True)

    pair: Mapped[str] = mapped_column(String(32), index=True)  # e.g. "CC/USDC"
    side: Mapped[str] = mapped_column(String(8))  # buy | sell
    type: Mapped[str] = mapped_column(String(8))  # limit | market

    price: Mapped[Decimal | None] = mapped_column(DecimalText(), nullable=True)  # None for market
    qty: Mapped[Decimal] = mapped_column(DecimalText())
    filled: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))

    # open | partial | filled | cancelled
    status: Mapped[str] = mapped_column(String(16), default="open", index=True)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)

    @property
    def remaining(self) -> Decimal:
        return self.qty - self.filled


class Trade(Base):
    __tablename__ = "trades"

    id: Mapped[int] = mapped_column(primary_key=True)
    pair: Mapped[str] = mapped_column(String(32), index=True)
    price: Mapped[Decimal] = mapped_column(DecimalText())
    qty: Mapped[Decimal] = mapped_column(DecimalText())

    taker_order_id: Mapped[int] = mapped_column(ForeignKey("orders.id"))
    maker_order_id: Mapped[int] = mapped_column(ForeignKey("orders.id"))
    taker_user_id: Mapped[int] = mapped_column(Integer)
    maker_user_id: Mapped[int] = mapped_column(Integer)

    taker_fee: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))
    maker_fee: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)
