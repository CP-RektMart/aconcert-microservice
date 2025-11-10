#!/bin/bash

echo "🔍 Verifying AConcert Kubernetes Setup"
echo "======================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

# Get script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Check Dockerfiles
echo "📦 Checking Dockerfiles..."
services=("auth" "event" "location" "gateway" "notification" "payment" "realtime" "reservation")
for service in "${services[@]}"; do
    if [ -f "$PROJECT_ROOT/services/$service/Dockerfile" ]; then
        echo -e "${GREEN}✓${NC} services/$service/Dockerfile"
    else
        echo -e "${RED}✗${NC} services/$service/Dockerfile - MISSING"
        ((ERRORS++))
    fi
done
echo ""

# Check K8s manifests
echo "📋 Checking Kubernetes Manifests..."
required_files=(
    "namespace.yaml"
    "databases/postgres.yaml"
    "databases/redis.yaml"
    "databases/mongo.yaml"
    "databases/rabbitmq.yaml"
    "services/auth-service.yaml"
    "services/event-service.yaml"
    "services/location-service.yaml"
    "services/reservation-service.yaml"
    "services/payment-service.yaml"
    "services/notification-service.yaml"
    "services/realtime-service.yaml"
    "services/gateway.yaml"
)

for file in "${required_files[@]}"; do
    if [ -f "$SCRIPT_DIR/$file" ]; then
        echo -e "${GREEN}✓${NC} .k8s/$file"
    else
        echo -e "${RED}✗${NC} .k8s/$file - MISSING"
        ((ERRORS++))
    fi
done
echo ""

# Check scripts
echo "🔧 Checking Scripts..."
scripts=("deploy.sh" "build-images.sh" "cleanup.sh")
for script in "${scripts[@]}"; do
    if [ -f "$SCRIPT_DIR/$script" ]; then
        if [ -x "$SCRIPT_DIR/$script" ]; then
            echo -e "${GREEN}✓${NC} $script (executable)"
        else
            echo -e "${YELLOW}⚠${NC} $script (not executable)"
            ((WARNINGS++))
        fi
    else
        echo -e "${RED}✗${NC} $script - MISSING"
        ((ERRORS++))
    fi
done
echo ""

# Check kubectl
echo "🔌 Checking Prerequisites..."
if command -v kubectl &> /dev/null; then
    echo -e "${GREEN}✓${NC} kubectl is installed"
    kubectl version --client --short 2>/dev/null || kubectl version --client 2>/dev/null | head -1
else
    echo -e "${YELLOW}⚠${NC} kubectl is not installed"
    ((WARNINGS++))
fi

if command -v docker &> /dev/null; then
    echo -e "${GREEN}✓${NC} docker is installed"
    docker version --format '{{.Client.Version}}' 2>/dev/null || echo "  Docker is installed"
else
    echo -e "${YELLOW}⚠${NC} docker is not installed"
    ((WARNINGS++))
fi
echo ""

# Check Kubernetes cluster
echo "☸️  Checking Kubernetes Cluster..."
if kubectl cluster-info &> /dev/null; then
    echo -e "${GREEN}✓${NC} Kubernetes cluster is accessible"
    kubectl cluster-info 2>/dev/null | head -1
else
    echo -e "${YELLOW}⚠${NC} Kubernetes cluster is not accessible"
    echo "  Make sure Docker Desktop Kubernetes or Minikube is running"
    ((WARNINGS++))
fi
echo ""

# Check for .dockerignore
echo "🐳 Checking Docker Configuration..."
if [ -f "$PROJECT_ROOT/.dockerignore" ]; then
    echo -e "${GREEN}✓${NC} .dockerignore file exists"
else
    echo -e "${YELLOW}⚠${NC} .dockerignore file not found (recommended)"
    ((WARNINGS++))
fi
echo ""

# Summary
echo "========================================"
echo "📊 Verification Summary"
echo "========================================"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed!${NC}"
    echo ""
    echo "You're ready to deploy:"
    echo "  cd $SCRIPT_DIR"
    echo "  make build   # Build Docker images"
    echo "  make deploy  # Deploy to Kubernetes"
    echo ""
    echo "Or run the automated deployment:"
    echo "  ./deploy.sh"
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  ${WARNINGS} warning(s) found${NC}"
    echo ""
    echo "You can proceed, but consider addressing the warnings."
    echo ""
    echo "To deploy:"
    echo "  cd $SCRIPT_DIR"
    echo "  make build"
    echo "  make deploy"
else
    echo -e "${RED}❌ ${ERRORS} error(s) found${NC}"
    [ $WARNINGS -gt 0 ] && echo -e "${YELLOW}⚠️  ${WARNINGS} warning(s) found${NC}"
    echo ""
    echo "Please fix the errors before deploying."
    exit 1
fi
