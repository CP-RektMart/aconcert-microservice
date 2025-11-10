# AConcert Kubernetes Deployment

Complete Kubernetes manifests for deploying the AConcert microservices platform using a service discovery pattern.

## 📋 Overview

This directory contains all necessary Kubernetes manifests to deploy the AConcert concert ticketing platform on a local Kubernetes cluster. The architecture follows a **service discovery pattern** where internal services communicate using Kubernetes DNS, and only the Gateway and Realtime services are exposed publicly.

## 🏗️ Architecture

### Public-Facing Services (NodePort)
- **Gateway** (port 30000) - Main API Gateway aggregating all backend services
- **Realtime Service** (port 30001) - WebSocket service for real-time notifications

### Internal Services (ClusterIP - Service Discovery)
- **Auth Service** - User authentication and authorization
- **Event Service** - Concert event management
- **Location Service** - Venue and location management
- **Reservation Service** - Ticket reservation management
- **Payment Service** - Payment processing via Stripe
- **Notification Service** - Event-driven notification consumer (no HTTP port)

### Databases
- **PostgreSQL**: auth-postgres, event-postgres, reservation-postgres
- **Redis**: auth-redis, event-redis, reservation-redis, realtime-redis
- **MongoDB**: location-mongo
- **RabbitMQ**: Message broker for event-driven architecture

## 📁 Directory Structure

```
.k8s/
├── namespace.yaml                  # AConcert namespace definition
├── databases/                      # All database manifests
│   ├── postgres.yaml              # 3x PostgreSQL instances
│   ├── redis.yaml                 # 4x Redis instances
│   ├── mongo.yaml                 # MongoDB for locations
│   └── rabbitmq.yaml              # RabbitMQ message broker
├── services/                       # All microservice manifests
│   ├── auth-service.yaml
│   ├── event-service.yaml
│   ├── location-service.yaml
│   ├── reservation-service.yaml
│   ├── payment-service.yaml
│   ├── notification-service.yaml
│   ├── realtime-service.yaml      # PUBLIC (NodePort)
│   └── gateway.yaml               # PUBLIC (NodePort)
├── deploy.sh                       # Automated deployment script
├── build-images.sh                 # Build all Docker images
├── cleanup.sh                      # Clean up all resources
├── Makefile                        # Convenient make commands
└── README.md                       # This file
```

## 🚀 Quick Start

### Prerequisites

1. **Kubernetes Cluster** (choose one):
   - Docker Desktop with Kubernetes enabled (recommended for macOS)
   - Minikube
   - Kind (Kubernetes in Docker)

2. **kubectl** installed and configured
   ```bash
   kubectl version --client
   kubectl cluster-info
   ```

3. **Docker** installed
   ```bash
   docker version
   ```

### Step 1: Build Docker Images

```bash
cd aconcert-microservice/.k8s
make build
```

Or manually:
```bash
./build-images.sh
```

**Note**: This builds images for services that have Dockerfiles (auth, event, location, gateway). You'll need to create Dockerfiles for notification, payment, realtime, and reservation services.

### Step 2: Deploy Everything

```bash
make deploy
```

Or manually:
```bash
./deploy.sh
```

### Step 3: Verify Deployment

```bash
# Check all pods are running
make status

# Or manually
kubectl get pods -n aconcert
```

Expected output:
```
NAME                                    READY   STATUS    RESTARTS   AGE
auth-postgres-xxx                       1/1     Running   0          2m
auth-redis-xxx                          1/1     Running   0          2m
auth-service-xxx                        1/1     Running   0          1m
event-postgres-xxx                      1/1     Running   0          2m
event-service-xxx                       1/1     Running   0          1m
gateway-xxx                             1/1     Running   0          30s
location-mongo-xxx                      1/1     Running   0          2m
location-service-xxx                    1/1     Running   0          1m
notification-service-xxx                1/1     Running   0          1m
payment-service-xxx                     1/1     Running   0          1m
rabbitmq-xxx                            1/1     Running   0          2m
realtime-redis-xxx                      1/1     Running   0          2m
realtime-service-xxx                    1/1     Running   0          1m
reservation-postgres-xxx                1/1     Running   0          2m
reservation-redis-xxx                   1/1     Running   0          2m
reservation-service-xxx                 1/1     Running   0          1m
```

