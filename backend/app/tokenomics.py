"""ComputeCoin tokenomics — the SINGLE SOURCE OF TRUTH for all economic constants.

Human-readable counterpart: docs/tokenomics.md (keep them in sync).

This module is part of the HIGH-RISK SHARED ZONE (CLAUDE.md). Changing anything here
changes the protocol. Always run `make test` (emission/ledger/cu invariants) after edits.

Design principles (compliance, see docs/compliance.md):
  * Everything is transparent and hard-coded here — no hidden mint switches.
  * All amounts are Decimal, never float (invariant #6).
  * Emission is a deterministic pure function of the elapsed halving period (invariant #3).
"""

from __future__ import annotations

from dataclasses import dataclass
from decimal import ROUND_DOWN, Decimal

# ---------------------------------------------------------------------------
# Token basics
# ---------------------------------------------------------------------------
SYMBOL = "CC"
DECIMALS = 8
_QUANT = Decimal(1).scaleb(-DECIMALS)  # 1e-8

# Hard cap (21 亿). No mint may ever push circulating + locked above this (invariant #2).
MAX_SUPPLY = Decimal("2100000000")

# Genesis: minted once at launch, fully transparent, split across the buckets below.
GENESIS_SUPPLY = Decimal("210000000")  # 10% of MAX_SUPPLY

# Emission pool = what remains, mined over time by suppliers via Proof-of-Delivered-Compute.
EMISSION_POOL = MAX_SUPPLY - GENESIS_SUPPLY  # 1,890,000,000


@dataclass(frozen=True)
class GenesisBucket:
    """A transparent genesis allocation bucket.

    fraction:    share of GENESIS_SUPPLY (all buckets must sum to 1).
    cliff_days:  no release before this many days after genesis.
    vesting_days:linear release window after the cliff (0 = unlocked immediately).
    note:        human explanation.
    """

    name: str
    fraction: Decimal
    cliff_days: int
    vesting_days: int
    note: str


# Sum of fractions MUST equal 1 (validated by tests). No hidden founder bucket:
# the founder's tokens are the publicly-disclosed, cliff-locked portion of `team`.
GENESIS_ALLOCATIONS: tuple[GenesisBucket, ...] = (
    GenesisBucket("ecosystem", Decimal("0.30"), 0, 0, "生态与流动性:做市/上市流动性"),
    GenesisBucket("supplier_reserve", Decimal("0.25"), 0, 1080, "供给激励储备,随时间释放"),
    GenesisBucket("team", Decimal("0.18"), 365, 1080, "团队/创始人(公开披露,12个月cliff)"),
    GenesisBucket("investors", Decimal("0.15"), 180, 720, "投资人(正规融资)"),
    GenesisBucket("treasury", Decimal("0.10"), 0, 0, "国库(治理多签支配)"),
    GenesisBucket("community", Decimal("0.02"), 0, 90, "社区与空投"),
)

# ---------------------------------------------------------------------------
# Emission & halving (deterministic)
# ---------------------------------------------------------------------------
INITIAL_DAILY_EMISSION = Decimal("1000000")  # 1,000,000 CC/day at period 0
HALVING_PERIOD_DAYS = 540  # ~18 months


def halving_index(days_elapsed: int) -> int:
    """Which halving period a given day belongs to. Pure function."""
    if days_elapsed < 0:
        raise ValueError("days_elapsed must be >= 0")
    return days_elapsed // HALVING_PERIOD_DAYS


def daily_emission(period: int) -> Decimal:
    """Base CC emitted per day for a halving period. Deterministic (invariant #3).

    period 0 -> INITIAL, period 1 -> INITIAL/2, period 2 -> INITIAL/4 ...
    """
    if period < 0:
        raise ValueError("period must be >= 0")
    return quantize(INITIAL_DAILY_EMISSION / (Decimal(2) ** period))


# ---------------------------------------------------------------------------
# Fees (route to the transparent treasury account)
# ---------------------------------------------------------------------------
TAKER_FEE = Decimal("0.001")  # 0.10%
MAKER_FEE = Decimal("0.0002")  # 0.02%

# ---------------------------------------------------------------------------
# CU (Compute Unit) standardization  ——  1 CU = 1 TFLOPS * 1 hour of FP16-equivalent compute
# ---------------------------------------------------------------------------
# FP16 TFLOPS reference for common accelerators (dense, vendor spec ballpark; tune freely).
GPU_TFLOPS_FP16: dict[str, Decimal] = {
    "H100": Decimal("989"),
    "H200": Decimal("989"),
    "A100": Decimal("312"),
    "A800": Decimal("312"),
    "L40S": Decimal("362"),
    "RTX4090": Decimal("165"),
    "RTX3090": Decimal("71"),
    "MI300X": Decimal("1307"),
    "MI250X": Decimal("383"),
    "TPUv5e": Decimal("197"),
}

# Token-delivery -> CU. Coefficient = CU credited per 1,000 delivered tokens for a model class.
# These are tunable economic parameters, not physical constants.
TOKEN_TO_CU: dict[str, Decimal] = {
    "default": Decimal("0.010"),
    "claude-3.5-sonnet": Decimal("0.020"),
    "claude-3-opus": Decimal("0.030"),
    "gpt-4o": Decimal("0.018"),
    "gpt-4-turbo": Decimal("0.025"),
    "gemini-1.5-pro": Decimal("0.016"),
    "gemini-1.5-flash": Decimal("0.006"),
    "deepseek-v3": Decimal("0.008"),
}

INTERCONNECT_FACTOR = {"nvlink": Decimal("1.15"), "pcie": Decimal("1.0")}


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
def quantize(amount: Decimal) -> Decimal:
    """Round DOWN to token precision. Rounding down protects the supply cap."""
    return amount.quantize(_QUANT, rounding=ROUND_DOWN)


def genesis_amount(bucket: GenesisBucket) -> Decimal:
    return quantize(GENESIS_SUPPLY * bucket.fraction)
