"""Exchange schemas."""

from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, Field, model_validator

from app.schemas.common import DecimalStr


class OrderCreate(BaseModel):
    pair: str = Field(default="CC/USDC", max_length=32)
    side: str = Field(pattern="^(buy|sell)$")
    type: str = Field(default="limit", pattern="^(limit|market)$")
    price: DecimalStr | None = None
    qty: DecimalStr

    @model_validator(mode="after")
    def _check(self) -> OrderCreate:
        if self.type == "limit" and (self.price is None or self.price <= 0):
            raise ValueError("limit order requires price > 0")
        if self.qty is None or self.qty <= 0:
            raise ValueError("qty must be > 0")
        return self


class OrderOut(BaseModel):
    id: int
    user_id: int
    pair: str
    side: str
    type: str
    price: DecimalStr | None
    qty: DecimalStr
    filled: DecimalStr
    status: str
    created_at: datetime

    model_config = {"from_attributes": True}


class TradeOut(BaseModel):
    id: int
    pair: str
    price: DecimalStr
    qty: DecimalStr
    taker_fee: DecimalStr
    maker_fee: DecimalStr
    created_at: datetime

    model_config = {"from_attributes": True}


class BookLevel(BaseModel):
    price: DecimalStr
    qty: DecimalStr


class OrderBookOut(BaseModel):
    pair: str
    bids: list[BookLevel]  # highest price first
    asks: list[BookLevel]  # lowest price first
