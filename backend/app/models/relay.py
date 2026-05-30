"""Relay / capacity models.

A CapacitySource is a piece of compute capacity a supplier LEGITIMATELY OWNS and contributes
(a sub2api gateway they operate, or a GPU node). source_attestation is MANDATORY and the
source is unverified until Proof-of-Delivered-Compute passes (CLAUDE.md invariant #7).

DeliveryRecord captures measured delivery (tokens served and/or GPU-hours) for a given
day index ("period"), later converted to CU and used by ledger emission.
"""

from __future__ import annotations

from datetime import UTC, datetime
from decimal import Decimal

from sqlalchemy import Boolean, DateTime, ForeignKey, Integer, String, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.core.database import Base
from app.models.types import DecimalText


def _utcnow() -> datetime:
    return datetime.now(UTC)


class CapacitySource(Base):
    __tablename__ = "capacity_sources"

    id: Mapped[int] = mapped_column(primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id"), index=True)

    # "sub2api" | "gpu"
    kind: Mapped[str] = mapped_column(String(32))
    name: Mapped[str] = mapped_column(String(128))

    # For kind="sub2api": base URL of the supplier's own sub2api gateway.
    endpoint: Mapped[str | None] = mapped_column(String(512), nullable=True)
    # For kind="gpu": accelerator model key (see tokenomics.GPU_TFLOPS_FP16).
    gpu_model: Mapped[str | None] = mapped_column(String(64), nullable=True)

    # MANDATORY legality declaration. Without it, registration is rejected.
    source_attestation: Mapped[str] = mapped_column(Text)

    verified: Mapped[bool] = mapped_column(Boolean, default=False)
    reputation: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))
    last_verified_at: Mapped[datetime | None] = mapped_column(
        DateTime(timezone=True), nullable=True
    )
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)

    deliveries: Mapped[list[DeliveryRecord]] = relationship(
        back_populates="source", cascade="all, delete-orphan"
    )


class DeliveryRecord(Base):
    __tablename__ = "delivery_records"

    id: Mapped[int] = mapped_column(primary_key=True)
    source_id: Mapped[int] = mapped_column(ForeignKey("capacity_sources.id"), index=True)

    # token-delivery fields
    model: Mapped[str | None] = mapped_column(String(64), nullable=True)
    delivered_tokens: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))

    # gpu-hours fields
    gpu_model: Mapped[str | None] = mapped_column(String(64), nullable=True)
    gpu_hours: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))

    # Standardized compute, set after verification.
    cu: Mapped[Decimal] = mapped_column(DecimalText(), default=Decimal("0"))
    verified: Mapped[bool] = mapped_column(Boolean, default=False)

    # Day index since genesis; emission is grouped/distributed per period.
    period: Mapped[int] = mapped_column(Integer, index=True, default=0)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)

    source: Mapped[CapacitySource] = relationship(back_populates="deliveries")
