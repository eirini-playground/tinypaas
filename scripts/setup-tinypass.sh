#!/bin/bash

set -ex

readonly REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
readonly EIRINI_DIR="$REPO_ROOT/../eirini"
readonly DOCKERHUB_USERNAME=eiriniuser
readonly DOCKERHUB_PASSWORD="${DOCKERHUB_PASSWORD:-$(pass eirini/docker-hub)}"

ensure_kind_cluster() {
  local cluster_name
  cluster_name="$1"
  if ! kind get clusters | grep -q "$cluster_name"; then
    current_cluster="$(kubectl config current-context)" || true

    kindConfig=$(mktemp)
    trap "rm -f $kindConfig" EXIT
    cat <<EOF >>"$kindConfig"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF

    kind create cluster --name "$cluster_name" --config "$kindConfig" --wait 5m
    if [[ -n "$current_cluster" ]]; then
      kubectl config use-context "$current_cluster"
    fi
  fi
}

install_eirini() {
  pushd "$EIRINI_DIR"
  {
    scripts/skaffold run
  }
  popd
}

install_ingress_controller() {
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/kind/deploy.yaml
  kubectl wait --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=90s
}

install_tinypaas() {
  pushd "$REPO_ROOT"
  {
    make routing image-controller
    make push-routing push-image-controller
    kubectl delete -f deploy/ || true
    kubectl apply -f deploy/
  }
  popd
}

install_kpack() {
  $REPO_ROOT/install_kpack.sh
}

main() {
  ensure_kind_cluster "eirini-tinypaas"
  install_ingress_controller
  install_eirini
  install_kpack
  install_tinypaas
}

main
