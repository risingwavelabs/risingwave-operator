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

${__E2E_SOURCE_TESTENV_LIB_SH__:=false} && return 0 || __E2E_SOURCE_TESTENV_LIB_SH__=true

source "$(dirname "${BASH_SOURCE[0]}")/../common/lib.sh"

source "$(dirname "${BASH_SOURCE[0]}")/kind/k8s.sh"
source "$(dirname "${BASH_SOURCE[0]}")/minikube/k8s.sh"
source "$(dirname "${BASH_SOURCE[0]}")/util/lib.sh"

function testenv::k8s::_use_local_context() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(testenv "k8s/local")

  local current_context
  if current_context=$(kubectl config current-context 2>/dev/null); then
    logging::info "Using local context: ${current_context}!"
  else
    logging::error "Current context is not set! Set it and re-run the tests!"
    return 1
  fi
}

function testenv::is_local() {
  local kubernetes_runtime=${E2E_KUBERNETES_RUNTIME:-"local"}
  [[ "${kubernetes_runtime}" == "local" ]]
}

#######################################
# Start a local Kubernetes cluster for E2E tests.
# Globals
#   E2E_KUBERNETES_RUNTIME, available values are "kind", "minikube", and "local". Defaults to "local".
# Arguments
#   None
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::provision() {
  local kubernetes_runtime=${E2E_KUBERNETES_RUNTIME:-"local"}

  case "${kubernetes_runtime}" in
  kind)
    testenv::k8s::kind::provision
    ;;
  minikube)
    testenv::k8s::minikube::provision
    ;;
  local)
    testenv::k8s::_use_local_context
    ;;
  *)
    logging::errorf "Unrecognized Kubernetes runtime: %s\n" "${kubernetes_runtime}"
    exit 1
    ;;
  esac
}

#######################################
# Stop the local Kubernetes cluster for E2E tests.
# Globals
#   E2E_KUBERNETES, available values are "kind", "minikube", and "local". Defaults to "local".
# Arguments
#   None
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::teardown() {
  local kubernetes_runtime=${E2E_KUBERNETES_RUNTIME:-"local"}

  case "${kubernetes_runtime}" in
  kind)
    testenv::k8s::kind::teardown
    ;;
  minikube)
    testenv::k8s::minikube::teardown
    ;;
  local)
    # Nothing to do.
    ;;
  *)
    logging::errorf "Unrecognized Kubernetes runtime: %s\n" "${kubernetes_runtime}"
    exit 1
    ;;
  esac
}

function testenv::k8s::load_docker_image() {
  local kubernetes_runtime=${E2E_KUBERNETES_RUNTIME:-"local"}

  case "${kubernetes_runtime}" in
  kind)
    testenv::k8s::kind::load_docker_image "$@"
    ;;
  minikube)
    testenv::k8s::minikube::load_docker_image "$@"
    ;;
  local)
    # Nothing to do.
    ;;
  *)
    logging::errorf "Unrecognized Kubernetes runtime: %s\n" "${kubernetes_runtime}"
    exit 1
    ;;
  esac
}

function testenv::k8s::cert_manager::install() {
  local version=${CERT_MANAGER_VERSION:-v1.9.1}
  shell::run k8s::kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/"${version}"/cert-manager.yaml

  testenv::k8s::cert_manager::_wait
}

function testenv::k8s::cert_manager::uninstall() {
  local version=${CERT_MANAGER_VERSION:-v1.9.1}
  shell::run k8s::kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/"${version}"/cert-manager.yaml
}

function testenv::k8s::cert_manager::_wait_validating_webhook_ca_bundle() {
  local certificate
  local threshold=60
  local current_epoch=0
  local interval=5
  while :; do
    certificate=$(k8s::kubectl::get validatingwebhookconfigurations cert-manager-webhook -o jsonpath='{.webhooks[0].clientConfig.caBundle}')
    if [ -n "$certificate" ]; then
      break
    fi
    if ((current_epoch == threshold)); then
      logging::error "Timeout waiting for cert-manager's CA bundle to be ready!"
      return 1
    fi

    ((current_epoch++))
    sleep "${interval}"
  done
}

function testenv::k8s::cert_manager::_wait_before_webhook_service_up() {
  testenv::util::network::wait_before_service_up cert-manager cert-manager-webhook
}

function testenv::k8s::cert_manager::_wait() {
  testenv::k8s::cert_manager::_wait_validating_webhook_ca_bundle
  testenv::k8s::cert_manager::_wait_before_webhook_service_up
}

_RISINGWAVE_OPERATOR_NAMESPACE="risingwave-operator-system"
_RISINGWAVE_OPERATOR_TEST_IMAGE="risingwavelabs/risingwave-operator:dev"
_RISINGWAVE_OPERATOR_MANIFEST_FOR_TEST_PATH="$(dirname "${BASH_SOURCE[0]}")/../../../config/risingwave-operator-test.yaml"

function testenv::k8s::risingwave_operator::install() {
  testenv::k8s::load_docker_image "${_RISINGWAVE_OPERATOR_TEST_IMAGE}"
  shell::run k8s::kubectl apply -f "${_RISINGWAVE_OPERATOR_MANIFEST_FOR_TEST_PATH}"

  # shellcheck disable=SC2034
  local KUBECTL_NAMESPACE="${_RISINGWAVE_OPERATOR_NAMESPACE}"
  k8s::deployment::wait_before_rollout "risingwave-operator-controller-manager"
  testenv::util::network::wait_before_service_up "${_RISINGWAVE_OPERATOR_NAMESPACE}" "risingwave-operator-webhook-service"
}

function testenv::k8s::risingwave_operator::uninstall() {
  shell::run k8s::kubectl delete -f "${_RISINGWAVE_OPERATOR_MANIFEST_FOR_TEST_PATH}"
}

function testenv::setup() {
  logging::info "Setting up test env..."
  testenv::k8s::provision || return $?
  testenv::util::install || return $?

  logging::info "Installing cert-manager..."
  testenv::k8s::cert_manager::install || return $?
  logging::info "Installing risingwave-operator..."
  testenv::k8s::risingwave_operator::install || return $?

  logging::info "Test env all set!"
}

function testenv::teardown() {
  logging::info "Tearing down the test env..."

  if testenv::is_local; then
    testenv::util::uninstall
    logging::info "Uninstalling cert-manager and risingwave-operator..."
    testenv::k8s::risingwave_operator::uninstall
    testenv::k8s::cert_manager::uninstall
  fi

  testenv::k8s::teardown
  logging::info "Test env teared down!"
}
