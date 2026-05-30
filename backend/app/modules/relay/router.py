"""Relay HTTP routes (thin)."""

from __future__ import annotations

from fastapi import APIRouter

from app.api.deps import CurrentUser, SessionDep
from app.modules.relay import service as relay_service
from app.modules.relay.sub2api_client import get_client
from app.schemas.relay import (
    CapacitySourceCreate,
    CapacitySourceOut,
    DeliveryCreate,
    DeliveryOut,
)

router = APIRouter(prefix="/relay", tags=["relay"])


@router.get("/sub2api/health")
async def sub2api_health() -> dict:
    return await get_client().health()


@router.post("/sources", response_model=CapacitySourceOut, status_code=201)
async def create_source(
    body: CapacitySourceCreate, user: CurrentUser, session: SessionDep
) -> CapacitySourceOut:
    source = await relay_service.register_source(session, user.id, body)
    return CapacitySourceOut.model_validate(source)


@router.get("/sources", response_model=list[CapacitySourceOut])
async def list_sources(user: CurrentUser, session: SessionDep) -> list[CapacitySourceOut]:
    sources = await relay_service.list_sources(session, user.id)
    return [CapacitySourceOut.model_validate(s) for s in sources]


@router.post("/sources/{source_id}/delivery", response_model=DeliveryOut, status_code=201)
async def record_delivery(
    source_id: int, body: DeliveryCreate, user: CurrentUser, session: SessionDep
) -> DeliveryOut:
    record = await relay_service.record_delivery(session, user.id, source_id, body)
    return DeliveryOut.model_validate(record)


@router.get("/sources/{source_id}/deliveries", response_model=list[DeliveryOut])
async def list_deliveries(
    source_id: int, user: CurrentUser, session: SessionDep
) -> list[DeliveryOut]:
    source = await relay_service.get_source(session, source_id)
    # ownership check kept simple; admins could see all (backlog)
    if source.user_id != user.id:
        return []
    records = await relay_service.list_deliveries(session, source_id)
    return [DeliveryOut.model_validate(r) for r in records]
