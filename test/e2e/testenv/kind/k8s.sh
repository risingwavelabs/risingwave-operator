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

${__E2E_SOURCE_TESTENV_KIND_K8S_SH__:=false} && return 0 || __E2E_SOURCE_TESTENV_KIND_K8S_SH__=true

_TEST_ENV_KIND_DIR=$(dirname "${BASH_SOURCE[0]}")
_TEST_ENV_KIND_CLUSTER_NAME="e2e"
_TEST_ENV_KIND_DEFAULT_IMAGE="kindest/node:v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1"

source "${_TEST_ENV_KIND_DIR}/../../common/lib.sh"

function testenv::k8s::kind::_kind_cluster_exists() {
  kind get clusters -q | grep -w -q "${_TEST_ENV_KIND_CLUSTER_NAME}"
}

#######################################
# Start a Kind cluster with the local config file.
# Globals
#   KIND_CLUSTER_IMAGE, defaults to ${_TEST_ENV_KIND_DEFAULT_IMAGE}
# Arguments
#   None
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::kind::provision() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(testenv "k8s/kind: ${_TEST_ENV_KIND_CLUSTER_NAME}")

  local kind_image=${KIND_CLUSTER_IMAGE:-${_TEST_ENV_KIND_DEFAULT_IMAGE}}

  if ! testenv::k8s::kind::_kind_cluster_exists; then
    logging::info "Start the Kind cluster..."
    if ! shell::run kind create cluster \
      --name="${_TEST_ENV_KIND_CLUSTER_NAME}" \
      --image="${kind_image}" \
      --wait=300s \
      --config="${_TEST_ENV_KIND_DIR}/config.yaml"; then
      logging::error "Failed to start!"
      return 1
    fi
    logging::info "Started!"
  else
    logging::warn "Kind cluster found, skip!"
  fi
}

#######################################
# Stop the running Kind cluster if there's one.
# Globals
#   None
# Arguments
#   None
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::kind::teardown() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(testenv "k8s/kind: ${_TEST_ENV_KIND_CLUSTER_NAME}")

  if testenv::k8s::kind::_kind_cluster_exists; then
    logging::info "Stop the Kind cluster..."
    if ! shell::run kind delete cluster --name="${_TEST_ENV_KIND_CLUSTER_NAME}"; then
      logging::error "Failed to stop!"
      return 1
    fi
    logging::info "Stopped!"
  else
    logging::warn "Kind cluster not found, skip!"
  fi
}

#######################################
# Load local Docker images into the Kind cluster.
# Globals
#   None
# Arguments
#   Local Docker image names
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::kind::load_docker_image() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(testenv "k8s/kind: ${_TEST_ENV_KIND_CLUSTER_NAME}")

  if ! testenv::k8s::kind::_kind_cluster_exists; then
    logging::error "Kind cluster doesn't exist!"
    return 1
  fi

  local image_names_in_lines=""
  IFS=$'\n  ' image_names_in_lines="$*" && unset IFS

  logging::infof "Loading local Docker images:\n  %s\n" "${image_names_in_lines}"

  if shell::run kind load docker-image --name="${_TEST_ENV_KIND_CLUSTER_NAME}" "$@"; then
    logging::info "Successfully loaded!"
  else
    logging::error "Failed to load!"
  fi
}

#######################################
# Load local image archive into the Kind cluster.
# Globals
#   None
# Arguments
#   Image archive file.
# Returns
#   0 on successful, non-zero otherwise.
#######################################
function testenv::k8s::kind::load_image_archive() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(testenv "k8s/kind: ${_TEST_ENV_KIND_CLUSTER_NAME}")

  if ! testenv::k8s::kind::_kind_cluster_exists; then
    logging::error "Kind cluster doesn't exist!"
    return 1
  fi

  logging::infof "Loading image archive:\n  %s\n" "$1"

  if shell::run kind load docker-image --name="${_TEST_ENV_KIND_CLUSTER_NAME}" "$1"; then
    logging::info "Successfully loaded!"
  else
    logging::error "Failed to load!"
  fi
}
