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

${__E2E_SOURCE_TESTS_RISINGWAVE_TESTS_SH__:=false} && return 0 || __E2E_SOURCE_TESTS_RISINGWAVE_TESTS_SH__=true

_E2E_RISINGWAVE_TEST_PATH="$(dirname "${BASH_SOURCE[0]}")"

source "${_E2E_RISINGWAVE_TEST_PATH}/meta.sh"

function test::risingwave::manifest_from() {
  local manifest_file="${_E2E_RISINGWAVE_TEST_PATH}/manifests/$1"

  if [[ ! -f "${manifest_file}" ]]; then
    logging::error "${manifest_file} isn't a regular file"
    return 1
  fi
  envsubst "${@:2}" <"${manifest_file}"
}

function test::risingwave::enable_openkruise() {
  logging::info "Enabling openkruise at Risingwave level"
  shell::run kubectl patch risingwave -n "${E2E_NAMESPACE}" "${E2E_RISINGWAVE_NAME}" --type merge -p '{\"spec\":{\"enableOpenKruise\":true}}'
}

function test::risingwave::start() {
  local relative_path="$1"

  if ! shell::run "test::risingwave::manifest_from ${relative_path} | k8s::kubectl apply -f -"; then
    logging::error "Failed to apply manifest!"
    return 1
  fi

  if [ "${OPEN_KRUISE_ENABLED_IN_RISINGWAVE}" -eq 1 ]; then
    test::risingwave::enable_openkruise
  fi

  if ! k8s::risingwave::wait_before_rollout "${E2E_RISINGWAVE_NAME}"; then
    logging::error "Timeout waiting for the rollout!"
    return 1
  fi

  testenv::util::network::wait_before_service_up "${E2E_NAMESPACE}" "${E2E_RISINGWAVE_NAME}-frontend" service
}

function test::risingwave::stop() {
  local relative_path="$1"

  if ! shell::run "test::risingwave::manifest_from ${relative_path} | shell::run k8s::kubectl delete -f - --ignore-not-found"; then
    logging::error "Failed to delete with manifest!"
    return 1
  fi
}

function test::run::risingwave::multi_meta() {
  test::risingwave::storage_support::_run_with_manifest multi-meta/multi-meta.yaml
}

function test::util::setup_multi_meta_failover() {
  # Tests if old leader exited after loosing leadership and asserts valid leader setup
  logging::info "Starting RisingWave..."
  if ! test::risingwave::start multi-meta/multi-meta.yaml; then
    return 1
  fi
  logging::info "Started!"

  # Check and see if the meta setup is valid, i.e., there must be only one meta leader.
  if ! risingwave::utils::is_meta_setup_valid; then
    logging::error "Invalid meta setup. Aborting test!"
    return 1
  fi
  return 0
}

function test::run::risingwave::multi_meta_failover_fencing() {
  if ! test::util::setup_multi_meta_failover; then
    return 1
  fi

  local meta_leaders_restarts
  meta_leaders_restarts="$(k8s::kubectl::get pod -l risingwave/component=meta -l risingwave/meta-role=leader -o=jsonpath='{.items..status.containerStatuses..restartCount}')"
  local meta_leader_names
  meta_leader_names="$(k8s::kubectl::get pod -l risingwave/component=meta -l risingwave/meta-role=leader -o=jsonpath='{.items..metadata.name}')"


  if ! risingwave::utils::delete_leader_lease; then 
    logging::error "Failed to delete leader lease"
    return 1
  fi
  
  logging::info "Waiting until deleting leader lease takes effect"
  sleep 10


  if ! risingwave::utils::wait_for_meta_valid_setup; then 
    logging::error "Meta not in valid setup after deleting leader lease"
    return 1
  fi 

  local new_meta_leaders_restarts
  new_meta_leaders_restarts="$(k8s::kubectl::get pod -l risingwave/component=meta -l risingwave/meta-role=leader -o=jsonpath='{.items..status.containerStatuses..restartCount}')"
  if [ "$new_meta_leaders_restarts" -le "$meta_leaders_restarts" ]; then
    logging::error "Leader did not restart"
    return 1
  fi

  local new_meta_leader_names
  new_meta_leader_names="$(k8s::kubectl::get pod -l risingwave/component=meta -l risingwave/meta-role=leader -o=jsonpath='{.items..metadata.name}')"
  if [ "$new_meta_leader_names" == "$meta_leader_names" ]; then
    logging::error "Leader did not change"
    return 1
  fi
  return 0
}

