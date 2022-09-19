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
E2E_TESTCASES=$(ls -C "$BASEDIR"/testcases)

source "$BASEDIR"/k8s/kubernetes
source "$BASEDIR"/env-utils
source "$BASEDIR"/job/lib

prepare_cluster
prepare_operator_image

function install_operator() {
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
}

install_operator

# Start e2e testing...
echo "Running E2E tests..."

_E2E_SOURCE_BASEDIR=$(dirname "${BASH_SOURCE[0]}")

function run_e2e_test() {
  testcase="e2e-$1"
  testcase_dir=$_E2E_SOURCE_BASEDIR/testcases/$1

  if [ ! -d "$testcase_dir" ]; then
    echo "ERROR: testcase $testcase not found"
    return 1
  fi

  echo "[E2E $testcase] Creating the RisingWave..."

  if ! kubectl get ns "$testcase" >/dev/null 2>&1; then
    kubectl create ns "$testcase"
  fi
  # shellcheck disable=SC2064
  trap "kubectl delete ns $testcase" RETURN

  kubectl -n "$testcase" apply -f "$testcase_dir"
  risingwave_name=$(kubectl -n "$testcase" get risingwave -o jsonpath='{.items[0].metadata.name}')

  echo "[E2E $testcase] Waiting the RisingWave $risingwave_name to be ready..."
  kubectl -n "$testcase" wait --timeout=300s --for=condition=Running risingwave "$risingwave_name"
  wait_until_service_ready "$testcase" "$risingwave_name-frontend"

  echo "[E2E $testcase] RisingWave ready! Run simple queries..."
  if ! kubectl exec -t psql-console -- psql -h "$risingwave_name-frontend.$testcase" -p 4567 -d dev -U root <"$_E2E_SOURCE_BASEDIR"/check.sql; then
    echo "[E2E $testcase] ERROR: failed to execute simple queries"
    return 1
  fi

  echo "[E2E $testcase] Succeeds!"
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
