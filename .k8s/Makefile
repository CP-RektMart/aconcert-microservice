.PHONY: help build deploy deploy-databases deploy-services deploy-gateway status logs clean

help: ## Show this help
	@echo "AConcert Kubernetes Deployment"
	@echo "=============================="
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## Build all Docker images
	@chmod +x build-images.sh
	@./build-images.sh

deploy-databases: ## Deploy only databases
	@echo "Deploying databases..."
	@kubectl apply -f namespace.yaml
	@kubectl apply -f databases/

deploy-services: ## Deploy only services
	@echo "Deploying services..."
	@kubectl apply -f services/

deploy: ## Full deployment (databases + services)
	@chmod +x deploy.sh
	@./deploy.sh

deploy-gateway: ## Deploy only gateway
	@kubectl apply -f services/gateway.yaml

status: ## Show deployment status
	@echo "Namespace status:"
	@kubectl get all -n aconcert
	@echo ""
	@echo "Pod details:"
	@kubectl get pods -n aconcert -o wide

logs: ## Show logs (usage: make logs SERVICE=gateway)
	@kubectl logs -f deployment/$(SERVICE) -n aconcert

clean: ## Clean up all resources
	@chmod +x cleanup.sh
	@./cleanup.sh

restart: ## Restart a service (usage: make restart SERVICE=gateway)
	@kubectl rollout restart deployment/$(SERVICE) -n aconcert

scale: ## Scale a service (usage: make scale SERVICE=gateway REPLICAS=2)
	@kubectl scale deployment/$(SERVICE) --replicas=$(REPLICAS) -n aconcert
