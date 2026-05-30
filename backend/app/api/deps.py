"""Shared API dependencies: current user, role checks, and the KYC compliance hook."""

from __future__ import annotations

from typing import Annotated

import jwt
from fastapi import Depends
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.config import settings
from app.core.database import get_session
from app.core.errors import Forbidden, Unauthorized
from app.core.security import decode_access_token
from app.models.user import User
from app.modules.identity import service as identity_service

_bearer = HTTPBearer(auto_error=False)

SessionDep = Annotated[AsyncSession, Depends(get_session)]


async def get_current_user(
    session: SessionDep,
    credentials: Annotated[HTTPAuthorizationCredentials | None, Depends(_bearer)],
) -> User:
    if credentials is None:
        raise Unauthorized("missing bearer token")
    try:
        payload = decode_access_token(credentials.credentials)
    except jwt.PyJWTError:
        raise Unauthorized("invalid or expired token") from None
    user_id = payload.get("sub")
    if user_id is None:
        raise Unauthorized("invalid token subject")
    user = await identity_service.get_by_id(session, int(user_id))
    if user is None:
        raise Unauthorized("user not found")
    return user


CurrentUser = Annotated[User, Depends(get_current_user)]


def require_role(*roles: str):
    async def _checker(user: CurrentUser) -> User:
        if user.role not in roles:
            raise Forbidden(f"requires role in {roles}")
        return user

    return _checker


async def require_kyc(user: CurrentUser) -> User:
    """Compliance hook (docs/compliance.md). Off by default in MVP."""
    if settings.require_kyc_for_trading and user.kyc_status != "approved":
        raise Forbidden("KYC approval required for this action")
    return user
