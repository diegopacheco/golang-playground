#!/bin/bash

set -e

echo "Stopping all services..."
docker-compose down -v

echo "All services stopped and volumes removed"
