"""Identity service: users, auth, KYC status.

On registration a user gets a corresponding ledger account (user:<id>). Passwords are only
ever stored as bcrypt hashes.
"""

from __future__ import annotations

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.errors import Conflict, NotFound
from app.core.security import hash_password, verify_password
from app.models.user import User


async def get_by_email(session: AsyncSession, email: str) -> User | None:
    return (
        await session.execute(select(User).where(User.email == email.lower()))
    ).scalar_one_or_none()


async def get_by_id(session: AsyncSession, user_id: int) -> User | None:
    return (await session.execute(select(User).where(User.id == user_id))).scalar_one_or_none()


async def register(
    session: AsyncSession,
    email: str,
    password: str,
    role: str = "user",
    *,
    auth_provider: str = "local",
    external_id: str | None = None,
) -> User:
    email = email.lower()
    if await get_by_email(session, email) is not None:
        raise Conflict("email already registered")
    user = User(
        email=email,
        password_hash=hash_password(password),
        role=role,
        auth_provider=auth_provider,
        external_id=external_id,
    )
    session.add(user)
    await session.flush()  # assign id

    # Create the user's ledger account up front.
    from app.modules.ledger import service as ledger_service

    await ledger_service.get_or_create_account(session, user.account_ref, owner_type="user")
    return user


async def authenticate(session: AsyncSession, email: str, password: str) -> User | None:
    user = await get_by_email(session, email)
    if user is None or not verify_password(password, user.password_hash):
        return None
    return user


async def set_kyc_status(session: AsyncSession, user_id: int, status: str) -> User:
    user = await get_by_id(session, user_id)
    if user is None:
        raise NotFound("user not found")
    user.kyc_status = status
    await session.flush()
    return user


async def ensure_user(session: AsyncSession, email: str, password: str, role: str) -> User:
    """Idempotently ensure a user exists (used for bootstrap admin / market maker)."""
    existing = await get_by_email(session, email)
    if existing is not None:
        return existing
    return await register(session, email, password, role)
