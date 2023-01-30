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

${__E2E_SOURCE_COMMON_HELM_SH__:=false} && return 0 || __E2E_SOURCE_COMMON_HELM_SH__=true

source "$(dirname "${BASH_SOURCE[0]}")/shell.sh"

#######################################
# Utility function for running the helm command.
# Globals
#   HELM_NAMESPACE
# Arguments
#   Arguments for running helm.
# Returns
#   Code that helm returns.
#######################################
function helm::helm() {
  local extra_args=()
  [[ -n "${HELM_NAMESPACE}" ]] && extra_args+=(--namespace "${HELM_NAMESPACE}")
  helm "${extra_args[@]}" "$@"
}

#######################################
# Utility function for checking if a helm release is deployed.
# Globals
#   HELM_NAMESPACE
# Arguments
#   Release name.
# Returns
#   0 if it is, non-zero otherwise.
#######################################
function helm::release::is_deployed() {
  helm::helm list --deployed -q | grep -w -q "$1"
}

#######################################
# Utility function for checking if a helm release exists, no matter the status.
# Globals
#   HELM_NAMESPACE
# Arguments
#   Release name.
# Returns
#   0 if it is, non-zero otherwise.
#######################################
function helm::release::exists() {
  helm::helm list -q | grep -w -q "$1"
}

#######################################
# Utility function for checking if a helm repo exists.
# Globals
#   HELM_NAMESPACE
# Arguments
#   Repo name.
# Returns
#   0 if it is, non-zero otherwise.
#######################################
function helm::repo::exists() {
  helm::helm repo list | awk 'NR>1 {print $1}' | grep -w -q "$1"
}
