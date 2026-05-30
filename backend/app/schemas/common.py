"""Shared schema primitives."""

from __future__ import annotations

from decimal import Decimal
from typing import Annotated

from pydantic import PlainSerializer

# Decimal serialized as a plain string in JSON so the frontend (JS) never loses precision.
DecimalStr = Annotated[
    Decimal,
    PlainSerializer(lambda v: format(v, "f"), return_type=str, when_used="json"),
]
