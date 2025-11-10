# AConcert Kubernetes - Quick Start Guide

## Prerequisites
- Docker Desktop with Kubernetes enabled OR Minikube
- kubectl installed

## Quick Deploy

```bash
cd .k8s

# 1. Build Docker images
make build

# 2. Deploy everything
make deploy

# 3. Check status
make status
```

## Access Services

- **Gateway API**: http://localhost:30000
- **Realtime WebSocket**: http://localhost:30001

## Useful Commands

```bash
make status                      # Check deployment
make logs SERVICE=gateway        # View logs
make restart SERVICE=gateway     # Restart service
make clean                       # Remove everything
```

## Troubleshooting

Check pod status:
```bash
kubectl get pods -n aconcert
kubectl describe pod <pod-name> -n aconcert
kubectl logs <pod-name> -n aconcert
```

See full documentation in [README.md](README.md)
