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

${__E2E_SOURCE_TESTS_RISINGWAVESCALEVIEW_TESTS_SH__:=false} && return 0 || __E2E_SOURCE_TESTS_RISINGWAVESCALEVIEW_TESTS_SH__=true

_E2E_RISINGWAVESCALEVIEW_TEST_PATH="$(dirname "${BASH_SOURCE[0]}")"

function test::risingwavescaleview::manifest_from() {
  local manifest_file="${_E2E_RISINGWAVESCALEVIEW_TEST_PATH}/manifests/$1"

  if [[ ! -f "${manifest_file}" ]]; then
    logging::error "${manifest_file} isn't a regular file"
    return 1
  fi

  envsubst "${@:2}" <"${manifest_file}"
}

function test::risingwavescaleview::start() {
  local relative_path="$1"

  if ! shell::run "test::risingwavescaleview::manifest_from ${relative_path} | k8s::kubectl apply -f -"; then
    logging::error "Failed to apply manifest!"
    return 1
  fi

  if ! k8s::risingwave::wait_before_rollout "${E2E_RISINGWAVE_NAME}"; then
    logging::error "Timeout waiting for the rollout!"
    return 1
  fi
}

function test::risingwavescaleview::scale_to() {
  local scaleview="$1"
  local replicas="$2"

  logging::info "Trying to scale to ${replicas}..."

  if ! shell::run k8s::kubectl scale risingwavescaleview/"${scaleview}" --replicas="${replicas}"; then
    logging::error "Failed to scale to ${replicas} with RisingWaveScaleView!"
    return 1
  fi

  if ! shell::run k8s::kubectl wait --timeout=60s --for=jsonpath='.status.replicas'="${replicas}" risingwavescaleview/"${scaleview}"; then
    logging::error "Timeout waiting before the replicas scaled to ${replicas}!"
    return 1
  fi
}

function test::risingwavescaleview::manual_checks() {
  local scaleview="${E2E_RISINGWAVE_NAME}-scaleview"

  if ! shell::run k8s::kubectl wait --timeout=10s --for=jsonpath='.status.locked'=true risingwavescaleview/"${scaleview}"; then
    logging::error "Failed to lock on target RisingWave!"
    return 1
  fi

  test::risingwavescaleview::scale_to "${scaleview}" 0 || return 1
  test::risingwavescaleview::scale_to "${scaleview}" 1 || return 1
  test::risingwavescaleview::scale_to "${scaleview}" 2 || return 1
  test::risingwavescaleview::scale_to "${scaleview}" 3 || return 1
  test::risingwavescaleview::scale_to "${scaleview}" 0 || return 1
}

function test::risingwavescaleview::stop() {
  local relative_path="$1"

  if ! shell::run "test::risingwavescaleview::manifest_from ${relative_path} | shell::run k8s::kubectl delete -f - --ignore-not-found"; then
    logging::error "Failed to delete with manifest!"
    return 1
  fi
}

function test::risingwavescaleview::manual_test_run() {
  local relative_path="$1"

  test::risingwavescaleview::start "${relative_path}" || return $?
  test::risingwavescaleview::manual_checks || return $?
  test::risingwavescaleview::stop "${relative_path}" || return $?
}

function test::run::risingwavescaleview::frontend_empty() {
  test::risingwavescaleview::manual_test_run manual/frontend-empty.yaml
}

function test::run::risingwavescaleview::frontend_multiple_group() {
  test::risingwavescaleview::manual_test_run manual/frontend-selected-group.yaml
}
