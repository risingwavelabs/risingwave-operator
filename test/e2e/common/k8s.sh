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

${__E2E_SOURCE_COMMON_K8S_SH__:=false} && return 0 || __E2E_SOURCE_COMMON_K8S_SH__=true

source "$(dirname "${BASH_SOURCE[0]}")/shell.sh"
source "$(dirname "${BASH_SOURCE[0]}")/logging.sh"

#######################################
# Utility function for running the kubectl command.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   Arguments for running kubectl.
# Returns
#   Code that kubectl returns.
#######################################
function k8s::kubectl() {
  local extra_args=()
  [[ -v "KUBECTL_NAMESPACE" && -n "${KUBECTL_NAMESPACE}" ]] && extra_args+=(-n "${KUBECTL_NAMESPACE}")
  kubectl "${extra_args[@]}" "$@"
}

#######################################
# Utility function for running the kubectl get command on a specified object. This wrapper hides the output
# from STDERR when kubectl fails.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   Resource kind, e.g., pod
#   Resource name, e.g., web-pod-13je7
#   Other kubectl arguments.
# Output
#   STDOUT when succeeds.
# Returns
#   0 if the object exists, 255 if not, error code returned by kubectl otherwise.
#   254 will be returned when the original exit code is 255 to avoid conflict.
#######################################
function k8s::kubectl::get() {
  (($# >= 2)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$1" ]] || { echo >&2 "resource kind must be provided" && return 1; }
  [[ -n "$2" ]] || { echo >&2 "resource name must be provided" && return 1; }

  if shell::run_and_capture_outputs k8s::kubectl get "$@"; then
    echo "${CAPTURED_STDOUT}"
    return 0
  else
    [[ "${CAPTURED_STDERR}" == *"not found"* ]] && return 255
    ((CAPTURED_EXIT_CODE == 255)) && return 254
    return "${CAPTURED_EXIT_CODE}"
  fi
}

#######################################
# Utility function for checking if the object exists.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   Resource kind, e.g., pod
#   Resource name, e.g., web-pod-13je7.
# Returns
#   0 if the object exists, 255 if not, error code returned by kubectl otherwise.
#######################################
function k8s::kubectl::object_exists() {
  k8s::kubectl::get "$1" "$2" >/dev/null
}

#######################################
# Utility function for checking if a Job is completed.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   Job name.
# Returns
#   0 if it is, 1 if not, 255 if the Job object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::job::is_completed() {
  local complete_status
  local exit_code=0
  complete_status=$(k8s::kubectl::get job "$1" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}') || exit_code=$?

  ((exit_code == 0)) || return "${exit_code}"

  [[ "${complete_status,,}" == "true" ]] || return 1
}

#######################################
# Utility function for checking if a Job is failed.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   Job name.
# Returns
#   0 if it is, 1 if not, 255 if the Job object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::job::is_failed() {
  local failed_status
  local exit_code=0

  failed_status=$(k8s::kubectl::get job "$1" -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}') || exit_code=$?

  ((exit_code == 0)) || return "${exit_code}"

  [[ "${failed_status,,}" == "true" ]] || return 1
}

K8S_JOB_COMPLETED=""
K8S_JOB_FAILED=""

#######################################
# Utility function for checking if a Job is completed or failed.
# Globals
#   KUBECTL_NAMESPACE
#   K8S_JOB_COMPLETED
#   K8S_JOB_FAILED
# Arguments
#   Job name.
# Returns
#   0 if it is, 1 if not, 255 if the object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::job::is_completed_or_failed() {
  local complete_and_failed_status
  local exit_code=0

  complete_and_failed_status=$(k8s::kubectl::get job "$1" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status},{.status.conditions[?(@.type=="Failed")].status},') || exit_code=$?

  ((exit_code == 0)) || return "${exit_code}"

  # Load output into the array.
  local -a array
  IFS="," read -r -a array <<<"${complete_and_failed_status}"

  local complete_status="${array[0]}"
  local failed_status="${array[1]}"

  # Set the global vars.
  K8S_JOB_COMPLETED=false
  K8S_JOB_FAILED=false

  if [[ "${complete_status,,}" == "true" ]]; then
    K8S_JOB_COMPLETED=true
  fi

  if [[ "${failed_status,,}" == "true" ]]; then
    K8S_JOB_FAILED=true
  fi

  [[ "${K8S_JOB_COMPLETED}" == "true" || "${K8S_JOB_FAILED}" == "true" ]] || return 1
}

