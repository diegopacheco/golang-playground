#!/bin/bash

set -e

docker-compose down -v 2>/dev/null || true

docker-compose build

docker-compose up -d

echo "Waiting for services to be healthy..."
MAX_RETRIES=30
RETRY=0
while [ $RETRY -lt $MAX_RETRIES ]; do
    if docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
        echo "PostgreSQL is ready"
        break
    fi
    RETRY=$((RETRY + 1))
    sleep 1
done

if [ $RETRY -eq $MAX_RETRIES ]; then
    echo "PostgreSQL failed to start"
    exit 1
fi

RETRY=0
while [ $RETRY -lt $MAX_RETRIES ]; do
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        echo "Application is ready"
        break
    fi
    RETRY=$((RETRY + 1))
    sleep 1
done

if [ $RETRY -eq $MAX_RETRIES ]; then
    echo "Application failed to start"
    exit 1
fi

echo "All services are running"
docker-compose ps
