#!/bin/bash

set -euxo pipefail

NAMESPACE=eirini-core
TMPDIR=$(mktemp -d)
#trap "rm -rf $TMPDIR" EXIT
trap "echo $TMPDIR" EXIT

DOCKER_REGISTRY_SECRET_NAME=tinypaas-registry-credentials
GIT_SECRET_NAME=tinypaas-git-secret
BUILDER_NAME=tinypaas-builder
KPACK_SERVICE_ACCOUNT=tinypaas-service-account
readonly DOCKERHUB_USERNAME="${DOCKERHUB_USERNAME:?"Please provide dockerhub username"}"
readonly DOCKERHUB_PASSWORD="${DOCKERHUB_PASSWORD:?"Please provide dockerhub password"}"

pushd $TMPDIR
echo "Fetching kpack release"
wget https://github.com/pivotal/kpack/releases/download/v0.1.4/release-0.1.4.yaml -O kpack.yaml

echo "Installing kpack"
kubectl apply -f kpack.yaml

if ! kubectl get secrets -n "$NAMESPACE" $DOCKER_REGISTRY_SECRET_NAME &>/dev/null; then
  echo "Installing kpack"
  kubectl create secret docker-registry $DOCKER_REGISTRY_SECRET_NAME \
    --docker-username=${DOCKERHUB_USERNAME} \
    --docker-password=${DOCKERHUB_PASSWORD} \
    --docker-server=https://index.docker.io/v1/ \
    --namespace "$NAMESPACE"
fi

if ! kubectl get secrets -n "$NAMESPACE" $GIT_SECRET_NAME &>/dev/null; then
  # TODO: Only supports GitHub for now. Fix this.
  echo "Generating ssh key pair for Git"
  ssh-keygen -t ed25519 -f sshkey -N "" -q
  cat >git-secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: $GIT_SECRET_NAME
  annotations:
    kpack.io/git: git@github.com
type: kubernetes.io/ssh-auth
stringData:
  ssh-privatekey: |
$(sed -e 's/^/    /g' sshkey)
  ssh-publickey: |
$(sed -e 's/^/    /g' sshkey.pub)
EOF
  kubectl apply -f git-secret.yaml -n "$NAMESPACE"
fi

if ! kubectl get serviceaccount -n "$NAMESPACE" "$KPACK_SERVICE_ACCOUNT" &>/dev/null; then
  echo "Creating service account for kpack"
  cat >service-account.yaml <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: $KPACK_SERVICE_ACCOUNT
  namespace: $NAMESPACE
secrets:
- name: $DOCKER_REGISTRY_SECRET_NAME
- name: $GIT_SECRET_NAME
imagePullSecrets:
- name: $DOCKER_REGISTRY_SECRET_NAME
EOF
  kubectl apply -f service-account.yaml -n "$NAMESPACE"
fi

if ! kubectl get clusterstore -n "$NAMESPACE" tinypaas-cluster-store &>/dev/null; then
  echo "Creating cluster store for kpack"
  cat >cluster-store.yaml <<EOF
apiVersion: kpack.io/v1alpha1
kind: ClusterStore
metadata:
  name: tinypaas-cluster-store
spec:
  sources:
  - image: gcr.io/paketo-buildpacks/java
  - image: gcr.io/paketo-buildpacks/nodejs
EOF
  kubectl apply -f cluster-store.yaml -n "$NAMESPACE"
fi

if ! kubectl get clusterstack -n "$NAMESPACE" tinypaas-cluster-stack &>/dev/null; then
  echo "Creating cluster stack for kpack"
  cat >cluster-stack.yaml <<EOF
apiVersion: kpack.io/v1alpha1
kind: ClusterStack
metadata:
  name: tinypaas-cluster-stack
spec:
  id: "io.buildpacks.stacks.bionic"
  buildImage:
    image: "paketobuildpacks/build:base-cnb"
  runImage:
    image: "paketobuildpacks/run:base-cnb"
EOF
  kubectl apply -f cluster-stack.yaml -n "$NAMESPACE"
fi

if ! kubectl get builder -n "$NAMESPACE" tinypaas-builder &>/dev/null; then
  echo "Creating builder for kpack"
  cat >builder.yaml <<EOF
apiVersion: kpack.io/v1alpha1
kind: Builder
metadata:
  name: $BUILDER_NAME
  namespace: $NAMESPACE
spec:
  serviceAccount: $KPACK_SERVICE_ACCOUNT
  tag: eiriniuser/builder
  stack:
    name: tinypaas-cluster-stack
    kind: ClusterStack
  store:
    name: tinypaas-cluster-store
    kind: ClusterStore
  order:
  - group:
    - id: paketo-buildpacks/java
  - group:
    - id: paketo-buildpacks/nodejs
EOF
  kubectl apply -f builder.yaml -n "$NAMESPACE"
fi

popd

echo "Generating tinypaas cli configuration file"
cat >config.yaml <<CONFIG
namespace: $NAMESPACE
git_secret_name: $GIT_SECRET_NAME
docker_registry_secret_name: $DOCKER_REGISTRY_SECRET_NAME
builder_name: $BUILDER_NAME
kpack_service_account: $KPACK_SERVICE_ACCOUNT
CONFIG