#######################################
# Utility function for waiting until a Job complete or fail.
# Globals
#   KUBECTL_NAMESPACE
#   KUBECTL_JOB_WAIT_RETRY_LIMIT
#   KUBECTL_JOB_WAIT_RETRY_INTERVAL
#   K8S_JOB_COMPLETED
#   K8S_JOB_FAILED
# Arguments
#   Job name.
# Returns
#   0 if it completes or fails, 255 if the object doesn't exist, 1 when timeout, and other codes when kubectl fails.
#######################################
function k8s::job::wait_before_completed_or_failed() {
  (($# == 1)) || { echo >&2 "not enough arguments" && return 1; }

  local retry_count=0
  local retry_limit=${KUBECTL_JOB_WAIT_RETRY_LIMIT:=60}
  local retry_interval=${KUBECTL_JOB_WAIT_RETRY_INTERVAL:=5}

  local exit_code=0
  while ((retry_count < retry_limit)); do
    ((retry_count != 0)) && sleep "${retry_interval}"

    k8s::job::is_completed_or_failed "$1"
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

  logging::debug "Timeout waiting for Job $1 to complete or fail!"
  return 1
}

#######################################
# Utility function for getting debug info for a Job.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   Job name.
# Outputs
#   STDOUT
# Returns
#   0 on success, 255 if the Job object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::job::debug() {
  local job=$1

  local manifest
  manifest=$(k8s::kubectl::get job "${job}" -o yaml)

  printf "Job manifest in YAML:\n%s\n" "${manifest}"
  printf "Pods controlled by Job %s:\n%s\n" "${job}" "$(k8s::kubectl::get pod -l job-name="${job}")"
}

#######################################
# Utility function for checking if the specified RisingWave is rolled out.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   RisingWave name
# Returns
#   0 if it is, 1 if not, 255 if the object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::risingwave::is_rolled_out() {
  local content
  content=$(k8s::kubectl::get risingwave "$1" -o jsonpath='{.metadata.generation},{.status.observedGeneration},{.status.conditions[?(@.type=="Running")].status},{.status.conditions[?(@.type=="Upgrading")].status},') || return $?

  # Load output into the array.
  local -a generation_and_conditions
  IFS="," read -r -a generation_and_conditions <<<"${content}"

  local current_generation="${generation_and_conditions[0]}"
  local observed_generation="${generation_and_conditions[1]}"
  local running_condition="${generation_and_conditions[2]}"
  local upgrading_condition="${generation_and_conditions[3]}"

  if ((current_generation == observed_generation)) &&
    [[ "${running_condition}" == "True" && ("${upgrading_condition}" == "" || "${upgrading_condition}" == "False") ]]; then
    return 0
  else
    return 1
  fi
}

#######################################
# Utility function for waiting before the specified RisingWave is rolled out.
# Globals
#   KUBECTL_NAMESPACE
#   KUBECTL_RISINGWAVE_WAIT_RETRY_LIMIT
#   KUBECTL_RISINGWAVE_WAIT_RETRY_INTERVAL
# Arguments
#   RisingWave name
# Returns
#   0 if it is, 1 if not, 255 if the object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::risingwave::wait_before_rollout() {
  (($# == 1)) || { echo >&2 "not enough arguments" && return 1; }

  local retry_count=0
  local retry_limit=${KUBECTL_RISINGWAVE_WAIT_RETRY_LIMIT:=60}
  local retry_interval=${KUBECTL_RISINGWAVE_WAIT_RETRY_INTERVAL:=5}

  local exit_code=0
  while ((retry_count < retry_limit)); do
    ((retry_count != 0)) && sleep "${retry_interval}"

    k8s::risingwave::is_rolled_out "$1"
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

  logging::debug "Timeout waiting for RisingWave $1 to rollout!"
  return 1
}

#######################################
# Utility function for getting debug info for a RisingWave.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   RisingWave name.
# Outputs
#   STDOUT
# Returns
#   0 on success, 255 if the RisingWave object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::risingwave::debug() {
  local risingwave=$1

  local manifest
  manifest=$(k8s::kubectl::get risingwave "${risingwave}" -o yaml)

  printf "RisingWave manifest in YAML:\n%s\n" "${manifest}"
  printf "Pods controlled by RisingWave %s:\n%s\n" "${risingwave}" "$(k8s::kubectl::get pod -l risingwave/name="${risingwave}")"
}

