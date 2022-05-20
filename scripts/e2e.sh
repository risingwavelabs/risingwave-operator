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

# cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
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
### check if endpoint is available
webhook_ip=$(kubectl get svc -n cert-manager | grep cert-manager-webhook  | awk '{print $3}')
webhook_port_raw=$(kubectl get svc -n cert-manager | grep cert-manager-webhook  | awk '{print $5}')
webhook_port=$(echo ${webhook_port_raw/\/TCP/""})
echo "cert-manager webhook endpoint found $webhook_ip:$webhook_port"
current_epoch=0
while :
do
    nc -zvw3 $webhook_ip $webhook_port
    nc_exit_code=$?
    if [ $nc_exit_code -eq 0 ]; then
        break
    fi
    if [ $current_epoch -eq $threshold ]; then
        echo "ERROR: timeout waiting for cert-manager webhook."
        exit 1
    fi
    current_epoch=$((current_epoch+1))
    echo "waiting for cert-manager webhook to be ready ($current_epoch / $threshold)"
    sleep 2
done
###
echo "cert-manager is ready."


# risingwave-operator-system
make build
make deploy
current_epoch=0
while :
do  
    result=$(kubectl get -f config/risingwave-operator.yaml -o jsonpath='{.items[*].status.conditions[?(.type == "Ready")]}' | jq .status | awk '{if($1 ==  "\"True\"") s += 1}END{print s == NR}')
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
### check if the webhook endpoint is ready for connection
webhook_ip=$(kubectl get svc -n risingwave-operator-system | grep risingwave-operator-webhook-service  | awk '{print $3}')
webhook_port_raw=$(kubectl get svc -n risingwave-operator-system | grep risingwave-operator-webhook-service  | awk '{print $5}')
webhook_port=$(echo ${webhook_port_raw/\/TCP/""})
echo "risingwave operator webhook endpoint found $webhook_ip:$webhook_port"
current_epoch=0

while :
do
    nc -zvw3 $webhook_ip $webhook_port
    nc_exit_code=$?
    if [ $nc_exit_code -eq 0 ]; then
        break
    fi
    if [ $current_epoch -eq $threshold ]; then
        echo "ERROR: timeout waiting for risingwave-operator webhook."
        exit 1
    fi
    current_epoch=$((current_epoch+1))
    echo "waiting for risingwave-operator webhook to be ready ($current_epoch / $threshold)"
    sleep 2
done
echo "risingwave-operator-system is ready."
###

# risingwave
namespace_exit=$(kubectl get namespaces | awk '{if($1 == "test")s=1}END{print s}')
if [ $namespace_exit -ne 1 ]; then
    kubectl create namespace test
fi
kubectl apply -f examples/minio-risingwave-amd.yaml
current_epoch=0
while :
do  
    # result=$(kubectl get po -n test  -o jsonpath={.items[*].status.conditions[*]} | jq .status | awk '{if($1 ==  "\"True\"") s += 1}END{print s == NR}')
    result=$(kubectl get -f examples/minio-risingwave-amd.yaml -o jsonpath='{.status.conditions[?(.type == "Running")]}' | jq .status | awk '{if($1 ==  "\"True\"") s += 1}END{print s == NR}')
    if [ $result -eq 1 ]; then
        break
    fi
    if [ $current_epoch -eq $threshold ]; then
        echo "ERROR: timeout waiting for risingwave"
        exit 1
    fi
    current_epoch=$((current_epoch+1))
    echo "waiting for risingwave to be ready ($current_epoch / $threshold)"
    sleep 2
done
echo "risingwave is ready."

# checking event log to see if there is some errors
current_epoch=0
check_times=30
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
