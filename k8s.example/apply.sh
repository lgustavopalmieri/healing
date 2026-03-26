#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Applying healing-specialist manifests..."
kubectl apply -f "$SCRIPT_DIR/healing-specialist.yaml"

echo "Done."
kubectl get pods -n healing -l app=healing-specialist
