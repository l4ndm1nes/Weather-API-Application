#!/bin/sh

HOST="$1"
PORT="$2"

echo "🔍 Waiting for PostgreSQL at $HOST:$PORT..."

until nc -z "$HOST" "$PORT"; do
  echo "⏳ Still waiting for PostgreSQL..."
  sleep 2
done

echo "✅ PostgreSQL is ready!"
