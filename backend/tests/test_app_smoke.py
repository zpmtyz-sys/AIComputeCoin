"""End-to-end smoke test: boot the full app (lifespan -> genesis + market maker),
then register -> faucet -> trade against the seeded market-maker liquidity.

Boots against the app's default SQLite DB but wipes it first for a clean, repeatable run.
"""

from pathlib import Path

import pytest

# The app's default DATABASE_URL is sqlite:///./computecoin.db relative to the CWD (repo root).
_DB_FILE = Path("computecoin.db")


@pytest.fixture(scope="module")
def client():
    if _DB_FILE.exists():
        _DB_FILE.unlink()

    from fastapi.testclient import TestClient

    from app.main import app

    with TestClient(app) as c:
        yield c

    if _DB_FILE.exists():
        _DB_FILE.unlink()


def test_health(client):
    r = client.get("/health")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"


def test_genesis_supply_present(client):
    r = client.get("/api/ledger/supply")
    assert r.status_code == 200
    body = r.json()
    assert float(body["minted"]) > 0  # genesis minted at startup


def test_market_maker_seeded_orderbook(client):
    r = client.get("/api/exchange/orderbook?pair=CC/USDC")
    assert r.status_code == 200
    book = r.json()
    # Market maker should have posted at least one ask (it holds CC).
    assert len(book["asks"]) >= 1


def test_register_faucet_and_trade(client):
    # register
    r = client.post(
        "/api/auth/register",
        json={"email": "alice@example.com", "password": "password123", "role": "user"},
    )
    assert r.status_code == 201, r.text

    # login
    r = client.post(
        "/api/auth/login",
        json={"email": "alice@example.com", "password": "password123"},
    )
    assert r.status_code == 200
    token = r.json()["access_token"]
    auth = {"Authorization": f"Bearer {token}"}

    # faucet testnet USDC
    r = client.post("/api/ledger/faucet", json={"asset": "USDC", "amount": 1000}, headers=auth)
    assert r.status_code == 200, r.text

    # buy 10 CC at a price that crosses the MM ask
    r = client.post(
        "/api/exchange/orders",
        json={"pair": "CC/USDC", "side": "buy", "type": "limit", "price": "1.05", "qty": "10"},
        headers=auth,
    )
    assert r.status_code == 201, r.text

    # alice should now hold ~10 CC
    r = client.get("/api/ledger/balance", headers=auth)
    assert r.status_code == 200
    balances = {b["asset"]: b["amount"] for b in r.json()["balances"]}
    assert float(balances.get("CC", "0")) == pytest.approx(10.0)
