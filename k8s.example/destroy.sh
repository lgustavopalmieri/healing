#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Deleting healing-specialist manifests..."
kubectl delete -f "$SCRIPT_DIR/healing-specialist.yaml" --ignore-not-found

echo "Done."
