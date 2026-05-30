"""Auth / identity schemas."""

from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, EmailStr, Field


class RegisterRequest(BaseModel):
    email: EmailStr
    password: str = Field(min_length=8, max_length=128)
    # Self-selected role for MVP. "supplier" if you intend to contribute compute.
    role: str = Field(default="user", pattern="^(user|supplier)$")


class LoginRequest(BaseModel):
    email: EmailStr
    password: str


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"


class UserOut(BaseModel):
    id: int
    email: EmailStr
    role: str
    kyc_status: str
    auth_provider: str
    created_at: datetime

    model_config = {"from_attributes": True}
