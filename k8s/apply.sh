#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Applying healing-specialist manifests..."

kubectl apply -f "$SCRIPT_DIR/namespace.yaml"
kubectl apply -f "$SCRIPT_DIR/configmap.yaml"
kubectl apply -f "$SCRIPT_DIR/networkpolicy.yaml"
kubectl apply -f "$SCRIPT_DIR/deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/service.yaml"
kubectl apply -f "$SCRIPT_DIR/pdb.yaml"
kubectl apply -f "$SCRIPT_DIR/hpa.yaml"
if kubectl api-resources --api-group=monitoring.coreos.com | grep -q ServiceMonitor; then
  kubectl apply -f "$SCRIPT_DIR/service-monitor.yaml"
else
  echo "⚠️  Skipping ServiceMonitor (Prometheus Operator CRDs not installed)"
fi

echo "✅ All manifests applied."
echo ""
kubectl get pods -n healing -l app.kubernetes.io/name=healing-specialist
