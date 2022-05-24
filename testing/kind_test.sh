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
OPERATOR_DEPLOYMENT=config/risingwave-operator-test.yaml
RW_DEPLOYMENT=$TESTING_DIR/rw.yaml
NAMESPACE=e2e-rw
OPERATOR_IMG=singularity-data/risingwave-operator:dev

if [ "$TAG" == "" ]; then
    IMG_TAG=latest
fi

echo $TESTING_DIR/utils.sh
source $TESTING_DIR/utils.sh

function kind_test() {
    # prepare kind cluster
    delete_kind_cluster
    start_kind_cluster $TESTING_DIR/kind-config.yaml

    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
    wait_cert_manager

    kind load docker-image $OPERATOR_IMG
    kubectl apply -f $OPERATOR_DEPLOYMENT
    wait_rw_operator $OPERATOR_DEPLOYMENT
    sleep 30

    echo 'Deploying risingwave...'
    deploy $NAMESPACE $RW_DEPLOYMENT
    wait_risingwave $NAMESPACE    
}

kind_test
