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

set -ex

BASEDIR=$(dirname "$0")
E2E_NAMESPACE="e2e-scaleview"

source "$BASEDIR"/k8s/kubernetes
source "$BASEDIR"/env-utils
source "$BASEDIR"/job/lib

prepare_cluster
overtrap stop_cluster EXIT
prepare_operator_image
install_locally_built_operator
overtrap uninstall_locally_built_risingwave_operator EXIT

function run_e2e_scale_view() {
  kubectl create ns $E2E_NAMESPACE
  # shellcheck disable=SC2064
  overtrap "kubectl delete ns $E2E_NAMESPACE" EXIT

  kubectl -n "$E2E_NAMESPACE" apply -f "$BASEDIR"/scaleview

  # Wait until the RisingWave is ready
  risingwave_name=$(kubectl -n "$E2E_NAMESPACE" get risingwave -o jsonpath='{.items[0].metadata.name}')
  echo "[E2E-SCALEVIEW] Waiting the RisingWave $risingwave_name to be ready..."
  kubectl -n "$E2E_NAMESPACE" wait --timeout=300s --for=condition=Running risingwave "$risingwave_name"
  wait_until_service_ready "$E2E_NAMESPACE" "$risingwave_name-frontend"

  # Wait until the RisingWaveScaleViews are locked, and then try scale.
  scale_view_names=$(kubectl -n "$E2E_NAMESPACE" get risingwavescaleview -o jsonpath='{.items[*].metadata.name}')

  for scale_view_name in $scale_view_names; do
    kubectl -n "$E2E_NAMESPACE" wait --timeout=10s --for=jsonpath='.status.locked'=true risingwavescaleview "$scale_view_name"

    kubectl -n "$E2E_NAMESPACE" scale risingwavescaleview/"$scale_view_name" --replicas=0

    kubectl -n "$E2E_NAMESPACE" wait --timeout=60s --for=jsonpath='.status.replicas'=0 risingwavescaleview "$scale_view_name"

    kubectl -n "$E2E_NAMESPACE" scale risingwavescaleview/"$scale_view_name" --replicas=3

    kubectl -n "$E2E_NAMESPACE" wait --timeout=300s --for=jsonpath='.status.replicas'=0 risingwavescaleview "$scale_view_name"

  done
}

run_e2e_scale_view