# Test if there is a single leader setup after the current leader failed
function test::run::risingwave::multi_meta_failover() {
  if ! test::util::setup_multi_meta_failover; then
    return 1
  fi

  # Simulate the failover by deleting the leader Pod.
  # NOTE: the leader doesn't necessarily change.
  logging::info "Killing the leader Pod..."
  risingwave::utils::kill_the_meta_leader_pod

  # Wait before the meta come back to a valid setup.
  if ! risingwave::utils::wait_for_meta_valid_setup; then
    logging::error "Invalid meta setup after meta crash!"
    return 1
  else
    logging::info "Failover successfully!"
  fi

  if ! test::risingwave::check_status_with_simple_queries; then
    logging::error "Queries run against storage failed!"
    return 1
  fi
  logging::info "Queries succeeded!"

  logging::info "Stopping RisingWave..."
  test::risingwave::stop multi-meta/multi-meta.yaml
  logging::info "Stopped!"
}

function test::risingwave::check_status_with_simple_queries() {
  local frontend_service_port
  frontend_service_port=$(k8s::kubectl get svc/"${E2E_RISINGWAVE_NAME}-frontend" -o jsonpath='{.spec.ports[?(@.name=="service")].port}')

  testenv::util::psql -h "${E2E_RISINGWAVE_NAME}-frontend.${E2E_NAMESPACE}" -p "${frontend_service_port}" -d dev -U root <<EOF
/* create a table */
create table t1(v1 int);

/* create a materialized view based on the previous table */
create materialized view mv1 as select sum(v1) as sum_v1 from t1;

/* insert some data into the source table */
insert into t1 values (1), (2), (3);

/* (optional) ensure the materialized view has been updated */
flush;

/* the materialized view should reflect the changes in source table */
select * from mv1;
EOF
}

function test::risingwave::storage_support::_run_with_manifest() {
  logging::info "Starting RisingWave..."
  if ! test::risingwave::start "$1"; then
    return 1
  fi
  logging::info "Started!"

  if ! test::risingwave::check_status_with_simple_queries; then
    logging::error "Queries run against storage failed!"
    return 1
  else
    logging::info "Queries succeeded!"
  fi

  logging::info "Stopping RisingWave..."
  test::risingwave::stop "$1"
  logging::info "Stopped!"
}

function test::run::risingwave::storage_support::meta_memory_object_memory() {
  test::risingwave::storage_support::_run_with_manifest storages/meta-memory-object-memory.yaml
}

function test::run::risingwave::storage_support::meta_etcd() {
  test::risingwave::storage_support::_run_with_manifest storages/meta-etcd.yaml
}

function test::run::risingwave::storage_support::object_minio() {
  test::risingwave::storage_support::_run_with_manifest storages/object-minio.yaml
}

function test::run::risingwave::openkruise_integration() {
  logging::info "Starting RisingWave..."
  if ! test::risingwave::start storages/meta-memory-object-memory.yaml; then
    return 1
  fi

  if [ "${OPEN_KRUISE_ENABLED_IN_RISINGWAVE}" -eq 1 ]; then
    if k8s::kubectl::object_exists deployments "${E2E_RISINGWAVE_NAME}-frontend"; then
      logging::error "Deployment objects should not exist when OpenKruise enabled."
      return 1
    fi
    logging::info "OpenKruise integration succeeded."
  else
    if k8s::kubectl::object_exists cloneset "${E2E_RISINGWAVE_NAME}-frontend"; then
      logging::error "CloneSet objects should not exist when OpenKruise disabled."
      return 1
    fi
  fi
}

function test::run::risingwave::connector_test() {
  logging::info "Starting RisingWave..."
  if ! test::risingwave::start connector/connector-test.yaml; then
    return 1
  fi
  logging::info "Started!"

  logging::info "Stopping RisingWave..."
  test::risingwave::stop connector/connector-test.yaml
  logging::info "Stopped!"
}

# Export the test case only when the required parameters exists.
if [[ -v "E2E_AWS_ACCESS_KEY_ID" && -v "E2E_AWS_SECRET_ACCESS_KEY_ID" && -v "E2E_AWS_S3_REGION" && -v "E2E_AWS_S3_BUCKET" ]]; then
  function test::run::risingwave::storage_support::object_aws_s3() {
    test::risingwave::storage_support::_run_with_manifest storages/object-aws-s3.yaml
  }
else
  logging::warn "Test case \"risingwave::storage_support::object_aws_s3\" is disabled due to lack of parameters!"
fi
