#!/bin/bash
#
# Copyright 2023 RisingWave Labs
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

set -ex

BASEDIR=$(dirname "$0")
E2E_TESTCASES=$(ls -C "$BASEDIR"/testcases)

source "$BASEDIR"/k8s/kubernetes
source "$BASEDIR"/env-utils
source "$BASEDIR"/job/lib

prepare_cluster
overtrap stop_cluster EXIT
prepare_operator_image
install_locally_built_operator
overtrap uninstall_locally_built_risingwave_operator EXIT

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
    # Print resources under the testcase namespace.
    kubectl -n "$testcase" get all
    return 1
  fi

  echo "[E2E $testcase] Succeeds!"
}

function run_test_on_scale_view() {
  namespace=$1
  scale_view_name=$2

  # Wait until the RisingWaveScaleView is ready.
  kubectl -n "$namespace" wait --timeout=10s --for=jsonpath='.status.locked'=true risingwavescaleview "$scale_view_name" || return 1

  # Scale the replicas to 0 and wait for the status to be 0.
  kubectl -n "$namespace" scale risingwavescaleview/"$scale_view_name" --replicas=0 || return 1
  kubectl -n "$namespace" wait --timeout=60s --for=jsonpath='.status.replicas'=0 risingwavescaleview "$scale_view_name" || return 1

  # Scale the replicas to 3 and wait for the status to be 1.
  kubectl -n "$namespace" scale risingwavescaleview/"$scale_view_name" --replicas=3 || return 1
  kubectl -n "$namespace" wait --timeout=300s --for=jsonpath='.status.replicas'=0 risingwavescaleview "$scale_view_name" || return 1
}

function run_e2e_scale_view() {
  E2E_NAMESPACE="e2e-scaleview"

  kubectl create ns $E2E_NAMESPACE
  # shellcheck disable=SC2064
  trap "kubectl delete ns $E2E_NAMESPACE" RETURN

  kubectl -n "$E2E_NAMESPACE" apply -f "$BASEDIR"/scaleview

  # Wait until the RisingWave is ready
  risingwave_name=$(kubectl -n "$E2E_NAMESPACE" get risingwave -o jsonpath='{.items[0].metadata.name}')
  echo "[E2E-SCALEVIEW] Waiting the RisingWave $risingwave_name to be ready..."
  kubectl -n "$E2E_NAMESPACE" wait --timeout=300s --for=condition=Running risingwave "$risingwave_name"
  wait_until_service_ready "$E2E_NAMESPACE" "$risingwave_name-frontend"

  # Wait until the RisingWaveScaleViews are locked, and then try scale.
  scale_view_names=$(kubectl -n "$E2E_NAMESPACE" get risingwavescaleview -o jsonpath='{.items[*].metadata.name}')

  for scale_view_name in $scale_view_names; do
    if ! run_test_on_scale_view $E2E_NAMESPACE "$scale_view_name"; then
      echo "[E2E-SCALEVIEW] ERROR: failed to run test on RisingWaveScaleView $scale_view_name"
      # Print resources under the testcase namespace.
      kubectl -n "$testcase" get all
      return 1
    fi
  done
}

# Run E2E testcases of RisingWave
echo "Testcases: ${E2E_TESTCASES}"
for testcase in ${E2E_TESTCASES}; do
  background "run_e2e_test $testcase"
done

# Run E2E testcases of RisingWaveScaleView
background run_e2e_scale_view

if reap; then
  echo "Excellent! All tests are passed!"
else
  echo "Ooops! Some tests failed!"
  exit 1
fi
