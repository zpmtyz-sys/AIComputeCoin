"""Relay module tests: capacity registration and delivery recording."""

from decimal import Decimal

import pytest

from app.core.errors import Forbidden
from app.modules.identity import service as identity
from app.modules.relay import service as relay
from app.schemas.relay import CapacitySourceCreate, DeliveryCreate


async def test_register_source_sub2api(session):
    user = await identity.register(session, "sup@x.com", "password123", "supplier")
    source = await relay.register_source(
        session,
        user.id,
        CapacitySourceCreate(
            kind="sub2api",
            name="My relay",
            endpoint="http://localhost:8080",
            source_attestation="I legitimately operate this sub2api gateway on my own server.",
        ),
    )
    assert source.id > 0
    assert source.verified is False
    assert source.kind == "sub2api"


async def test_register_source_gpu(session):
    user = await identity.register(session, "gpu@x.com", "password123", "supplier")
    source = await relay.register_source(
        session,
        user.id,
        CapacitySourceCreate(
            kind="gpu",
            name="My H100",
            gpu_model="H100",
            source_attestation="I own this H100 server in my data center, serial #XYZ.",
        ),
    )
    assert source.gpu_model == "H100"
    assert source.verified is False


async def test_record_delivery_and_cu_calculation(session):
    user = await identity.register(session, "del@x.com", "password123", "supplier")
    source = await relay.register_source(
        session,
        user.id,
        CapacitySourceCreate(
            kind="sub2api",
            name="relay1",
            endpoint="http://localhost:8080",
            source_attestation="I own and operate this gateway legitimately.",
        ),
    )
    record = await relay.record_delivery(
        session,
        user.id,
        source.id,
        DeliveryCreate(model="claude-3.5-sonnet", delivered_tokens=Decimal("100000")),
    )
    assert record.cu > 0
    # Unverified source -> delivery.verified = False
    assert record.verified is False


async def test_delivery_ownership_enforced(session):
    owner = await identity.register(session, "own@x.com", "password123", "supplier")
    other = await identity.register(session, "oth@x.com", "password123", "supplier")
    source = await relay.register_source(
        session,
        owner.id,
        CapacitySourceCreate(
            kind="gpu",
            name="node",
            gpu_model="A100",
            source_attestation="My own hardware in my DC.",
        ),
    )
    with pytest.raises(Forbidden):
        await relay.record_delivery(
            session,
            other.id,
            source.id,
            DeliveryCreate(gpu_model="A100", gpu_hours=Decimal("1")),
        )


async def test_list_sources(session):
    user = await identity.register(session, "ls@x.com", "password123", "supplier")
    await relay.register_source(
        session,
        user.id,
        CapacitySourceCreate(
            kind="gpu",
            name="node1",
            gpu_model="RTX4090",
            source_attestation="My gaming rig, legally purchased.",
        ),
    )
    sources = await relay.list_sources(session, user.id)
    assert len(sources) == 1
    assert sources[0].name == "node1"
