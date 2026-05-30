"""Ledger schemas."""

from __future__ import annotations

from pydantic import BaseModel, Field

from app.schemas.common import DecimalStr


class BalanceOut(BaseModel):
    asset: str
    amount: DecimalStr


class AccountOut(BaseModel):
    ref: str
    owner_type: str
    balances: list[BalanceOut]


class SupplyOut(BaseModel):
    symbol: str
    decimals: int
    max_supply: DecimalStr
    genesis_supply: DecimalStr
    minted: DecimalStr  # genesis + emission minted so far
    remaining_to_mint: DecimalStr


class TransferRequest(BaseModel):
    to_ref: str = Field(min_length=1, max_length=128)
    asset: str = Field(default="CC", max_length=16)
    amount: DecimalStr
    memo: str | None = Field(default=None, max_length=255)


class FaucetRequest(BaseModel):
    """Testnet-only: mint demo quote currency so users can trade in the MVP."""

    asset: str = Field(default="USDC", max_length=16)
    amount: DecimalStr = Field(default=1000)
