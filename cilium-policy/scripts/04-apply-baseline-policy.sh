#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=_lib.sh
source "${SCRIPT_DIR}/_lib.sh"
use_demo_context

kubectl apply -f "${DEMO_DIR}/manifests/provider-egress.yaml"
kubectl apply -f "${DEMO_DIR}/manifests/default-deny-egress.yaml"
kubectl -n applications get ciliumnetworkpolicy default-deny-application-egress

echo "Application-to-application traffic is now default-denied."
echo "Only the supplied provider plumbing and cluster DNS remain allowed."
echo "Run ./scripts/05-smoke-test.sh blocked to observe the expected failure."