### Step 4: Access Services

**Gateway API:**
```bash
curl http://localhost:30000/health
```

**Realtime WebSocket:**
```bash
# Using wscat (install: npm install -g wscat)
wscat -c ws://localhost:30001
```

## 📚 Available Commands

### Using Makefile

```bash
make help                           # Show all available commands
make build                          # Build all Docker images
make deploy                         # Full deployment (recommended)
make deploy-databases               # Deploy only databases
make deploy-services                # Deploy only services
make deploy-gateway                 # Deploy only gateway
make status                         # Show deployment status
make logs SERVICE=gateway           # View logs for a service
make restart SERVICE=gateway        # Restart a service
make scale SERVICE=gateway REPLICAS=2  # Scale a service
make clean                          # Clean up everything
```

### Manual Commands

```bash
# Deploy everything
kubectl apply -f namespace.yaml
kubectl apply -f databases/
kubectl apply -f services/

# Check pod status
kubectl get pods -n aconcert

# Watch pods in real-time
kubectl get pods -n aconcert -w

# Check services
kubectl get svc -n aconcert

# View logs
kubectl logs -f deployment/gateway -n aconcert
kubectl logs -f deployment/auth-service -n aconcert

# Describe a pod
kubectl describe pod <pod-name> -n aconcert

# Get into a pod shell
kubectl exec -it <pod-name> -n aconcert -- sh

# Port forward (alternative to NodePort)
kubectl port-forward -n aconcert svc/gateway 8000:8000
```

## 🔧 Configuration

### Service Discovery

Internal services communicate using Kubernetes DNS:
```
<service-name>.aconcert.svc.cluster.local:<port>
```

Examples:
- `auth-service.aconcert.svc.cluster.local:8000`
- `event-service.aconcert.svc.cluster.local:8000`
- `location-service.aconcert.svc.cluster.local:8000`
- `reservation-service.aconcert.svc.cluster.local:8000`
- `payment-service.aconcert.svc.cluster.local:8000`
- `realtime-service.aconcert.svc.cluster.local:8000`

### Updating Configuration

To update environment variables, edit the ConfigMaps in each service YAML file:

```bash
# Edit gateway configuration
kubectl edit configmap gateway-config -n aconcert

# Restart deployment to pick up changes
kubectl rollout restart deployment/gateway -n aconcert
```

### Important Configuration Items

Before deploying to production, update these values:

1. **JWT Secret** (in `services/gateway.yaml` and `services/auth-service.yaml`):
   ```yaml
   JWT_SECRET: "your-secure-jwt-secret-here"
   ```

2. **Stripe Keys** (in `services/payment-service.yaml` and `services/reservation-service.yaml`):
   ```yaml
   STRIPE_SECRET_KEY: "sk_live_your_real_stripe_key"
   ```

3. **Database Passwords** (in `databases/*.yaml`):
   ```yaml
   POSTGRES_PASSWORD: "strong-password-here"
   MONGO_INITDB_ROOT_PASSWORD: "strong-password-here"
   ```

## 💾 Resource Requirements

Optimized for **8GB RAM** systems:

| Component | Memory Request | Memory Limit | CPU Request | CPU Limit |
|-----------|----------------|--------------|-------------|-----------|
| Backend Services | 128Mi | 256Mi | 100m | 200m |
| PostgreSQL | 128Mi | 256Mi | 100m | 200m |
| MongoDB | 256Mi | 512Mi | 100m | 200m |
| Redis | 64Mi | 128Mi | 50m | 100m |
| RabbitMQ | 256Mi | 512Mi | 100m | 200m |

**Total Estimated Usage**: ~3-4GB RAM

## 🔍 Monitoring & Debugging

### View Logs

```bash
# Follow logs for a specific service
make logs SERVICE=gateway

# View all logs for a service
kubectl logs deployment/gateway -n aconcert

# View logs from all pods with a label
kubectl logs -l app=auth-service -n aconcert

# View logs from init containers
kubectl logs <pod-name> -c wait-for-postgres -n aconcert
```

### Check Pod Status

```bash
# Get detailed pod information
kubectl describe pod <pod-name> -n aconcert

# Get pod events
kubectl get events -n aconcert --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n aconcert
kubectl top nodes
```

