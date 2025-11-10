#!/bin/bash

# Kubernetes Network Debugging Script
# This script helps diagnose network connectivity issues within the cluster

set -e

NAMESPACE=${1:-aconcert}

echo "========================================"
echo "Kubernetes Network Diagnostics"
echo "Namespace: $NAMESPACE"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# 1. Check if namespace exists
print_header "1. Namespace Check"
if kubectl get namespace $NAMESPACE &> /dev/null; then
    print_success "Namespace '$NAMESPACE' exists"
else
    print_error "Namespace '$NAMESPACE' does not exist!"
    exit 1
fi

# 2. List all pods and their status
print_header "2. Pod Status"
kubectl get pods -n $NAMESPACE -o wide

# 3. Check services
print_header "3. Services and Endpoints"
echo "Services:"
kubectl get svc -n $NAMESPACE
echo ""
echo "Endpoints:"
kubectl get endpoints -n $NAMESPACE

# 4. Check DNS resolution
print_header "4. DNS Resolution Test"
echo "Testing DNS from a pod..."

# Find a running pod
POD=$(kubectl get pods -n $NAMESPACE --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -z "$POD" ]; then
    print_warning "No running pods found. Deploying test pod..."

    # Create a test pod
    kubectl run network-test --image=busybox:1.35 --restart=Never -n $NAMESPACE -- sleep 3600

    echo "Waiting for test pod to be ready..."
    kubectl wait --for=condition=Ready pod/network-test -n $NAMESPACE --timeout=60s

    POD="network-test"
    TEST_POD_CREATED=true
fi

echo "Using pod: $POD"
echo ""

# Test DNS resolution
echo "Testing DNS resolution..."

# Test Kubernetes DNS
echo -n "  - kubernetes.default.svc.cluster.local: "
if kubectl exec -n $NAMESPACE $POD -- nslookup kubernetes.default.svc.cluster.local &> /dev/null; then
    print_success "OK"
else
    print_error "FAILED"
fi

# Test namespace DNS
echo -n "  - Services in namespace: "
SERVICES=$(kubectl get svc -n $NAMESPACE -o jsonpath='{.items[*].metadata.name}')
if [ -n "$SERVICES" ]; then
    for svc in $SERVICES; do
        echo -n "      $svc.$NAMESPACE.svc.cluster.local: "
        if kubectl exec -n $NAMESPACE $POD -- nslookup $svc.$NAMESPACE.svc.cluster.local &> /dev/null; then
            print_success "OK"
        else
            print_error "FAILED"
        fi
    done
else
    print_warning "No services found"
fi

# 5. Test connectivity to services
print_header "5. Service Connectivity Test"

# Check postgres databases
echo "Testing PostgreSQL databases:"
for db_svc in auth-postgres event-postgres reservation-postgres; do
    echo -n "  - $db_svc:5432: "
    if kubectl exec -n $NAMESPACE $POD -- nc -z -w 2 $db_svc.$NAMESPACE.svc.cluster.local 5432 &> /dev/null; then
        print_success "REACHABLE"
    else
        print_error "UNREACHABLE"
    fi
done

echo ""
echo "Testing Redis instances:"
for redis_svc in auth-redis event-redis reservation-redis location-redis; do
    echo -n "  - $redis_svc:6379: "
    if kubectl exec -n $NAMESPACE $POD -- nc -z -w 2 $redis_svc.$NAMESPACE.svc.cluster.local 6379 &> /dev/null; then
        print_success "REACHABLE"
    else
        print_error "UNREACHABLE"
    fi
done

echo ""
echo "Testing RabbitMQ:"
echo -n "  - rabbitmq:5672: "
if kubectl exec -n $NAMESPACE $POD -- nc -z -w 2 rabbitmq.$NAMESPACE.svc.cluster.local 5672 &> /dev/null; then
    print_success "REACHABLE"
else
    print_error "UNREACHABLE"
fi

echo ""
echo "Testing MongoDB:"
echo -n "  - location-mongo:27017: "
if kubectl exec -n $NAMESPACE $POD -- nc -z -w 2 location-mongo.$NAMESPACE.svc.cluster.local 27017 &> /dev/null; then
    print_success "REACHABLE"
else
    print_error "UNREACHABLE"
fi

echo ""
echo "Testing Application Services:"
for app_svc in event-service reservation-service location-service auth-service gateway realtime-service payment-service; do
    echo -n "  - $app_svc:8000: "
    if kubectl exec -n $NAMESPACE $POD -- nc -z -w 2 $app_svc.$NAMESPACE.svc.cluster.local 8000 &> /dev/null; then
        print_success "REACHABLE"
    else
        print_error "UNREACHABLE"
    fi
done

# 6. Check network policies
print_header "6. Network Policies"
NETPOL=$(kubectl get networkpolicies -n $NAMESPACE 2>/dev/null | tail -n +2)
if [ -z "$NETPOL" ]; then
    print_success "No network policies (all traffic allowed)"
