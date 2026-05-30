"""Startup bootstrap: transparent genesis, admin account, and demo market-maker liquidity.

Called once during app startup (after tables exist). Everything here is idempotent.
"""

from __future__ import annotations

import secrets
from decimal import Decimal

from sqlalchemy.ext.asyncio import AsyncSession

from app.core.config import settings
from app.core.logging import get_logger
from app.modules.identity import service as identity_service
from app.modules.ledger import service as ledger_service
from app.modules.marketmaker import service as mm_service

log = get_logger(__name__)


async def bootstrap(session: AsyncSession) -> None:
    # 1) Transparent genesis allocation (also establishes genesis time via mint_source).
    created = await ledger_service.seed_genesis(session)
    if created:
        log.info("Genesis allocation minted.")

    # 2) Bootstrap admin (optional, from .env).
    if settings.admin_email and settings.admin_password:
        await identity_service.ensure_user(
            session, settings.admin_email, settings.admin_password, "admin"
        )
        log.info("Admin ensured: %s", settings.admin_email)

    # 3) Demo market-maker liquidity (never in production).
    if settings.environment != "production":
        await _seed_market_maker(session)


async def _seed_market_maker(session: AsyncSession) -> None:
    mm = await identity_service.ensure_user(
        session, mm_service.MM_EMAIL, secrets.token_urlsafe(24), "supplier"
    )

    # Fund the MM transparently: CC from the public ecosystem bucket, testnet USDC via mint.
    if await ledger_service.get_balance(session, mm.account_ref, "CC") == 0:
        await ledger_service.transfer(
            session,
            from_ref="genesis:ecosystem",
            to_ref=mm.account_ref,
            asset="CC",
            amount=Decimal("100000"),
            kind="transfer",
            memo="seed MM liquidity (ecosystem bucket)",
        )
    if await ledger_service.get_balance(session, mm.account_ref, "USDC") == 0:
        await ledger_service.mint(
            session,
            dest_ref=mm.account_ref,
            asset="USDC",
            amount=Decimal("100000"),
            kind="deposit",
            ref="seed:mm:usdc",
            memo="seed MM quote currency (testnet)",
        )

    result = await mm_service.refresh_quotes(session)
    log.info("Market maker seeded; quotes posted=%s", result.get("posted"))
