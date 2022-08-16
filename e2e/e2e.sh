#!/bin/bash
#
# Copyright 2022 Singularity Data
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e

BASEDIR=$(dirname "$0")
K8S_ENV_LIST=$(ls -C "$BASEDIR"/envs)
E2E_TESTCASES=$(ls -C "$BASEDIR"/testcases)

function check_k8s_env() {
  echo "$K8S_ENV_LIST" | grep -w "$1" >/dev/null
}

K8S_ENV="${K8S_ENV:-kind-4}"
if ! check_k8s_env "$K8S_ENV"; then
  echo "ERROR: env $K8S_ENV not found"
  echo "Available envs: $K8S_ENV_LIST"
  exit 1
fi

# Source the env functions.
echo "Starting environment $K8S_ENV ..."

# shellcheck disable=SC1090
source "$BASEDIR"/envs/"$K8S_ENV"/env

# Start the Kubernetes.
start_kubernetes_env
echo "Started!"

# Set the trap for cleaning the Kubernetes.
trap stop_kubernetes_env EXIT

OS="$(uname -s)"
case "${OS}" in
Linux*) NIGHTLY_IMAGE_TAG="nightly"-$(date -d "2 day ago" '+%Y%m%d') ;;
Darwin*) NIGHTLY_IMAGE_TAG="nightly"-$(date -v-2d '+%Y%m%d') ;;
*) echo "ERROR: unsupported platform $OS" && exit 1 ;;
esac

# FIXME: currently the nightly tags aren't continuous.
NIGHTLY_IMAGE_TAG="nightly-20220727"
echo "Using a nightly tag $NIGHTLY_IMAGE_TAG for RisingWave images..."

# Prepare images...
function lazy_pull_image() {
  repo=$1
  tag=${2:-latest}

  if [ "$tag" = "latest" ] || [ -z "$(docker image ls -q "$repo":"$tag")" ]; then
    echo "Pulling image $repo:$tag..."
    docker pull "$repo:$tag"
  else
    echo "Image $repo:$tag already exists, skip"
  fi
}

echo "Pulling images..."
lazy_pull_image postgres
lazy_pull_image praqma/network-multitool
lazy_pull_image ghcr.io/singularity-data/risingwave "$NIGHTLY_IMAGE_TAG"
docker tag ghcr.io/singularity-data/risingwave:"$NIGHTLY_IMAGE_TAG" ghcr.io/singularity-data/risingwave:e2e
echo "Pulled!"

# Load images...
echo "Loading images..."
kubernetes_load_image postgres
kubernetes_load_image praqma/network-multitool
kubernetes_load_image ghcr.io/singularity-data/risingwave:e2e
kubernetes_load_image docker.io/singularity-data/risingwave-operator:dev
echo "Loaded!"

# Source the Kubernetes functions.
source "$BASEDIR"/k8s/kubernetes

# Install the RisingWave operator.
echo "Installing the RisingWave operator..."
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
trap "kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml" EXIT

function wait_cert_manager_certificate() {
  # wait for certificate
  threshold=40
  current_epoch=0
  while :; do
    certificate=$(kubectl get validatingwebhookconfigurations cert-manager-webhook -o jsonpath='{.webhooks[0].clientConfig.caBundle}')
    if [ -n "$certificate" ]; then
      break
    fi
    if [ $current_epoch -eq $threshold ]; then
      echo "ERROR: timeout waiting for cert-manager"
      exit 1
    fi
    sleep 2
    current_epoch=$((current_epoch + 1))
    echo "waiting for cert-manager to be ready ($current_epoch / $threshold)..."
  done
}

wait_cert_manager_certificate
wait_until_service_ready cert-manager cert-manager-webhook

kubectl apply -f "$BASEDIR"/../config/risingwave-operator-test.yaml
trap 'kubectl delete -f $BASEDIR/../config/risingwave-operator-test.yaml' EXIT

wait_until_service_ready risingwave-operator-system risingwave-operator-webhook-service
echo "RisingWave operator installed!"

# Start e2e testing...
source "$BASEDIR"/e2e

echo "Running E2E tests..."

## associative array for job status
JOBS=()

## run command in the background
background() {
  eval $1 &
  JOBS[$!]="$1"
}

## check exit status of each job
## preserve exit status in ${JOBS}
## returns 1 if any job failed
reap() {
  local cmd
  local status=0
  for pid in "${!JOBS[@]}"; do
    cmd=${JOBS[${pid}]}
    wait "${pid}"
    JOBS[${pid}]=$?
    if [[ ${JOBS[${pid}]} -ne 0 ]]; then
      status=${JOBS[${pid}]}
      echo -e "[${pid}] Exited with status: ${status}\n${cmd}"
    fi
  done
  return ${status}
}

echo "Testcases: ${E2E_TESTCASES}"
for testcase in ${E2E_TESTCASES}; do
  background "run_e2e_test $testcase"
done

if reap; then
  echo "Excellent! All tests are passed!"
else
  echo "Ooops! Some tests failed!"
  exit 1
fi
