#!/bin/bash
#
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
#

BASEDIR=$(dirname "$0")
K8S_ENV_LIST=$(ls -C "$BASEDIR"/envs)

function check_k8s_env() {
  echo "$K8S_ENV_LIST" | grep -w "$1" >/dev/null
}

# Prepare images...
function lazy_pull_image() {
  repo=$1
  tag=${2:-latest}

  if [ "$tag" = "latest" ] || [ -z "$(docker image ls -q "$repo":"$tag")" ]; then
    echo "Pulling image $repo:$tag..."
    docker pull "$repo:$tag"
  else
    echo "Image $repo:$tag already exists, skip"
  fi
}

function prepare_cluster() {
  K8S_ENV="${K8S_ENV:-kind-4}"
  if ! check_k8s_env "$K8S_ENV"; then
    echo "ERROR: env $K8S_ENV not found"
    echo "Available envs: $K8S_ENV_LIST"
    exit 1
  fi

  # Source the env functions.
  echo "Starting environment $K8S_ENV ..."

  # shellcheck disable=SC1090
  source "$BASEDIR"/envs/"$K8S_ENV"/env

  # Start the Kubernetes.
  start_kubernetes_env
  echo "Started!"

  # Set the trap for cleaning the Kubernetes.
  trap stop_kubernetes_env EXIT

  OS="$(uname -s)"
  case "${OS}" in
  Linux*) NIGHTLY_IMAGE_TAG="nightly"-$(date -d "2 day ago" '+%Y%m%d') ;;
  Darwin*) NIGHTLY_IMAGE_TAG="nightly"-$(date -v-2d '+%Y%m%d') ;;
  *) echo "ERROR: unsupported platform $OS" && exit 1 ;;
  esac

  # FIXME: currently the nightly tags aren't continuous.
  NIGHTLY_IMAGE_TAG="nightly-20220727"
  echo "Using a nightly tag $NIGHTLY_IMAGE_TAG for RisingWave images..."

  echo "Pulling images..."
  lazy_pull_image postgres
  lazy_pull_image praqma/network-multitool
  lazy_pull_image ghcr.io/singularity-data/risingwave "$NIGHTLY_IMAGE_TAG"
  docker tag ghcr.io/singularity-data/risingwave:"$NIGHTLY_IMAGE_TAG" ghcr.io/singularity-data/risingwave:e2e
  echo "Pulled!"

  # Load images...
  echo "Loading images..."
  kubernetes_load_image postgres
  kubernetes_load_image praqma/network-multitool
  kubernetes_load_image ghcr.io/singularity-data/risingwave:e2e
  kubernetes_load_image docker.io/singularity-data/risingwave-operator:dev
  echo "Loaded!"
}

