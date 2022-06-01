#!/bin/bash

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
readonly REPO_ROOT="${SCRIPT_DIR}/.."
readonly CLUSTER=${CLUSTER:-"dns-operator-gcp-acceptance"}
readonly KIND="${REPO_ROOT}/bin/kind"
readonly IMG=${IMG:-quay.io/giantswarm/dns-operator-gcp:latest}

ensure_kind_cluster() {
  local cluster
  cluster="$1"
  if ! "$KIND" get clusters | grep -q "$cluster"; then
    "$KIND" create cluster --name "$cluster" --wait 5m
  fi
  "$KIND" export kubeconfig --name "$cluster" --kubeconfig "$HOME/.kube/$cluster.yml"
}

ensure_kind_cluster "$CLUSTER"
kubectl create namespace giantswarm --kubeconfig "$HOME/.kube/$CLUSTER.yml" || true
kubectl apply -f "${SCRIPT_DIR}/assets/ingress-service.yaml"
"$KIND" load docker-image --name "$CLUSTER" "$IMG"
