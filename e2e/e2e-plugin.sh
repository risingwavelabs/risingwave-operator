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

BASEDIR=$(dirname "$0")
NAME_SPACE="plugin-e2e"
NAME="example-rw"

source "$BASEDIR"/k8s/kubernetes
source "$BASEDIR"/env-utils
source "$BASEDIR"/job/lib

prepare_cluster

function install_operator() {
  echo "Install risingwave operator"
  do_command "kubectl rw install"
}

function create_instance() {
  namespace=$1
  name=$2
  echo "Creating instance $name in $namespace ..."
  do_command "kubectl rw create $name -n $namespace"
}

function restart_instance() {
  namespace=$1
  name=$2
  echo "Restarting instance $name in $namespace ..."
  do_command "kubectl rw restart $name -n $namespace"
}

function stop_instance() {
  namespace=$1
  name=$2
  echo "Stopping instance $name in $namespace ..."
  do_command "kubectl rw stop $name -n $namespace"
}

function resume_instance() {
  namespace=$1
  name=$2
  echo "Resuming instance $name in $namespace ..."
  do_command "kubectl rw resume $name -n $namespace"
}

function update_instance() {
  namespace=$1
  name=$2
  echo "Updating instance $name in $namespace ..."
  do_command "kubectl rw update $name -cr 200m -cl 1000m -n $namespace"
}

function do_command() {
    eval $1
    if [ $? -ne 0 ]; then
      echo "command ${1} failed"
      exit 1
    fi
}

function e2e-plugin() {
    install_operator

    # test create
    kubectl delete namespace $NAME_SPACE || true
    do_command "kubectl create namespace $NAME_SPACE"
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
e2e-plugin