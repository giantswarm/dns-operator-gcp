#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="${SCRIPT_DIR}/.."
CLUSTER=${CLUSTER:-"dns-operator-gcp-acceptance"}
IMG=${IMG:-dns-operator-gcp:latest}

ensure_kind_cluster() {
  local cluster
  cluster="$1"
  if ! kind get clusters | grep -q "$cluster"; then
    kind create cluster --name "$cluster" --wait 5m
  fi
  kind export kubeconfig --name "$cluster" --kubeconfig "$HOME/.kube/$cluster.yml"
}

ensure_kind_cluster "$CLUSTER"
GCP_B64ENCODED_CREDENTIALS="" clusterctl init --kubeconfig "$HOME/.kube/$CLUSTER.yml" --infrastructure=gcp --wait-providers || true
kubectl create namespace giantswarm --kubeconfig "$HOME/.kube/$CLUSTER.yml" || true
kind load docker-image --name "$CLUSTER" "$IMG"
