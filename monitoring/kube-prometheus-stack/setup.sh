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

# install the necessary helm charts if not installed

# _SCRIPT_BASEDIR=$(dirname "$0")
# CLUSTER_NAME="onebox"
# cluster="kind"
# RISINGWAVE_OPERATOR_NAMESPACE="risingwave-operator-system"
# PROMETHEUS_NAMESPACE="default"
# PROMETHEUS_RELEASE_NAME="prometheus"
# REMOTE_WRITE_ADAPTER_NAME="prometheus-remote-write-adapter"
# ZOOKEEPER_NAME="zkp"
# KAFKA_NAME="kstack"
# KAFKA_UI_NAME="kafka-ui"
# KAFKA_UI_PORT=8081
# MONITORING_TENANT_NAME="monitoring"

# if [ "$(kind get clusters | grep ${CLUSTER_NAME})" == "" ]; then
#     echo "Ensure the cluster ${CLUSTER_NAME} is up and running!"
#     exit 1
# fi 

# if ! [ -x "$(command -v rwc 2>&-)" ]; then
#     echo "Ensure that the rwc cli is installed"
#     exit 1
# fi

# if ! kubectl get secrets | grep -E '^aws-prometheus-credentials'>/dev/null; then
#     echo "Ensure that the secret aws-prometheus-credentials is set"
#     exit 1
# fi

# load_image_in_cluster() {
#     set -e 
#     repo=$1 
#     tag=$2
#     if [[ "$(docker images -q $repo:$tag 2> /dev/null)" == "" ]]; then
#         echo "Pulling image: ${repo}:${tag}"
#         docker pull $repo:$tag
#     fi
#     kind load docker-image ${repo}:${tag} --name ${CLUSTER_NAME}
#     set +e
# }

# ### Deploy Promtheus if not deployed ###
# if [ "$(kubectl get pods --namespace "${PROMETHEUS_NAMESPACE}" | grep "${PROMETHEUS_RELEASE_NAME}")" == "" ]; then
#     echo "Creating prometheus deployment..."
#     helm upgrade --install "${PROMETHEUS_RELEASE_NAME}" prometheus-community/kube-prometheus-stack \
#     -f https://raw.githubusercontent.com/singularity-data/risingwave-operator/main/monitoring/kube-prometheus-stack/kube-prometheus-stack.yaml \
#     -f "${_SCRIPT_BASEDIR}"/prometheus-remote-write-aws.yaml
# fi

# ### Deploy Read Write adapter if not deployed ###
# if [ "$(kubectl get pods --namespace "${RISINGWAVE_OPERATOR_NAMESPACE}" | grep "${REMOTE_WRITE_ADAPTER_NAME}")" == "" ]; then
#     echo "Creating adapter deployment..."
#     kubectl apply -f "${_SCRIPT_BASEDIR}"/remote-write-adapter.yaml
# fi

# ### Deploy zookeeper if not deployed ###
# if [ "$(kubectl get pods --namespace ${RISINGWAVE_OPERATOR_NAMESPACE} | grep ${ZOOKEEPER_NAME})" == "" ]; then
#     echo "Loading zookeeper in cluster..."
#     helm upgrade --install "${ZOOKEEPER_NAME}" --namespace ${RISINGWAVE_OPERATOR_NAMESPACE} -f ./zookeeper-deployment.yaml rhcharts/zookeeper --wait --timeout 20m
# fi

# ### Connect kafka to zookeeper ###
# if [ "$(kubectl get pods --namespace "${RISINGWAVE_OPERATOR_NAMESPACE}" | grep "${KAFKA_NAME}")" == "" ]; then
#     echo "Loading kafka in cluster..."
#     helm upgrade --install "${KAFKA_NAME}" --namespace "${RISINGWAVE_OPERATOR_NAMESPACE}" -f ./kafka-deployment.yaml rhcharts/kafka --wait --timeout 20m
# fi

# ### Install kafka ui for monitoring ###
# if [ "$(kubectl get pods --namespace "${RISINGWAVE_OPERATOR_NAMESPACE}" | grep "${KAFKA_UI_NAME}")" == "" ]; then
#     echo "Loading kafka UI in cluster..."
#     helm upgrade --install "${KAFKA_UI_NAME}" kafka-ui/kafka-ui -f kafka-ui.yaml --namespace="${RISINGWAVE_OPERATOR_NAMESPACE}" --wait --timeout 2m
#     ulimit -n 65536	
#     echo "You can access the kafka ui at port "${KAFKA_UI_PORT}""
# fi

echo "Login to CLI required..."
rwc config -region local
if [ $? -ne 0 ]; then 
    echo "Could not configure the rwc cli instance"
    exit 1
fi
echo "Configured rwc cli instance"
PS3="Have you registered for a local account yet: "
select registered in yes no; do
    read -p "Enter your username: " rwc_username
    read -s -p "Enter your password: " rwc_password
    case $registered in
        yes)
            rwc login -account ${rwc_username} -password ${rwc_password}
            break;;
        no) 
            rwc register -account ${rwc_username} -password ${rwc_password}
            break;;
        *)
            echo "Invalid option: $REPLY";;
    esac
done  
if [ $? -ne 0 ]; then
    echo "Could not login to the risingwave cli"
    exit 1 
fi

$namespace="$(kubectl get namespace | grep -o "^\S\+" | grep ${MONITORING_TENANT_NAME})"

if [ "$(kubectl get pods --namespace "${namespace}")" != "" ]; then
    echo "existing monitoring tenant detected. cleaning up"
    kubectl delete pods --namespace="${namespace}"
    kubectl delete deployments --namespace="${namespace}"
    echo "cleaning up existing resources..."
    kubectl wait --for delete pods --namespace="${namespace}" --timeout=5m
    kubectl wait --for delete deployments --namespace="${namespace}" --timeout=5m
    if [ $? -ne 0 ]; then
        echo "Could not delete all deployments in namespace ${namespace}"
        exit 1 
    fi
fi

sleep 1 # wait for credentials to be persisted

echo "creating monitoring tenant..."
rwc tenant create -name "${namespace}"

# get namespace of risingwave tenant created

if [ $namespace == "" ]; then
    echo "Error! Namespace of tenant not found..."
    exit 1 
fi

kubectl wait --for=condition=Ready pods --namespace "${namespace}" --timeout=5m
psql "$(rwc tenant endpoint -name "${MONITORING_TENANT_NAME}")"
if [ $? -ne 0 ]; then
    echo "Could not connect to tenant "${MONITORING_TENANT_NAME}""
    exit 1 
fi
# kubectl port-forward --namespace risingwave-operator-system svc/kafka-ui "${KAFKA_UI_PORT}":80




