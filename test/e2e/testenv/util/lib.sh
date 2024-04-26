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

${__E2E_SOURCE_TESTENV_UTIL_LIB_SH__:=false} && return 0 || __E2E_SOURCE_TESTENV_UTIL_LIB_SH__=true

_TEST_ENV_UTIL_PATH="$(dirname "${BASH_SOURCE[0]}")"
_UTIL_NAMESPACE="util"

source "${_TEST_ENV_UTIL_PATH}/../../common/lib.sh"

function testenv::util::_relative_path() {
	echo "${_TEST_ENV_UTIL_PATH}/$1"
}

function testenv::util::_is_installed() {
	k8s::kubectl::object_exists namespace "${_UTIL_NAMESPACE}"
}

function testenv::util::install() {
	# shellcheck disable=SC2034
	local LOGGING_TAGS=(e2e utility "namespace/${_UTIL_NAMESPACE}")

	if ! shell::run "kubectl create namespace ${_UTIL_NAMESPACE} --dry-run=client -o yaml | shell::run kubectl apply -f -"; then
		logging::error "Failed to create namespace for installing utilities!"
		return 1
	fi

	# shellcheck disable=SC2034
	local KUBECTL_NAMESPACE="${_UTIL_NAMESPACE}"

	if ! shell::run k8s::kubectl apply -f "$(testenv::util::_relative_path manifests)"; then
		logging::error "Failed to apply manifests for utilities!"
		return 1
	fi

	if ! shell::run k8s::kubectl wait --timeout=300s --for=condition=Ready pod/psql; then
		logging::error "Pod psql failed to be ready!"
		return 1
	fi

	if ! shell::run k8s::kubectl wait --timeout=300s --for=condition=Ready pod/network-utils; then
		logging::error "Pod network-utils failed to be ready!"
		return 1
	fi
}

function testenv::util::uninstall() {
	# shellcheck disable=SC2034
	local LOGGING_TAGS=(e2e utility "namespace/${_UTIL_NAMESPACE}")

	if ! shell::run "kubectl create namespace ${_UTIL_NAMESPACE} --dry-run=client -o yaml | kubectl delete --ignore-not-found -f -"; then
		logging::error "Failed to delete utilities namespace!"
		return 1
	fi
}

function testenv::util::psql() {
	if [[ -n ${PSQL_SCRIPT_FILE+x} ]]; then
		kubectl -n "${_UTIL_NAMESPACE}" exec -i psql -c psql -- psql "$@" <"${PSQL_SCRIPT_FILE}"
	else
		kubectl -n "${_UTIL_NAMESPACE}" exec -i psql -c psql -- psql "$@"
	fi
}

function testenv::util::network::test_connectivity() {
	# shellcheck disable=SC2034
	local KUBECTL_NAMESPACE="${_UTIL_NAMESPACE}"

	shell::run k8s::kubectl exec network-utils -c network-utils -- nc -zvw3 "$1" "$2"
}

function testenv::util::network::is_k8s_service_up() {
	local namespace=$1
	local service=$2
	local port=""

	# shellcheck disable=SC2034
	local KUBECTL_NAMESPACE="${namespace}"

	if (($# >= 3)); then port=$3; fi

	# If port isn't a number, get it via kubectl
	if [[ -z ${port} ]]; then
		port=$(k8s::kubectl::get svc/"${service}" -o jsonpath="{.spec.ports[0].port}") || return $?
	elif [[ $port =~ ^[0-9]+$ ]]; then
		port=$(k8s::kubectl::get svc/"${service}" -o jsonpath="{.spec.ports[?(@.port==${port})].port}") || return $?
	else
		port=$(k8s::kubectl::get svc/"${service}" -o jsonpath="{.spec.ports[?(@.name==\"${port}\")].port}") || return $?
	fi

	if [[ -z ${port} ]] || ((port <= 0 || port > 65535)); then
		logging::error "Invalid port!"
		return 1
	fi

	# Run connectivity check.
	testenv::util::network::test_connectivity "${service}.${namespace}.svc" "${port}"
}

function testenv::util::network::wait_before_service_up() {
	local threshold=60
	local current_epoch=0
	local interval=5
	while :; do
		if testenv::util::network::is_k8s_service_up "${@}"; then
			break
		fi

		if ((current_epoch == threshold)); then
			logging::error "Timeout waiting for service ${service} under namespace ${namespace} to start!"
			return 1
		fi

		((current_epoch++))
		sleep "${interval}"
	done
}
