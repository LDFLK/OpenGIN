#!/bin/bash
set -e

BAKED="/opt/pgdata"
LIVE="/tmp/pgdata"

echo "Copying baked data to writable directory..."
mkdir -p "$LIVE"
cp -rp "$BAKED/." "$LIVE/"
chmod 700 "$LIVE"

echo "Starting PostgreSQL..."
exec postgres -D "$LIVE" -k /tmp
