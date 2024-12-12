#!/usr/bin/env bash

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

${__MANIFEST_TEST_SH__:=false} && return 0 || __MANIFEST_TEST_SH__=true

MANIFEST_DIR="$(dirname "${BASH_SOURCE[0]}")/../../../docs/manifests/risingwave"

test_file_list=(
	"risingwave-customize-config"
	"risingwave-in-memory"
)

function manifest_test::test_risingwave_file() {
	local file="$1"
	local file_path="${MANIFEST_DIR}/${file}.yaml"
	local ns="$1"

	logging::info "Start apply the RisingWave manifest file: ${file}.yaml"

	if ! shell::run k8s::kubectl create ns "${ns}"; then
		logging::error "Failed to create the namespace ${ns}!"
		exit 1
	fi

	shell::run kubectl apply -f "${file_path}" -n "${ns}" --dry-run=server

	logging::info "Succeed to test RisingWave manifest file ${file}.yaml"
}

function manifest_test::start() {
	for f in "${test_file_list[@]}"; do
		SHOW_COMMAND_OUTPUT=true manifest_test::test_risingwave_file "${f}"
	done
	exit 0
}
