"""Compute oracle service.

CU standardization (pure functions, coefficients from tokenomics.py) + a PoDC verification
entry point. The MVP verification marks a source (and its pending deliveries) as verified and
bumps reputation; real benchmark challenges / cross-node / TEE attestation are backlog
(docs/workflow.md T2.2 / T2.4). See CLAUDE.md compute invariants.
"""

from __future__ import annotations

from datetime import UTC, datetime
from decimal import Decimal

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app import tokenomics as tk
from app.core.errors import BadRequest, NotFound
from app.models.relay import CapacitySource, DeliveryRecord

ZERO = Decimal("0")


# ---------------------------------------------------------------------------
# CU standardization (pure)
# ---------------------------------------------------------------------------
def tokens_to_cu(model: str | None, tokens: Decimal) -> Decimal:
    """Convert delivered tokens to CU using tokenomics.TOKEN_TO_CU (per 1,000 tokens)."""
    if tokens <= 0:
        return ZERO
    coeff = tk.TOKEN_TO_CU.get(model or "default", tk.TOKEN_TO_CU["default"])
    return tk.quantize(tokens / Decimal(1000) * coeff)


def gpu_hours_to_cu(
    gpu_model: str,
    hours: Decimal,
    *,
    bandwidth_factor: Decimal = Decimal("1"),
    interconnect: str = "pcie",
    availability: Decimal = Decimal("1"),
) -> Decimal:
    """CU = TFLOPS_fp16 * hours * bandwidth_factor * interconnect_factor * availability."""
    if hours <= 0:
        return ZERO
    tflops = tk.GPU_TFLOPS_FP16.get(gpu_model)
    if tflops is None:
        raise BadRequest(f"unknown gpu_model '{gpu_model}'")
    icf = tk.INTERCONNECT_FACTOR.get(interconnect, Decimal("1"))
    return tk.quantize(tflops * hours * bandwidth_factor * icf * availability)


def delivery_to_cu(record: DeliveryRecord) -> Decimal:
    """Total CU implied by a delivery record (token-based + gpu-hour-based)."""
    cu = ZERO
    if record.delivered_tokens and record.delivered_tokens > 0:
        cu += tokens_to_cu(record.model, record.delivered_tokens)
    if record.gpu_hours and record.gpu_hours > 0 and record.gpu_model:
        cu += gpu_hours_to_cu(record.gpu_model, record.gpu_hours)
    return tk.quantize(cu)


# ---------------------------------------------------------------------------
# Verification (PoDC) — MVP
# ---------------------------------------------------------------------------
async def run_verification(session: AsyncSession, source_id: int) -> CapacitySource:
    source = (
        await session.execute(select(CapacitySource).where(CapacitySource.id == source_id))
    ).scalar_one_or_none()
    if source is None:
        raise NotFound("capacity source not found")

    # MVP: accept and verify. Real PoDC (benchmark challenge, cross-node, TEE) -> backlog T2.2/T2.4.
    source.verified = True
    source.reputation = (source.reputation or ZERO) + Decimal("1")
    source.last_verified_at = datetime.now(UTC)

    # Back-fill: mark this source's unverified delivery records as verified.
    records = (
        await session.execute(
            select(DeliveryRecord).where(
                DeliveryRecord.source_id == source_id, DeliveryRecord.verified.is_(False)
            )
        )
    ).scalars()
    for rec in records:
        rec.verified = True
        if rec.cu == ZERO:
            rec.cu = delivery_to_cu(rec)
    await session.flush()
    return source


async def verified_cu_by_supplier(session: AsyncSession, day_index: int) -> dict[int, Decimal]:
    """Sum verified CU per supplier (user_id) for a given day index."""
    stmt = (
        select(CapacitySource.user_id, DeliveryRecord.cu)
        .join(CapacitySource, CapacitySource.id == DeliveryRecord.source_id)
        .where(DeliveryRecord.period == day_index, DeliveryRecord.verified.is_(True))
    )
    rows = (await session.execute(stmt)).all()
    out: dict[int, Decimal] = {}
    for user_id, cu in rows:
        out[user_id] = out.get(user_id, ZERO) + (cu or ZERO)
    return {k: v for k, v in out.items() if v > 0}
