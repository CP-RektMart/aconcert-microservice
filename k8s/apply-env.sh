#!/bin/bash
set -e
NAMESPACE=aconcert

echo "🚀 Creating namespace (if not exists)..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

for svc in auth event location reservation payment notification realtime gateway; do
  echo "📦 Creating secret for $svc ..."
  kubectl delete secret ${svc}-env -n $NAMESPACE --ignore-not-found
  kubectl create secret generic ${svc}-env \
    --from-env-file=services/$svc/.env.k8s \
    -n $NAMESPACE
done

echo "✅ All .env.k8s files have been imported as Secrets!"
