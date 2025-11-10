# Docker Build Guide for AConcert Microservices

## Overview

All microservices follow the same Dockerfile pattern for consistency and optimization.

## Dockerfile Structure

Each service uses a multi-stage build pattern:

```dockerfile
FROM golang:1.25.0-alpine3.21

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
COPY services/<service>/go.mod services/<service>/go.sum ./services/<service>/
COPY pkg ./pkg/

# Download dependencies
RUN cd services/<service> && go mod download

# Copy source code
COPY services/<service> ./services/<service>/

# Build the application
RUN cd services/<service> && CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/main.go

# Expose port (if HTTP service)
EXPOSE 8080

# Run the application
CMD ["/app/app"]
```

## Building Individual Services

### Build a Single Service

```bash
# From project root
docker build -t aconcert/<service>-service:latest -f services/<service>/Dockerfile .

# Examples:
docker build -t aconcert/auth-service:latest -f services/auth/Dockerfile .
docker build -t aconcert/gateway:latest -f services/gateway/Dockerfile .
docker build -t aconcert/realtime-service:latest -f services/realtime/Dockerfile .
```

### Build All Services

```bash
# From .k8s directory
cd .k8s
make build

# Or run the script directly
./build-images.sh
```

## Service List

All services with Dockerfiles:

1. **auth-service** - Authentication & authorization
2. **event-service** - Event management
3. **location-service** - Location/venue management
4. **gateway** - API Gateway
5. **notification-service** - Notification consumer
6. **payment-service** - Payment processing
7. **realtime-service** - WebSocket/real-time updates
8. **reservation-service** - Reservation management

## Build Optimization

### .dockerignore

The `.dockerignore` file excludes:
- Git files and history
- Documentation files
- IDE/editor files
- Test files and coverage reports
- Temporary files
- Kubernetes manifests
- CI/CD files
- Development files (.env, etc.)

This significantly reduces build context and speeds up builds.

### Build Arguments

You can customize builds with arguments:

```bash
# Build with specific Go version
docker build \
  --build-arg GO_VERSION=1.25.0 \
  -t aconcert/auth-service:latest \
  -f services/auth/Dockerfile .

# Build with specific tag
docker build \
  -t aconcert/auth-service:v1.0.0 \
  -f services/auth/Dockerfile .
```

## Tagging Strategy

### Development
```bash
docker build -t aconcert/auth-service:dev -f services/auth/Dockerfile .
```

### Staging
```bash
docker build -t aconcert/auth-service:staging -f services/auth/Dockerfile .
```

### Production
```bash
docker build -t aconcert/auth-service:v1.0.0 -f services/auth/Dockerfile .
docker tag aconcert/auth-service:v1.0.0 aconcert/auth-service:latest
```

## Pushing to Registry

### Docker Hub
```bash
# Login
docker login

# Tag
docker tag aconcert/auth-service:latest yourusername/auth-service:latest

# Push
docker push yourusername/auth-service:latest
```

### Private Registry
```bash
# Tag with registry URL
docker tag aconcert/auth-service:latest registry.company.com/aconcert/auth-service:latest

# Push
docker push registry.company.com/aconcert/auth-service:latest
```

## For Local Kubernetes

### Docker Desktop
Images are automatically available to Kubernetes.

### Minikube
Load images into Minikube:

```bash
# Point to Minikube's Docker daemon
eval $(minikube docker-env)

# Build images (they'll be available in Minikube)
make build

# Or load pre-built images
minikube image load aconcert/auth-service:latest
minikube image load aconcert/gateway:latest
# ... repeat for all services
```

### Kind (Kubernetes in Docker)
```bash
# Load images into Kind cluster
kind load docker-image aconcert/auth-service:latest --name aconcert-cluster
kind load docker-image aconcert/gateway:latest --name aconcert-cluster
# ... repeat for all services
```

## Build Troubleshooting

### Issue: "go.mod file not found"
**Solution**: Ensure you're running docker build from the project root, not from the service directory.

```bash
# Wrong
cd services/auth
docker build -t aconcert/auth-service .

# Correct
docker build -t aconcert/auth-service -f services/auth/Dockerfile .
```

### Issue: "Cannot download dependencies"
**Solution**: Check network connectivity and Go module proxy settings.

```bash
# Build with proxy
docker build \
  --build-arg GOPROXY=https://proxy.golang.org,direct \
  -t aconcert/auth-service:latest \
  -f services/auth/Dockerfile .
```

### Issue: "Build context too large"
**Solution**: Ensure `.dockerignore` is present and properly configured.

```bash
# Check what's being sent to Docker daemon
docker build --no-cache --progress=plain -f services/auth/Dockerfile . 2>&1 | head -20
```

### Issue: "Out of disk space"
**Solution**: Clean up old images and build cache.

```bash
# Remove unused images
docker image prune -a

# Remove build cache
docker builder prune -a

# Complete cleanup
docker system prune -a --volumes
```

## Best Practices

1. **Always build from project root**: Docker context needs access to `pkg/` directory

2. **Use specific tags**: Don't rely only on `latest` for production

3. **Multi-stage builds**: Consider using multi-stage builds for smaller final images

4. **Layer caching**: Order Dockerfile commands to maximize cache hits

5. **Security scanning**: Scan images for vulnerabilities
   ```bash
   docker scan aconcert/auth-service:latest
   ```

6. **Image size**: Keep images small
   ```bash
   docker images | grep aconcert
   ```

## Quick Commands Reference

```bash
# Build all services
make build

# Build specific service
docker build -t aconcert/auth-service:latest -f services/auth/Dockerfile .

# List built images
docker images | grep aconcert

# Remove all aconcert images
docker rmi $(docker images | grep aconcert | awk '{print $3}')

# Build with no cache
docker build --no-cache -t aconcert/auth-service:latest -f services/auth/Dockerfile .

# Check image size
docker images aconcert/auth-service:latest

# Inspect image layers
docker history aconcert/auth-service:latest
```

## Next Steps

After building images:
1. Verify images are built: `docker images | grep aconcert`
2. Deploy to Kubernetes: `make deploy`
3. Check deployment: `make status`
