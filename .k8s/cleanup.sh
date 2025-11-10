#!/bin/bash

echo "🧹 Cleaning up AConcert Kubernetes deployment"
echo "============================================="

echo "⚠️  This will delete all resources in the 'aconcert' namespace"
read -p "Are you sure? (y/N) " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️  Deleting namespace and all resources..."
    kubectl delete namespace aconcert
    echo "✅ Cleanup complete!"
else
    echo "❌ Cleanup cancelled"
fi
