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

usage() {
    {
        echo "This script installs the kube-prometheus-stack stack"
        echo ""
        echo "Usage:"
        echo "$0 [-r] [-k <aws_access_key>] [-s <aws_secret_key>]"
        echo ""
        echo "-d    Dry-run. Print what would be done without executing"
        echo "-h    Show this help message"
        echo "-k    AWS access key"
        echo "-r    Enable prometheus remote write (AWS). Requires that -k and -s are set"
        echo "-s    AWS secret key"
    } 1>&2

    exit 1
}

# TODO: Is it secure to pass the secret key via the command line? Or should we pass this via an env var?

r=false
dry=false

while getopts ":k:s:rhd" o; do
    case "${o}" in
        k)
            k=${OPTARG}
            ;;
        s)
            s=${OPTARG}
            ;;
        r)
            r=true
            ;;
        h)
            usage
            ;;
        d)
            dry=true
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))


# We require credentials, if we use prometheus remote write
if [[ $r = true ]]; then
    if [ -z "${s}" ] || [ -z "${k}" ]; then
        usage
    fi
fi

dryParam=""
if [[ $dry = true ]]; then 
    echo "Dry-run modus activated in $0"
    dryParam="--dry-run"
fi

_SCRIPT_BASEDIR=$(dirname "$0")

msg="Installing prometheus setup with local source" 
_DATA_SOURCE=grafana-loki-data-source.yaml
if [[ $r = true ]]; then
    msg="Installing prometheus setup with remote source" 
    _DATA_SOURCE=prometheus-remote-write-aws.yaml
fi
echo $msg

helm --namespace monitoring upgrade --install --create-namespace prometheus prometheus-community/kube-prometheus-stack \
  -f "${_SCRIPT_BASEDIR}"/kube-prometheus-stack.yaml \
  -f "${_SCRIPT_BASEDIR}"/${_DATA_SOURCE} \
  $dryParam

# Create secret if required
# TODO: Maybe this needs to be before helm upgrade,
# but then we need to check if the monitoring ns exists first

dryParam="none"
if [[ $dry = true ]]; then 
    dryParam="client"
fi

# Create secret with credentials
kubectl create secret generic aws-prometheus-credentials \
  --from-literal AccessKey=${k} --from-literal SecretAccessKey=${s} \
  --dry-run=$dryParam




