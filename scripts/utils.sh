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

function start_kind_cluster() {
    kind create cluster --config $1
}

function delete_kind_cluster() {
    kind delete cluster
}

function wait_webhook() {
    service_name = $1
    webhook_name = $2
    webhook_ip=$(kubectl get svc -n $service_name | grep $service_name  | awk '{print $3}')
    webhook_port_raw=$(kubectl get svc -n $service_name | grep $service_name  | awk '{print $5}')
    webhook_port=$(echo ${webhook_port_raw/\/TCP/""})
    echo "cert-manager webhook endpoint found $webhook_ip:$webhook_port"
    threshold=40
    current_epoch=0
    while :
    do
        nc -zvw3 $webhook_ip $webhook_port
        nc_exit_code=$?
        if [ $nc_exit_code -eq 0 ]; then
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
    deployment = examples/minio-risingwave-amd.yaml
    current_epoch=0
    threshold=40
    while :
    do  
        result=$(kubectl get -f $deployment -o jsonpath='{.status.conditions[?(.type == "Running")]}' | jq .status | awk '{if($1 ==  "\"True\"") s += 1}END{print s == NR}')
        if [ $result -eq 1 ]; then
            break
        fi
        if [ $current_epoch -eq $threshold ]; then
            echo "ERROR: timeout waiting for risingwave ($deployment)"
            exit 1
        fi
        current_epoch=$((current_epoch+1))
        echo "waiting for risingwave ($deployment) to be ready ($current_epoch / $threshold)"
        sleep 2
    done
    echo "risingwave ($deployment) is ready."
}

function check_event_logs() {
    # checking event log to see if there is some errors
    current_epoch=0
    check_times=20
    while :
    do 
        failed_event=$(kubectl get events -A  | awk '{if($3 == "Failed")print $0}')
        if [ "$failed_event" != "" ]; then
            echo "Failed events found in the system"
            echo $failed_event
            exit 1
        fi
        if [ $current_epoch -eq $check_times ]; then
            break
        fi
        current_epoch=$((current_epoch+1))
        echo "checking failed event ($current_epoch / $check_times)"
        sleep 1
    done
}

