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

set -e

echo "begin plugin e2e"
BASEDIR=$(dirname "$0")
K8S_ENV_LIST=$(ls -C "$BASEDIR"/envs)
NAME_SPACE="plugin-e2e"
NAME="example-rw"

source "$BASEDIR"/cluster.sh
source "$BASEDIR"/util.sh
source "$BASEDIR"/k8s/kubernetes

prepare_cluster

function create_instance() {
  namespace=$1
  name=$2
  echo "Creating instance $name in $namespace ..."
  kubectl rw create $name -n $namespace
}

function restart_instance() {
  namespace=$1
  name=$2
  echo "Restarting instance $name in $namespace ..."
  kubectl rw restart $name -n $namespace
}

function stop_instance() {
  namespace=$1
  name=$2
  echo "Stopping instance $name in $namespace ..."
  kubectl rw stop $name -n $namespace
}

function resume_instance() {
  namespace=$1
  name=$2
  echo "Resuming instance $name in $namespace ..."
  kubectl rw resume $name -n $namespace
}

function update_instance() {
  namespace=$1
  name=$2
  echo "Updating instance $name in $namespace ..."
  kubectl rw update $name -cr 200m -cl 1000m -n $namespace
}

function e2e-plugin() {
    ##TODO: add test for plugin install & uninstall
    echo "Install risingwave operator"
    kubectl rw install

    # test create
    kubectl create namespace $NAME_SPACE
    create_instance $NAME_SPACE $NAME
    sleep 1
    wait_until_service_ready $NAME_SPACE ${NAME}-frontend

    # test restart
    pre_pod_name=`kubectl get po -n $NAME_SPACE | grep "frontend"|awk '{print $1}'`
    echo "Before restart, frontend pod name: $pre_pod_name"
    restart_instance $NAME_SPACE $NAME
    sleep 1
    wait_until_service_ready $NAME_SPACE ${NAME}-frontend
    post_pod_name=`kubectl get po -n $NAME_SPACE | grep "frontend"|awk '{print $1}'`
    echo "After restart, frontend pod name: $post_pod_name"

    if [ $pre_pod_name = $post_pod_name ]; then
        echo "Restart instance failed"
        return 1
    fi

    # test stop & resume
    pre_rs=`kubectl get risingwave -n $NAME_SPACE $NAME -o=jsonpath="{.spec.components.compute.groups[0].replicas}"`
    echo "Before stop, compute node rs is $pre_rs"
    stop_instance $NAME_SPACE $NAME
    sleep 1
    rs=`kubectl get risingwave -n $NAME_SPACE $NAME -o=jsonpath="{.spec.components.compute.groups[0].replicas}"`

    if [ ! -n "$rs" ]; then
      echo "Instance has stopped!!"
    fi

    resume_instance $NAME_SPACE $NAME
    sleep 1
    post_rs=`kubectl get risingwave -n $NAME_SPACE $NAME -o=jsonpath="{.spec.components.compute.groups[0].replicas}"`
    echo "After resume, compute node rs is $post_rs"
    if [ $pre_rs -eq $post_rs ]; then
      echo "Succeed to resume instance"
    else
      echo "Resume instance failed!!"
      return 1
    fi

    ##TODO: add test cases for update & upgrade

    echo "All tests are passed!"
}

echo "Running plugin e2e test"
background "e2e-plugin"