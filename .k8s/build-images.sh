#!/bin/bash

set -e

echo "🔨 Building Docker images for AConcert services"
echo "==============================================="

cd ..

services=(
    "auth"
    "event"
    "location"
    "gateway"
    "notification"
    "payment"
    "realtime"
    "reservation"
)

for service in "${services[@]}"; do
    echo ""
    echo "🔨 Building $service-service..."
    if [ -f "services/$service/Dockerfile" ]; then
        docker build -t aconcert/$service-service:latest -f services/$service/Dockerfile .
        echo "✅ $service-service built successfully"
    else
        echo "⚠️  Dockerfile not found for $service-service, skipping..."
    fi
done

echo ""
echo "✅ All images built successfully!"
echo ""
echo "📦 Built images:"
docker images | grep aconcert
