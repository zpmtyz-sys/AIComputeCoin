"""User model (identity module)."""

from __future__ import annotations

from datetime import UTC, datetime

from sqlalchemy import DateTime, String
from sqlalchemy.orm import Mapped, mapped_column

from app.core.database import Base


def _utcnow() -> datetime:
    return datetime.now(UTC)


class User(Base):
    __tablename__ = "users"

    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    password_hash: Mapped[str] = mapped_column(String(255))

    # user | supplier | admin
    role: Mapped[str] = mapped_column(String(32), default="user")

    # Compliance hook (docs/compliance.md): none | pending | approved | rejected
    kyc_status: Mapped[str] = mapped_column(String(32), default="none")

    # sub2api SSO support (mode B). "local" for native accounts.
    auth_provider: Mapped[str] = mapped_column(String(32), default="local")
    external_id: Mapped[str | None] = mapped_column(String(128), nullable=True, index=True)

    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=_utcnow)

    @property
    def account_ref(self) -> str:
        return f"user:{self.id}"
