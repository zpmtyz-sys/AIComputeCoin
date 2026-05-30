"""Identity HTTP routes (thin)."""

from __future__ import annotations

from fastapi import APIRouter

from app.api.deps import CurrentUser, SessionDep
from app.core.errors import Unauthorized
from app.core.security import create_access_token
from app.modules.identity import service as identity_service
from app.schemas.auth import LoginRequest, RegisterRequest, TokenResponse, UserOut

router = APIRouter(prefix="/auth", tags=["identity"])


@router.post("/register", response_model=UserOut, status_code=201)
async def register(body: RegisterRequest, session: SessionDep) -> UserOut:
    user = await identity_service.register(session, body.email, body.password, body.role)
    return UserOut.model_validate(user)


@router.post("/login", response_model=TokenResponse)
async def login(body: LoginRequest, session: SessionDep) -> TokenResponse:
    user = await identity_service.authenticate(session, body.email, body.password)
    if user is None:
        raise Unauthorized("invalid email or password")
    token = create_access_token(user.id, extra={"role": user.role})
    return TokenResponse(access_token=token)


@router.get("/me", response_model=UserOut)
async def me(user: CurrentUser) -> UserOut:
    return UserOut.model_validate(user)
