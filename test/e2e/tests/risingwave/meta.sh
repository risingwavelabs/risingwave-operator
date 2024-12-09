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

${__E2E_SOURCE_TESTS_RISINGWAVE_META_SH__:=false} && return 0 || __E2E_SOURCE_TESTS_RISINGWAVE_META_SH__=true

function risingwave::utils::kill_the_meta_leader_pod() {
	local meta_leaders
	meta_leaders="$(k8s::kubectl::get pod -l risingwave/component=meta -l risingwave/meta-role=leader -o=jsonpath='{.items..metadata.name}')"

	if [ -z "$meta_leaders" ]; then
		logging::error "No meta leader node found"
		return 1
	fi

	if [ "$(echo "$meta_leaders" | wc -l)" -gt 1 ]; then
		logging::error "More than one meta leader node found"
		return 1
	fi

	shell::run k8s::kubectl delete pod "$meta_leaders"
}

#######################################
# Check if meta is in a valid setup
# Returns
#   1 if no meta pod present, split-brain or no leader
#######################################
function risingwave::utils::is_meta_setup_valid() {
	local meta_leaders
	meta_leaders="$(k8s::kubectl::get pod -l risingwave/component=meta -l risingwave/meta-role=leader -o=jsonpath='{.items..metadata.name}')"

	if [ -z "$meta_leaders" ]; then
		logging::warn "No meta leader node found"
		return 1
	fi

	if [ "$(echo "$meta_leaders" | wc -l)" -gt 1 ]; then
		logging::warn "More than one meta leader node found"
		return 1
	fi
	return 0
}

function risingwave::utils::wait_for_meta_valid_setup() {
	local retry_count=0
	local retry_limit=${KUBECTL_JOB_WAIT_RETRY_LIMIT:=60}
	local retry_interval=${KUBECTL_JOB_WAIT_RETRY_INTERVAL:=1}
	while ((retry_count < retry_limit)); do
		((retry_count != 0)) && sleep "${retry_interval}"

		risingwave::utils::is_meta_setup_valid
		exit_code=$?

		# Condition met, return.
		((exit_code == 0)) && return 0

		# Condition unmet, retry.
		if ((exit_code == 1)); then
			retry_count=$((retry_count + 1))
			continue
		fi

		# On other errors, just return the exit code.
		return "${exit_code}"
	done
	logging::error "Timeout! Meta nodes are still in invalid setup!"
	return 1
}