else
    print_warning "Network policies found (may restrict traffic):"
    kubectl get networkpolicies -n $NAMESPACE
fi

# 7. Check CoreDNS
print_header "7. CoreDNS Status"
kubectl get pods -n kube-system -l k8s-app=kube-dns -o wide

echo ""
echo "CoreDNS logs (last 10 lines):"
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=10

# 8. Check pod logs for network errors
print_header "8. Recent Pod Logs (Network-related)"
echo "Checking for network/connection errors in pod logs..."
echo ""

PODS=$(kubectl get pods -n $NAMESPACE -o jsonpath='{.items[*].metadata.name}')
for pod in $PODS; do
    if [[ "$pod" != "network-test" ]]; then
        echo "Pod: $pod"
        ERROR_LOGS=$(kubectl logs $pod -n $NAMESPACE --tail=50 2>/dev/null | grep -iE "connection refused|timeout|network|dial|refused|unreachable" | tail -5 || true)
        if [ -n "$ERROR_LOGS" ]; then
            echo "$ERROR_LOGS" | sed 's/^/  /'
        else
            echo "  (no network-related errors found)"
        fi
        echo ""
    fi
done

# 9. Check init containers
print_header "9. Init Container Status"
echo "Checking init containers (dependency waiters)..."
echo ""

for pod in $PODS; do
    if [[ "$pod" != "network-test" ]]; then
        INIT_STATUS=$(kubectl get pod $pod -n $NAMESPACE -o jsonpath='{.status.initContainerStatuses[*].state}' 2>/dev/null || true)
        if [ -n "$INIT_STATUS" ]; then
            echo "Pod: $pod"
            kubectl get pod $pod -n $NAMESPACE -o jsonpath='{range .status.initContainerStatuses[*]}{.name}{": "}{.state}{"\n"}{end}' | sed 's/^/  /'
            echo ""
        fi
    fi
done

# 10. Check cluster networking components
print_header "10. Cluster Networking Components"
echo "CNI Plugin Pods:"
kubectl get pods -n kube-system -l component=kube-proxy -o wide 2>/dev/null || print_warning "kube-proxy not found"
echo ""
kubectl get pods -n kube-system | grep -iE "calico|flannel|weave|cilium|canal" || print_warning "No common CNI plugins found"

# 11. Test external connectivity
print_header "11. External Connectivity Test"
echo -n "  - Internet (google.com): "
if kubectl exec -n $NAMESPACE $POD -- nc -z -w 2 google.com 80 &> /dev/null; then
    print_success "REACHABLE"
else
    print_error "UNREACHABLE (may be expected if egress is blocked)"
fi

echo -n "  - DNS resolution (google.com): "
if kubectl exec -n $NAMESPACE $POD -- nslookup google.com &> /dev/null; then
    print_success "OK"
else
    print_error "FAILED"
fi

# 12. Summary and recommendations
print_header "12. Summary & Recommendations"

echo "Common issues and solutions:"
echo ""
echo "1. If DNS fails:"
echo "   - Check CoreDNS pods are running: kubectl get pods -n kube-system -l k8s-app=kube-dns"
echo "   - Check CoreDNS service: kubectl get svc -n kube-system kube-dns"
echo "   - Restart CoreDNS: kubectl rollout restart deployment/coredns -n kube-system"
echo ""
echo "2. If services are unreachable:"
echo "   - Check if pods are running: kubectl get pods -n $NAMESPACE"
echo "   - Check service endpoints: kubectl get endpoints -n $NAMESPACE"
echo "   - Verify service selector matches pod labels"
echo ""
echo "3. If init containers are stuck:"
echo "   - Check if database/messaging pods are ready"
echo "   - Verify service names match init container checks"
echo "   - Check logs: kubectl logs <pod> -c <init-container-name> -n $NAMESPACE"
echo ""
echo "4. If network policies block traffic:"
echo "   - Review policies: kubectl get networkpolicies -n $NAMESPACE"
echo "   - Describe policy: kubectl describe networkpolicy <name> -n $NAMESPACE"
echo ""
echo "5. To get detailed service info:"
echo "   kubectl describe svc <service-name> -n $NAMESPACE"
echo ""
echo "6. To check pod-to-pod connectivity:"
echo "   kubectl exec -it <pod1> -n $NAMESPACE -- ping <pod2-ip>"
echo ""

# Cleanup test pod if we created it
if [ "$TEST_POD_CREATED" = true ]; then
    print_header "Cleanup"
    echo "Removing test pod..."
    kubectl delete pod network-test -n $NAMESPACE --ignore-not-found=true
    print_success "Test pod removed"
fi

echo ""
print_header "Diagnostics Complete"
echo "Review the output above to identify network issues."
echo ""
