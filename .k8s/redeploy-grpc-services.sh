#!/bin/bash

# Script to rebuild and redeploy gRPC services with health checks
# Services: event, reservation, location

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🚀 Rebuilding and redeploying gRPC services with health checks"
echo "================================================================"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if namespace exists
print_status "Checking if namespace 'aconcert' exists..."
if ! kubectl get namespace aconcert &> /dev/null; then
    print_warning "Namespace 'aconcert' does not exist. Creating it..."
    kubectl apply -f "$SCRIPT_DIR/namespace.yaml"
    print_success "Namespace created"
else
    print_success "Namespace exists"
fi

# Array of services to update
SERVICES=("event" "reservation" "location")

echo ""
print_status "Services to rebuild and redeploy:"
for service in "${SERVICES[@]}"; do
    echo "  - ${service}-service"
done

echo ""
read -p "Continue with rebuild and deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Deployment cancelled"
    exit 0
fi

# Step 1: Build Docker images
echo ""
echo "================================================"
print_status "Step 1: Building Docker images"
echo "================================================"

for service in "${SERVICES[@]}"; do
    print_status "Building ${service}-service..."

    DOCKERFILE="$PROJECT_ROOT/services/$service/Dockerfile"
    IMAGE_NAME="aconcert/${service}-service:latest"

    if [ ! -f "$DOCKERFILE" ]; then
        print_error "Dockerfile not found at $DOCKERFILE"
        exit 1
    fi

    if docker build -t "$IMAGE_NAME" -f "$DOCKERFILE" "$PROJECT_ROOT"; then
        print_success "Built $IMAGE_NAME"
    else
        print_error "Failed to build $IMAGE_NAME"
        exit 1
    fi
done

# Step 2: Delete old deployments
echo ""
echo "================================================"
print_status "Step 2: Deleting old deployments"
echo "================================================"

for service in "${SERVICES[@]}"; do
    print_status "Deleting ${service}-service deployment..."

    if kubectl delete deployment "${service}-service" -n aconcert --ignore-not-found=true; then
        print_success "Deleted ${service}-service deployment"
    else
        print_warning "No existing deployment for ${service}-service"
    fi
done

# Give Kubernetes time to clean up
print_status "Waiting for pods to terminate..."
sleep 5

# Step 3: Apply new deployments
echo ""
echo "================================================"
print_status "Step 3: Deploying updated services"
echo "================================================"

for service in "${SERVICES[@]}"; do
    print_status "Deploying ${service}-service..."

    MANIFEST="$SCRIPT_DIR/services/${service}-service.yaml"

    if [ ! -f "$MANIFEST" ]; then
        print_error "Manifest not found at $MANIFEST"
        exit 1
    fi

    if kubectl apply -f "$MANIFEST"; then
        print_success "Applied ${service}-service manifest"
    else
        print_error "Failed to apply ${service}-service manifest"
        exit 1
    fi
done

# Step 4: Wait for deployments to be ready
echo ""
echo "================================================"
print_status "Step 4: Waiting for deployments to be ready"
echo "================================================"

for service in "${SERVICES[@]}"; do
    print_status "Waiting for ${service}-service..."

    if kubectl wait --for=condition=available --timeout=120s \
        deployment/"${service}-service" -n aconcert 2>/dev/null; then
        print_success "${service}-service is ready"
    else
        print_warning "${service}-service did not become ready within 120s"
        print_warning "Check logs: kubectl logs -l app=${service}-service -n aconcert"
    fi
done

# Step 5: Verify deployments
echo ""
echo "================================================"
print_status "Step 5: Verification"
echo "================================================"

print_status "Pod status:"
kubectl get pods -n aconcert | grep -E "NAME|event-service|reservation-service|location-service"

echo ""
print_status "Service endpoints:"
kubectl get endpoints -n aconcert | grep -E "NAME|event-service|reservation-service|location-service"

# Check for unhealthy events
echo ""
print_status "Checking for probe failures in the last 5 minutes..."
UNHEALTHY_EVENTS=$(kubectl get events -n aconcert \
    --field-selector reason=Unhealthy \
    --sort-by='.lastTimestamp' 2>/dev/null | tail -n +2)

if [ -z "$UNHEALTHY_EVENTS" ]; then
    print_success "No unhealthy probe events found!"
else
    print_warning "Found some unhealthy probe events:"
    echo "$UNHEALTHY_EVENTS"
fi

# Summary
echo ""
echo "================================================================"
print_success "Deployment complete!"
echo "================================================================"
echo ""
echo "Next steps:"
echo "  1. Monitor pod status:"
echo "     kubectl get pods -n aconcert -w"
echo ""
echo "  2. Check logs for any service:"
echo "     kubectl logs -l app=event-service -n aconcert --tail=50"
echo "     kubectl logs -l app=reservation-service -n aconcert --tail=50"
echo "     kubectl logs -l app=location-service -n aconcert --tail=50"
echo ""
echo "  3. Test gRPC health check (port-forward first):"
echo "     kubectl port-forward svc/event-service 8000:8000 -n aconcert"
echo "     grpcurl -plaintext localhost:8000 grpc.health.v1.Health/Check"
echo ""
echo "  4. Check for probe failures:"
echo "     kubectl get events -n aconcert --field-selector reason=Unhealthy"
echo ""
echo "See GRPC_HEALTH_CHECK_GUIDE.md for more details"
