"""Thin HTTP adapter for sub2api (https://github.com/Wei-Shaw/sub2api).

IMPORTANT (docs/compliance.md): ComputeCoin integrates with sub2api as a SEPARATE process
over HTTP. We do NOT fork or import sub2api source code. This keeps ComputeCoin under MIT and
isolates sub2api's LGPL.

Every method degrades gracefully to a local mock when SUB2API_BASE_URL is not configured, so
this repository runs and tests fully on its own. To go live, set SUB2API_BASE_URL /
SUB2API_ADMIN_TOKEN in .env and flesh out the methods against sub2api's real endpoints
(good first tasks for /workflow: docs/workflow.md T2.1).
"""

from __future__ import annotations

from datetime import datetime
from decimal import Decimal
from typing import Any

import httpx

from app.core.config import settings
from app.core.logging import get_logger

log = get_logger(__name__)


class Sub2ApiClient:
    def __init__(self, base_url: str | None = None, token: str | None = None) -> None:
        self.base_url = (base_url or settings.sub2api_base_url or "").rstrip("/")
        self.token = token or settings.sub2api_admin_token

    @property
    def configured(self) -> bool:
        return bool(self.base_url)

    def _headers(self) -> dict[str, str]:
        h = {"Accept": "application/json"}
        if self.token:
            h["Authorization"] = f"Bearer {self.token}"
        return h

    async def health(self) -> dict[str, Any]:
        if not self.configured:
            return {"configured": False, "status": "mock", "note": "SUB2API_BASE_URL not set"}
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                # sub2api exposes a health/status endpoint; adjust path when wiring for real.
                resp = await client.get(f"{self.base_url}/healthz", headers=self._headers())
                return {"configured": True, "status_code": resp.status_code}
        except Exception as exc:  # noqa: BLE001
            log.warning("sub2api health check failed: %s", exc)
            return {"configured": True, "status": "unreachable", "error": str(exc)}

    async def fetch_usage(
        self, api_key_ref: str, since: datetime | None = None
    ) -> dict[str, Decimal]:
        """Return delivered tokens per model for a supplier's key, since a timestamp.

        Mock returns empty (no fabricated usage). Real impl should call sub2api's usage/billing
        API and map model -> delivered tokens. Contract intentionally minimal; see T2.1.
        """
        if not self.configured:
            return {}
        # TODO(T2.1): implement against sub2api usage endpoint.
        return {}


_default_client: Sub2ApiClient | None = None


def get_client() -> Sub2ApiClient:
    global _default_client
    if _default_client is None:
        _default_client = Sub2ApiClient()
    return _default_client
