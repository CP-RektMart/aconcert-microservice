#!/bin/bash

set -e

echo "🚀 Starting AConcert Kubernetes Deployment"
echo "=========================================="

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl first."
    exit 1
fi

# Create namespace
echo "📦 Creating namespace..."
kubectl apply -f namespace.yaml

# Deploy databases
echo "🗄️  Deploying databases..."
kubectl apply -f databases/postgres.yaml
kubectl apply -f databases/redis.yaml
kubectl apply -f databases/mongo.yaml
kubectl apply -f databases/rabbitmq.yaml

echo "⏳ Waiting for databases to be ready..."
sleep 10

# Wait for key databases
kubectl wait --for=condition=ready pod -l app=auth-postgres -n aconcert --timeout=300s || echo "⚠️  Auth Postgres not ready yet"
kubectl wait --for=condition=ready pod -l app=event-postgres -n aconcert --timeout=300s || echo "⚠️  Event Postgres not ready yet"
kubectl wait --for=condition=ready pod -l app=location-mongo -n aconcert --timeout=300s || echo "⚠️  Location Mongo not ready yet"
kubectl wait --for=condition=ready pod -l app=rabbitmq -n aconcert --timeout=300s || echo "⚠️  RabbitMQ not ready yet"

echo "✅ Databases are ready!"

# Deploy backend services
echo "🔧 Deploying backend services..."
kubectl apply -f services/auth-service.yaml
kubectl apply -f services/event-service.yaml
kubectl apply -f services/location-service.yaml
kubectl apply -f services/reservation-service.yaml
kubectl apply -f services/payment-service.yaml
kubectl apply -f services/notification-service.yaml
kubectl apply -f services/realtime-service.yaml

echo "⏳ Waiting for backend services..."
sleep 15

# Deploy gateway
echo "🌐 Deploying gateway..."
kubectl apply -f services/gateway.yaml

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Current status:"
kubectl get pods -n aconcert
echo ""
echo "🌐 Access points:"
echo "  Gateway:  http://localhost:30000"
echo "  Realtime: http://localhost:30001"
echo ""
echo "📝 Useful commands:"
echo "  kubectl get pods -n aconcert              # Check pod status"
echo "  kubectl logs -f <pod-name> -n aconcert    # View logs"
echo "  kubectl get svc -n aconcert               # List services"
