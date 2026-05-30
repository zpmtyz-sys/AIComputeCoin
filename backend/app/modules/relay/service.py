"""Relay service: register supplier capacity and record delivery (metered compute).

A capacity source is unverified at creation; its deliveries do not count toward emission
until verified (CLAUDE.md invariant #7). source_attestation is enforced by the schema.
"""

from __future__ import annotations

from collections.abc import Sequence
from decimal import Decimal

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.errors import Forbidden, NotFound
from app.models.relay import CapacitySource, DeliveryRecord
from app.schemas.relay import CapacitySourceCreate, DeliveryCreate

ZERO = Decimal("0")


async def register_source(
    session: AsyncSession, user_id: int, data: CapacitySourceCreate
) -> CapacitySource:
    source = CapacitySource(
        user_id=user_id,
        kind=data.kind,
        name=data.name,
        endpoint=data.endpoint,
        gpu_model=data.gpu_model,
        source_attestation=data.source_attestation,
        verified=False,
        reputation=ZERO,
    )
    session.add(source)
    await session.flush()
    return source


async def list_sources(session: AsyncSession, user_id: int) -> Sequence[CapacitySource]:
    stmt = (
        select(CapacitySource)
        .where(CapacitySource.user_id == user_id)
        .order_by(CapacitySource.id.desc())
    )
    return (await session.execute(stmt)).scalars().all()


async def get_source(session: AsyncSession, source_id: int) -> CapacitySource:
    source = (
        await session.execute(select(CapacitySource).where(CapacitySource.id == source_id))
    ).scalar_one_or_none()
    if source is None:
        raise NotFound("capacity source not found")
    return source


async def record_delivery(
    session: AsyncSession, user_id: int, source_id: int, data: DeliveryCreate
) -> DeliveryRecord:
    source = await get_source(session, source_id)
    if source.user_id != user_id:
        raise Forbidden("not your capacity source")

    # Lazy cross-module imports (CLAUDE.md): compute for CU, ledger for the current day index.
    from app.modules.compute import service as compute_service
    from app.modules.ledger import service as ledger_service

    period = await ledger_service.current_day_index(session)

    record = DeliveryRecord(
        source_id=source_id,
        model=data.model,
        delivered_tokens=data.delivered_tokens or ZERO,
        gpu_model=data.gpu_model or source.gpu_model,
        gpu_hours=data.gpu_hours or ZERO,
        period=period,
        verified=source.verified,
    )
    record.cu = compute_service.delivery_to_cu(record)
    session.add(record)
    await session.flush()
    return record


async def list_deliveries(
    session: AsyncSession, source_id: int, limit: int = 100
) -> Sequence[DeliveryRecord]:
    stmt = (
        select(DeliveryRecord)
        .where(DeliveryRecord.source_id == source_id)
        .order_by(DeliveryRecord.id.desc())
        .limit(limit)
    )
    return (await session.execute(stmt)).scalars().all()
