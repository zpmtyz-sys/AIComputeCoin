"""Tokenomics invariants: transparent genesis + deterministic halving."""

from decimal import Decimal

from app import tokenomics as tk


def test_genesis_fractions_sum_to_one():
    total = sum((b.fraction for b in tk.GENESIS_ALLOCATIONS), Decimal("0"))
    assert total == Decimal("1")


def test_genesis_amounts_sum_to_genesis_supply():
    total = sum((tk.genesis_amount(b) for b in tk.GENESIS_ALLOCATIONS), Decimal("0"))
    # Rounding-down per bucket may leave a tiny dust below GENESIS_SUPPLY; never above.
    assert total <= tk.GENESIS_SUPPLY
    assert tk.GENESIS_SUPPLY - total < Decimal("1")


def test_team_bucket_has_cliff_and_vesting():
    team = next(b for b in tk.GENESIS_ALLOCATIONS if b.name == "team")
    assert team.cliff_days > 0
    assert team.vesting_days > 0


def test_daily_emission_halving():
    assert tk.daily_emission(0) == tk.INITIAL_DAILY_EMISSION
    assert tk.daily_emission(1) == tk.quantize(tk.INITIAL_DAILY_EMISSION / 2)
    assert tk.daily_emission(2) == tk.quantize(tk.INITIAL_DAILY_EMISSION / 4)


def test_daily_emission_deterministic():
    assert tk.daily_emission(5) == tk.daily_emission(5)


def test_halving_index_boundaries():
    p = tk.HALVING_PERIOD_DAYS
    assert tk.halving_index(0) == 0
    assert tk.halving_index(p - 1) == 0
    assert tk.halving_index(p) == 1
    assert tk.halving_index(2 * p) == 2


def test_genesis_below_cap():
    assert tk.GENESIS_SUPPLY < tk.MAX_SUPPLY
    assert tk.EMISSION_POOL == tk.MAX_SUPPLY - tk.GENESIS_SUPPLY
