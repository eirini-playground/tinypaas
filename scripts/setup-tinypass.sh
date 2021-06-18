#!/bin/bash

set -ex

readonly REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
readonly EIRINI_DIR="$REPO_ROOT/../eirini"
readonly EIRINI_RELEASE_DIR="$REPO_ROOT/../eirini-release"
readonly DOCKERHUB_USERNAME=eiriniuser
readonly DOCKERHUB_PASSWORD="${DOCKERHUB_PASSWORD:-$(pass eirini/docker-hub)}"
export DOCKERHUB_USERNAME
export DOCKERHUB_PASSWORD

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
  "$EIRINI_RELEASE_DIR/scripts/generate-secrets.sh" "*.eirini-core.svc" "nothing-to-wiremock-here"

  pushd "$EIRINI_DIR"
  {
    render_dir=$(mktemp -d)
    trap "rm -rf $render_dir" EXIT
    ca_bundle="$(kubectl get secret -n eirini-core eirini-instance-index-env-injector-certs -o jsonpath="{.data['tls\.ca']}")"
    "$EIRINI_RELEASE_DIR/scripts/render-templates.sh" eirini-core "$render_dir" \
      --values "$EIRINI_RELEASE_DIR/scripts/assets/value-overrides.yml" \
      --set "webhook_ca_bundle=$ca_bundle,resource_validator_ca_bundle=$ca_bundle"
    kbld -f "$render_dir" -f "./scripts/kbld-local-eirini.yml" >"$render_dir/rendered.yml"
    for img in $(grep -oh "kbld:.*" "$render_dir/rendered.yml"); do
      kind load docker-image --name "eirini-tinypaas" "$img"
    done
    kapp -y deploy -a eirini -f "$render_dir/rendered.yml"
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
  "$REPO_ROOT"/install_kpack.sh
}

main() {
  ensure_kind_cluster "eirini-tinypaas"
  install_ingress_controller
  install_eirini
  install_kpack
  install_tinypaas
}

main
