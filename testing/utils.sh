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

if ! command -v jq &> /dev/null
then
    sudo apt update
    sudo apt install jq
fi

function prepare_e2e() {
    kind_config=$1
    delete_kind_cluster
    start_kind_cluster $kind_config
    start_test_pod 
}

function start_kind_cluster() {
    kind create cluster --config $1
}

function delete_kind_cluster() {
    kind delete cluster
    sleep 5
}

function start_test_pod() {
    kubectl apply -f testing/test-pod.yaml
    threshold=40
    current_epoch=0
    while :
    do
        result=$(kubectl get po -A -n default | grep debug-pod | grep Running)
        if [ "$result" != "" ]; then
            break
        fi
        if [ $current_epoch -eq $threshold ]; then
            echo "ERROR: timeout waiting for test pod"
            exit 1
        fi
        current_epoch=$((current_epoch+1))
        echo "waiting for debug-pod ($current_epoch / $threshold)"
        sleep 2
    done
}

function wait_webhook() {
    service_name=$1
    webhook_name=$2
    webhook_ip=$(kubectl get svc -n $service_name | grep $webhook_name  | awk '{print $3}')
    webhook_port_raw=$(kubectl get svc -n $service_name | grep $webhook_name  | awk '{print $5}')
    webhook_port=$(echo ${webhook_port_raw/\/TCP/""})
    echo "$service_name webhook endpoint found $webhook_ip:$webhook_port"
    threshold=40
    current_epoch=0
    while :
    do
        result=$(kubectl exec --stdin --tty debug-pod -n default -- nc -zvw3 $webhook_ip $webhook_port | grep succeeded)
        if [ "$result" != "" ]; then
            break
        fi
        if [ $current_epoch -eq $threshold ]; then
            echo "ERROR: timeout waiting for $service_name $webhook_name webhook."
            exit 1
        fi
        current_epoch=$((current_epoch+1))
        echo "waiting for $service_name $webhook_name webhook to be ready ($current_epoch / $threshold)"
        sleep 2
    done
}

function wait_cert_manager() {
    # wait for certificate
    threshold=40
    current_epoch=0
    while :
    do  
        ca_bundle=$(kubectl get validatingwebhookconfigurations cert-manager-webhook -o yaml | grep caBundle)
        if [ "$ca_bundle" != "" ]; then
            break
        fi
        if [ $current_epoch -eq $threshold ]; then
            echo "ERROR: timeout waiting for cert-manager"
            exit 1
        fi
        sleep 2
        current_epoch=$((current_epoch+1))
        echo "waiting for cert-manager to be ready ($current_epoch / $threshold)..."
    done

    wait_webhook cert-manager cert-manager-webhook
}

function wait_rw_operator() {
    deployment=$1
    current_epoch=0
    threshold=40
    while :
    do  
        result=$(kubectl get -f $deployment -o jsonpath='{.items[*].status.conditions[?(.type == "Ready")]}' | jq .status | awk '{if($1 ==  "\"True\"") s += 1}END{print s == NR}')
        if [ $result -eq 1 ]; then
            break
        fi
        if [ $current_epoch -eq $threshold ]; then
            echo "ERROR: timeout waiting for risingwave-operator-system"
            exit 1
        fi
        sleep 2
        current_epoch=$((current_epoch+1))
        echo "waiting for risingwave-operator-system to be ready ($current_epoch / $threshold)..."
    done

    wait_webhook risingwave-operator-system risingwave-operator-webhook-service
}

function deploy() {
    namespace=$1
    deployment=$2

    kubectl create namespace $namespace
    kubectl apply -f $deployment
}

function wait_risingwave() {
    namespace=$1
    current_epoch=0
    check_times=60
    while :
    do
        echo "Waiting for risingwave..."
        sleep 4
        current_epoch=$((current_epoch+1))
        if [ $current_epoch -eq $threshold ]; then
            echo "ERROR: timeout waiting for risingwave"
            exit 1
        fi
        c=$(check_svc $namespace meta-node)
        if [ "$c" == "0" ];then continue; fi
        c=$(check_svc $namespace compute-node)
        if [ "$c" == "0" ];then continue; fi
        c=$(check_svc $namespace compactor-node)
        if [ "$c" == "0" ];then continue; fi
        c=$(check_svc $namespace frontend)
        if [ "$c" == "0" ];then continue; fi
        break 
    done

    echo "risingwave ($deployment) is ready."
}

function check_svc() {
    namespace=$1
    svc_name=$2
    result=$(kubectl get svc -n $namespace | grep $svc_name)
    if [ "$result" != "" ]; then
        echo 1
    else 
        echo 0
    fi
}
