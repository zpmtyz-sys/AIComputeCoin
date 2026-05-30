"""Compute oracle HTTP routes (thin)."""

from __future__ import annotations

from decimal import Decimal

from fastapi import APIRouter, Query

from app import tokenomics as tk
from app.api.deps import CurrentUser, SessionDep
from app.core.errors import BadRequest, Forbidden
from app.modules.compute import service as compute_service
from app.modules.relay import service as relay_service
from app.schemas.relay import CapacitySourceOut

router = APIRouter(prefix="/compute", tags=["compute"])


@router.get("/cu/quote")
async def cu_quote(
    model: str | None = Query(default=None),
    tokens: Decimal = Query(default=Decimal("0")),
    gpu_model: str | None = Query(default=None),
    hours: Decimal = Query(default=Decimal("0")),
) -> dict:
    """Estimate CU for a token delivery and/or GPU-hours. Pure, public."""
    cu = Decimal("0")
    if tokens > 0:
        cu += compute_service.tokens_to_cu(model, tokens)
    if hours > 0:
        if not gpu_model:
            raise BadRequest("gpu_model required when hours > 0")
        cu += compute_service.gpu_hours_to_cu(gpu_model, hours)
    return {"cu": format(tk.quantize(cu), "f"), "definition": "1 CU = 1 TFLOPS * 1h FP16"}


@router.post("/verify/{source_id}", response_model=CapacitySourceOut)
async def verify_source(
    source_id: int, user: CurrentUser, session: SessionDep
) -> CapacitySourceOut:
    source = await relay_service.get_source(session, source_id)
    if user.role != "admin" and source.user_id != user.id:
        raise Forbidden("not allowed to verify this source")
    verified = await compute_service.run_verification(session, source_id)
    return CapacitySourceOut.model_validate(verified)
