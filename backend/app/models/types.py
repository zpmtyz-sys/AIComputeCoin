"""Custom SQLAlchemy column types.

DecimalText stores a Python Decimal as a canonical, non-scientific string. This guarantees
EXACT precision on both SQLite (local/test) and PostgreSQL (prod), avoiding the float
rounding that SQLAlchemy's Numeric can introduce on SQLite. See CLAUDE.md invariant #6.
"""

from __future__ import annotations

from decimal import Decimal

from sqlalchemy import String
from sqlalchemy.types import TypeDecorator


class DecimalText(TypeDecorator):
    impl = String
    cache_ok = True

    def __init__(self, length: int = 64) -> None:
        super().__init__(length)

    def process_bind_param(self, value, dialect):  # noqa: ANN001
        if value is None:
            return None
        if not isinstance(value, Decimal):
            value = Decimal(str(value))
        # 'f' format -> plain decimal string (no exponent), preserves precision.
        return format(value, "f")

    def process_result_value(self, value, dialect):  # noqa: ANN001
        if value is None:
            return None
        return Decimal(value)
