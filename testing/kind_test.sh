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
TESTING_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

set -e

RW_DEPLOYMENT=rw.yaml
OPERATOR_DEPLOYMENT=config/risingwave-operator.yaml
NAMESPACE=e2e-rw

if [ "$TAG" == "" ]; then
    IMG_TAG=latest
fi

source $TESTING_DIR/utils.sh

function kind_test() {
    # prepare kind cluster
    delete_kind_cluster
    start_kind_cluster $TESTING_DIR/kind-config.yaml

    make install
    make run-local &
    export tmp_test_operator_pid=$!

    deploy $NAMESPACE $RW_DEPLOYMENT
    wait_risingwave $RW_DEPLOYMENT

    check_event_logs
}

function end_test() {
    if [ "$tmp_test_operator_pid" != "" ]; then
        kill tmp_test_operator_pid
    fi
    check_namespace=$(kubectl get namespace | grep $NAMESPACE)
    if [ "$check_namespace" != "" ]; then
        kubectl delete all --all -n $NAMESPACE
        kubectl delete namespace $NAMESPACE
    fi
}

# kind_test
# end_test
