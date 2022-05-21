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
RW_DEPLOYMENT=examples/minio-risingwave-amd.yaml
OPERATOR_DEPLOYMENT=config/risingwave-operator.yaml
if [ "$TAG" == "" ]; then
    IMG_TAG=latest
fi

source ./utils.sh

function kind_test() {
    # prepare kind cluster
    delete_kind_cluster
    start_kind_cluster ./kind-config.yaml

    # apply cert manager
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
    wait_cert_manager

    make install
    wait_rw_operator $OPERATOR_DEPLOYMENT

    deploy "test" ../examples/minio-risingwave-amd.yaml
    wait_risingwave $RW_DEPLOYMENT

    check_event_logs
}
