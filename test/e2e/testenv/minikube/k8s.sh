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

${__E2E_SOURCE_TESTENV_MINIKUBE_K8S_SH__:=false} && return 0 || __E2E_SOURCE_TESTENV_MINIKUBE_K8S_SH__=true

_TEST_ENV_MINIKUBE_DIR=$(dirname "${BASH_SOURCE[0]}")
_TEST_ENV_MINIKUBE_PROFILE="e2e"
_TEST_ENV_MINIKUBE_DEFAULT_NODES=4
_TEST_ENV_MINIKUBE_DEFAULT_KUBERNETES_VERSION="v1.25.3"

source "${_TEST_ENV_MINIKUBE_DIR}/../../common/lib.sh"

#######################################
# Start a local minikube cluster.
# Globals
#   MINIKUBE_NODES, defaults to ${_TEST_ENV_MINIKUBE_DEFAULT_NODES}
#   MINIKUBE_KUBERNETES_VERSION, defaults to ${_TEST_ENV_MINIKUBE_DEFAULT_KUBERNETES_VERSION}
# Arguments
#   None
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::minikube::provision() {
	# shellcheck disable=SC2034
	local LOGGING_TAGS=(testenv "k8s/minikube: ${_TEST_ENV_MINIKUBE_PROFILE}")

	local nodes=${MINIKUBE_NODES:-${_TEST_ENV_MINIKUBE_DEFAULT_NODES}}
	local kubernetes_version=${MINIKUBE_KUBERNETES_VERSION:-${_TEST_ENV_MINIKUBE_DEFAULT_KUBERNETES_VERSION}}

	logging::info "Start minikube cluster..."
	if shell::run minikube --profile="${_TEST_ENV_MINIKUBE_PROFILE}" start \
		--nodes="${nodes}" \
		--kubernetes-version="${kubernetes_version}"; then
		logging::info "Started!"
	else
		logging::error "Failed to start!"
		return 1
	fi
}

#######################################
# Stop the local minikube cluster.
# Globals
#   None
# Arguments
#   None
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::minikube::teardown() {
	# shellcheck disable=SC2034
	local LOGGING_TAGS=(testenv "k8s/minikube: ${_TEST_ENV_MINIKUBE_PROFILE}")

	logging::info "Stop minikube cluster..."
	if shell::run minikube --profile="${_TEST_ENV_MINIKUBE_PROFILE}" delete; then
		logging::info "Stopped!"
	else
		logging::error "Failed to stop!"
		return 1
	fi
}

#######################################
# Load local Docker images into the minikube cluster.
# Globals
#   None
# Arguments
#   Local Docker image name
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::minikube::load_docker_image() {
	# shellcheck disable=SC2034
	local LOGGING_TAGS=(testenv "k8s/minikube: ${_TEST_ENV_MINIKUBE_PROFILE}")

	logging::infof "Loading local Docker images:\n  %s\n" "$1"

	if shell::run minikube --profile="${_TEST_ENV_MINIKUBE_PROFILE}" image load "$1"; then
		logging::info "Successfully loaded!"
	else
		logging::error "Failed to load!"
	fi
}
