"""Ledger HTTP routes (thin)."""

from __future__ import annotations

from fastapi import APIRouter

from app.api.deps import CurrentUser, SessionDep
from app.core.errors import Forbidden
from app.modules.ledger import service as ledger_service
from app.schemas.ledger import (
    AccountOut,
    BalanceOut,
    FaucetRequest,
    SupplyOut,
    TransferRequest,
)

router = APIRouter(prefix="/ledger", tags=["ledger"])

# System accounts whose balances are public (transparency).
_PUBLIC_PREFIXES = ("treasury", "genesis:", "mint_source", "mm", "emission")


def _is_public(ref: str) -> bool:
    return any(ref == p or ref.startswith(p) for p in _PUBLIC_PREFIXES)


@router.get("/supply", response_model=SupplyOut)
async def supply(session: SessionDep) -> SupplyOut:
    return SupplyOut.model_validate(await ledger_service.supply_info(session))


@router.get("/balance", response_model=AccountOut)
async def my_balance(user: CurrentUser, session: SessionDep) -> AccountOut:
    balances = await ledger_service.balances_for(session, user.account_ref)
    return AccountOut(
        ref=user.account_ref,
        owner_type="user",
        balances=[BalanceOut(asset=a, amount=amt) for a, amt in balances.items()],
    )


@router.get("/accounts/{ref}", response_model=AccountOut)
async def account(ref: str, user: CurrentUser, session: SessionDep) -> AccountOut:
    if not _is_public(ref) and ref != user.account_ref and user.role != "admin":
        raise Forbidden("not allowed to view this account")
    balances = await ledger_service.balances_for(session, ref)
    owner_type = "user" if ref.startswith("user:") else "system"
    return AccountOut(
        ref=ref,
        owner_type=owner_type,
        balances=[BalanceOut(asset=a, amount=amt) for a, amt in balances.items()],
    )


@router.post("/transfer")
async def transfer(body: TransferRequest, user: CurrentUser, session: SessionDep) -> dict:
    entry = await ledger_service.transfer(
        session,
        from_ref=user.account_ref,
        to_ref=body.to_ref,
        asset=body.asset,
        amount=body.amount,
        kind="transfer",
        memo=body.memo,
    )
    return {"entry_id": entry.id, "status": "ok"}


@router.post("/faucet")
async def faucet(body: FaucetRequest, user: CurrentUser, session: SessionDep) -> dict:
    """TESTNET ONLY: mint demo quote currency so you can try trading in the MVP.

    Disabled outside development to avoid minting real value. The CC token is NEVER mintable
    this way — only via transparent genesis/emission (CLAUDE.md invariants #2/#3).
    """
    from app import tokenomics as tk
    from app.core.config import settings

    if settings.environment == "production":
        raise Forbidden("faucet disabled in production")
    if body.asset == tk.SYMBOL:
        raise Forbidden("CC cannot be minted via faucet; it is emission-only")

    entry = await ledger_service.mint(
        session,
        dest_ref=user.account_ref,
        asset=body.asset,
        amount=body.amount,
        kind="deposit",
        ref=f"faucet:{user.id}",
        memo="testnet faucet",
    )
    return {"entry_id": entry.id, "asset": body.asset, "amount": format(body.amount, "f")}