#######################################
# Check if the deployment has been rolled out.
# Globals
#   KUBECTL_NAMESPACE
# Arguments
#   Deployment name.
# Returns
#   0 for true, non-zero when false or error occurs.
#######################################
function k8s::deployment::is_rolled_out() {
  local content
  content=$(k8s::kubectl::get deployment "$1" -o jsonpath='{.metadata.generation},{.status.observedGeneration},{.spec.replicas},{.status.availableReplicas},{.status.readyReplicas},{.status.replicas},{.status.updatedReplicas},') || return $?

  # Load output into the array.
  local -a array
  IFS="," read -r -a array <<<"${content}"

  local generation="${array[0]}"
  local observed_generation="${array[1]}"
  local replicas="${array[2]}"
  local available_replicas="${array[3]}"
  # local ready_replicas="${array[4]}"
  local current_replicas="${array[5]}"
  local updated_replicas="${array[6]}"

  ((generation == observed_generation)) || return 1
  ((updated_replicas >= replicas)) || return 1
  ((current_replicas <= updated_replicas)) || return 1
  ((available_replicas >= updated_replicas)) || return 1
}

#######################################
# Utility function for waiting before the specified Deployment is rolled out.
# Globals
#   KUBECTL_NAMESPACE
#   KUBECTL_WAIT_RETRY_LIMIT
#   KUBECTL_WAIT_RETRY_INTERVAL
# Arguments
#   Deployment name
# Returns
#   0 if it is, 1 if not, 255 if the object doesn't exists, and other codes when kubectl fails.
#######################################
function k8s::deployment::wait_before_rollout() {
  (($# == 1)) || { echo >&2 "not enough arguments" && return 1; }

  local retry_count=0
  local retry_limit=${KUBECTL_WAIT_RETRY_LIMIT:=60}
  local retry_interval=${KUBECTL_WAIT_RETRY_INTERVAL:=5}

  local exit_code=0
  while ((retry_count < retry_limit)); do
    ((retry_count != 0)) && sleep "${retry_interval}"

    k8s::deployment::is_rolled_out "$1"
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

  logging::debug "Timeout waiting for Deployment $1 to rollout!"
  return 1
}

function k8s::pod::delete_meat_leader_pod() {
  local meta_names
  meta_names="$(k8s::kubectl::get pod -l risingwave/component=meta -o=jsonpath='{.items..metadata.name}')"
  if [ -z "$meta_names" ] ; then 
    logging::error "failed to retrieve meta nodes"
    return 1
  fi
  
  local leader=""
  for p in $meta_names ; do 
    local is_leader
    is_leader=$(k8s::kubectl logs "$p" | grep "Defining leader services")
    if [ "$is_leader" != "" ]; then 
      if [ "$leader" != "" ]; then 
        logging::error "Split brain detected! $p and $leader are leaders!"
        # TODO: also abort on split-brain here
      fi
      leader="$p"
    fi
  done

  k8s::kubectl delete pod "$leader" --force
}

#######################################
# Check if meta is in a valid setup
# Returns
#   1 if no meta pod present, split-brain or no leader
#######################################
function k8s::pod::meta_pod_valid_setup() {
  local meta_names
  meta_names="$(k8s::kubectl::get pod -l risingwave/component=meta -o=jsonpath='{.items..metadata.name}')"
  if [ -z "$meta_names" ] ; then 
    logging::error "failed to retrieve meta nodes"
    return 1
  fi

  local leader=""
  for p in $meta_names ; do 
    local is_leader
    is_leader="$(k8s::kubectl logs "$p" | grep "Defining leader services")"
    if [ "$is_leader" != "" ]; then 
      if [ "$leader" != "" ]; then 
        logging::error "Split brain detected! $p and $leader are leaders!"
        # return 1 # TODO: enable abort
      fi
      leader="$p"
    fi
  done

  if [ -z "$leader" ]; then 
    logging::error "no meta leader found"
    return 1
  fi 
  return 0
}


function k8s::pod::wait_for_meta_pod_valid_setup() {
  local retry_count=0
  local retry_limit=${KUBECTL_JOB_WAIT_RETRY_LIMIT:=60}
  local retry_interval=${KUBECTL_JOB_WAIT_RETRY_INTERVAL:=5}
  while ((retry_count < retry_limit)); do
    ((retry_count != 0)) && sleep "${retry_interval}"

    k8s::pod::meta_pod_valid_setup
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
  logging::error "Timeout reached. Meta still in invalid setup"
  return 1
}