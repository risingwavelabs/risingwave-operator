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

function risingwave::utils::delete_leader_lease() {
  local meta_leader_pod_names
  meta_leader_pod_names="$(k8s::kubectl::get pod -l risingwave/component=meta -l risingwave/meta-role=leader -o=jsonpath='{.items..metadata.name}')"

  if [ -z "$meta_leader_pod_names" ]; then
    logging::error "No meta leader node found"
    return 1
  fi

  if [ "$(echo "$meta_leader_pod_names" | wc -l)" -gt 1 ]; then
    logging::error "More than one meta leader node found"
    return 1
  fi

  k8s::kubectl port-forward svc/etcd 2388 &
  sleep 3

  # Iterate over the etcd election kv pairs. Delete leader lease if found, else abort the test
  local del_lease=false
  for i in $(ETCDCTL_API=3 etcdctl get __meta_election_ --prefix="true" --write-out="json" --endpoints=127.0.0.1:2388 | tail -1 | jq -c '.kvs[]'); do
    if [[ "$(echo "$i" | jq -r .value | base64 --decode)" == *"${meta_leader_pod_names}"* ]] ; then
      del_lease=true
      logging::info "found leader lease. Deleting it"
      
      # delete leader lease
      ETCDCTL_API=3 etcdctl del "$(echo "$i" | jq -r .key | base64 --decode)" --endpoints=127.0.0.1:2388
      break
    fi
  done
  
  kill $(pgrep kubectl)

  if [ "$del_lease" = false ] ; then
    logging::error "Could not delete leader lease. Leader pod names were ${meta_leader_pod_names}:5690"
    return 1
  fi

  return 0
}

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
