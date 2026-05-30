#!/usr/bin/env python3
"""Generate random secrets into .env.

- Creates .env from .env.example if it doesn't exist.
- Replaces placeholder values (those containing "change-me") for SECRET_KEY,
  POSTGRES_PASSWORD, and ADMIN_PASSWORD with strong random values.
- Never overwrites values you've already customized.
"""
from __future__ import annotations

import secrets
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ENV = ROOT / ".env"
EXAMPLE = ROOT / ".env.example"

KEYS = {
    "SECRET_KEY": lambda: secrets.token_hex(32),
    "POSTGRES_PASSWORD": lambda: secrets.token_hex(24),
    "ADMIN_PASSWORD": lambda: secrets.token_urlsafe(18),
}


def main() -> None:
    if not ENV.exists():
        if not EXAMPLE.exists():
            raise SystemExit(".env.example not found; cannot bootstrap .env")
        ENV.write_text(EXAMPLE.read_text(encoding="utf-8"), encoding="utf-8")
        print("Created .env from .env.example")

    lines = ENV.read_text(encoding="utf-8").splitlines()
    changed = []
    for i, line in enumerate(lines):
        if "=" not in line or line.lstrip().startswith("#"):
            continue
        key = line.split("=", 1)[0].strip()
        val = line.split("=", 1)[1].strip()
        if key in KEYS and ("change-me" in val or val == ""):
            lines[i] = f"{key}={KEYS[key]()}"
            changed.append(key)

    ENV.write_text("\n".join(lines) + "\n", encoding="utf-8")
    if changed:
        print("Generated secrets for:", ", ".join(changed))
    else:
        print("No placeholder secrets to replace (already customized).")
    print(f"Wrote {ENV}")


if __name__ == "__main__":
    main()
