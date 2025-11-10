#!/bin/bash
set -e
NAMESPACE=aconcert

echo "💥 Deleting entire namespace $NAMESPACE..."
kubectl delete namespace $NAMESPACE --ignore-not-found

echo "🕐 Waiting for namespace deletion..."
while kubectl get namespace $NAMESPACE >/dev/null 2>&1; do
  sleep 1
done

echo "🚀 Recreating namespace and secrets..."
kubectl create namespace $NAMESPACE

for svc in auth event location reservation payment notification realtime gateway; do
  echo "📦 Creating secret for $svc ..."
  kubectl create secret generic ${svc}-env \
    --from-env-file=services/$svc/.env.k8s \
    -n $NAMESPACE
done

echo "✅ Clean namespace ready for fresh deployment!"