### Common Issues

#### Pods stuck in Pending
```bash
# Check pod details
kubectl describe pod <pod-name> -n aconcert

# Common causes:
# - Insufficient resources (check: kubectl top nodes)
# - PVC not bound (check: kubectl get pvc -n aconcert)
```

#### Pods in CrashLoopBackOff
```bash
# Check logs
kubectl logs <pod-name> -n aconcert

# Check previous logs
kubectl logs <pod-name> -n aconcert --previous

# Common causes:
# - Database connection issues
# - Missing environment variables
# - Application errors
```

#### Service cannot connect to database
```bash
# Test connectivity from within the cluster
kubectl run -it --rm debug --image=busybox --restart=Never -n aconcert -- sh

# Inside the pod, test connections:
nc -zv auth-postgres.aconcert.svc.cluster.local 5432
nc -zv auth-redis.aconcert.svc.cluster.local 6379
nc -zv location-mongo.aconcert.svc.cluster.local 27017
nc -zv rabbitmq.aconcert.svc.cluster.local 5672
```

#### Image Pull Errors
```bash
# For Docker Desktop: Images should be available automatically

# For Minikube: Load images into minikube
minikube image load aconcert/gateway:latest
minikube image load aconcert/auth-service:latest
# ... repeat for all services
```

## 🔄 Maintenance

### Restart a Service

```bash
make restart SERVICE=gateway

# Or manually
kubectl rollout restart deployment/gateway -n aconcert
```

### Scale a Service

```bash
make scale SERVICE=gateway REPLICAS=2

# Or manually
kubectl scale deployment/gateway --replicas=2 -n aconcert
```

### Update a Service

```bash
# Update the image
kubectl set image deployment/gateway gateway=aconcert/gateway:v2 -n aconcert

# Or edit the deployment
kubectl edit deployment gateway -n aconcert
```

### Rollback a Deployment

```bash
# View rollout history
kubectl rollout history deployment/gateway -n aconcert

# Rollback to previous version
kubectl rollout undo deployment/gateway -n aconcert

# Rollback to specific revision
kubectl rollout undo deployment/gateway --to-revision=2 -n aconcert
```

## 🧹 Cleanup

### Remove Everything

```bash
make clean

# Or manually
kubectl delete namespace aconcert
```

### Remove Specific Resources

```bash
# Remove a specific service
kubectl delete -f services/payment-service.yaml

# Remove all services
kubectl delete -f services/

# Remove all databases
kubectl delete -f databases/
```

## 🔐 Security Considerations

For production deployments:

1. **Use Secrets instead of ConfigMaps** for sensitive data:
   ```bash
   kubectl create secret generic db-credentials \
     --from-literal=username=postgres \
     --from-literal=password=strong-password \
     -n aconcert
   ```

2. **Enable RBAC** (Role-Based Access Control)

3. **Use Network Policies** to restrict pod-to-pod communication

4. **Enable TLS/SSL** for external traffic (use Ingress with cert-manager)

5. **Regular security updates** for base images

6. **Use private container registry** for production images

## 📈 Next Steps

1. **Add Ingress Controller** for better routing and TLS termination
2. **Implement Horizontal Pod Autoscaler (HPA)** for auto-scaling
3. **Set up Monitoring** with Prometheus and Grafana
4. **Add Logging** with ELK stack or Loki
5. **Implement CI/CD** pipeline with GitHub Actions or GitLab CI
6. **Add Health Checks** in all services at `/health` endpoint
7. **Create Dockerfiles** for missing services (notification, payment, realtime, reservation)

## 🆘 Support

For issues or questions:

1. Check pod logs: `kubectl logs <pod-name> -n aconcert`
2. Check events: `kubectl get events -n aconcert`
3. Review this README's troubleshooting section
4. Check Kubernetes documentation: https://kubernetes.io/docs/

## 📝 Notes

- All services listen on port 8000 internally
- Gateway is exposed on NodePort 30000
- Realtime service is exposed on NodePort 30001
- Database passwords are set to "password" (change for production!)
- All services are in the `aconcert` namespace
- PersistentVolumeClaims use default storage class (1Gi each)
