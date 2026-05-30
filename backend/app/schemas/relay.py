"""Relay / capacity schemas."""

from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, Field, model_validator

from app.schemas.common import DecimalStr


class CapacitySourceCreate(BaseModel):
    kind: str = Field(pattern="^(sub2api|gpu)$")
    name: str = Field(min_length=1, max_length=128)
    endpoint: str | None = Field(default=None, max_length=512)
    gpu_model: str | None = Field(default=None, max_length=64)
    # MANDATORY: declare the capacity is legitimately owned/operated by you.
    source_attestation: str = Field(min_length=10, max_length=2000)

    @model_validator(mode="after")
    def _check_kind_fields(self) -> CapacitySourceCreate:
        if self.kind == "sub2api" and not self.endpoint:
            raise ValueError("endpoint is required for kind=sub2api")
        if self.kind == "gpu" and not self.gpu_model:
            raise ValueError("gpu_model is required for kind=gpu")
        return self


class CapacitySourceOut(BaseModel):
    id: int
    user_id: int
    kind: str
    name: str
    endpoint: str | None
    gpu_model: str | None
    verified: bool
    reputation: DecimalStr
    created_at: datetime

    model_config = {"from_attributes": True}


class DeliveryCreate(BaseModel):
    model: str | None = None
    delivered_tokens: DecimalStr = Field(default=0)
    gpu_model: str | None = None
    gpu_hours: DecimalStr = Field(default=0)


class DeliveryOut(BaseModel):
    id: int
    source_id: int
    model: str | None
    delivered_tokens: DecimalStr
    gpu_model: str | None
    gpu_hours: DecimalStr
    cu: DecimalStr
    verified: bool
    period: int
    created_at: datetime

    model_config = {"from_attributes": True}
