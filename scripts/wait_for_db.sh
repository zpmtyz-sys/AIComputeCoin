#!/usr/bin/env bash
# wait_for_db.sh — wait for PostgreSQL to become reachable before starting the app.
# Used as a docker-compose entrypoint wrapper when `depends_on: condition` is insufficient.
set -e

MAX_RETRIES=${MAX_RETRIES:-30}
RETRY_INTERVAL=${RETRY_INTERVAL:-2}

echo "[wait_for_db] Waiting for database..."
for i in $(seq 1 $MAX_RETRIES); do
  if python -c "
import sys, asyncio, asyncpg
async def check():
    try:
        import os
        url = os.environ.get('DATABASE_URL','')
        if 'sqlite' in url:
            return True
        # Parse asyncpg URL from SQLAlchemy format
        url = url.replace('postgresql+asyncpg://', 'postgresql://')
        conn = await asyncpg.connect(url, timeout=3)
        await conn.close()
        return True
    except Exception:
        return False
sys.exit(0 if asyncio.run(check()) else 1)
" 2>/dev/null; then
    echo "[wait_for_db] Database is ready."
    exec "$@"
  fi
  echo "[wait_for_db] Attempt $i/$MAX_RETRIES — retrying in ${RETRY_INTERVAL}s..."
  sleep "$RETRY_INTERVAL"
done

echo "[wait_for_db] ERROR: Database not reachable after $MAX_RETRIES attempts."
exit 1
