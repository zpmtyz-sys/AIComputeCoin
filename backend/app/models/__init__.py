"""ORM models. Importing this package registers every model on Base.metadata.

Persistence layer = single source of truth for data shape. Cross-module table access is
forbidden (CLAUDE.md): go through the owning module's service layer instead.
"""

from app.models.exchange import Order, Trade
from app.models.ledger import Account, Balance, EmissionEpoch, LedgerEntry
from app.models.relay import CapacitySource, DeliveryRecord
from app.models.user import User

__all__ = [
    "User",
    "CapacitySource",
    "DeliveryRecord",
    "Account",
    "Balance",
    "LedgerEntry",
    "EmissionEpoch",
    "Order",
    "Trade",
]
