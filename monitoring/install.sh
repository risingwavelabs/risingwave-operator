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
        echo "This script installs the monitoring stack"
        echo ""
        echo "Usage:"
        echo "$0 [-h] [-r] [-d] [-k <aws_access_key>] [-s <aws_secret_key>]"
        echo ""
        echo "-h    Show this help message"
        echo "-r    Enable prometheus remote write (AWS). Requires that -k and -s are set"
        echo "-k    AWS access key"
        echo "-s    AWS secret key"
        echo "-d    Dry-run. Print what would be done without executing"
    } 1>&2

    exit 1
}

# TODO: add dryrun param
# TODO: Is it secure to pass the secret key via the command line? Or should we pass this via an env var?

r=false
dry=false

while getopts ":k:s:rd" o; do
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
        d)
            dry=true
            ;;
        h)
            usage
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

if [[ $dry = true ]]; then 
    echo "Dry-run modus activated"
    echo "Would add helm repositories if not in dry-mode"
fi

_SCRIPT_BASEDIR=$(dirname "$0")

cd "${_SCRIPT_BASEDIR}" || exit

# Set up the helm repos and update
if [[ $dry = false ]]; then 
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo update
fi

dryParam=""
if [[ $dry = true ]]; then 
    dryParam=" -d "
fi

# Install the `kube-prometheus-stack` locally
rParams=""
if [[ $r = true ]]; then
    # Install the `kube-prometheus-stack` with remote write
    rParams=" -r -s $s -k $k "
fi 
./kube-prometheus-stack/install.sh $rParams $dryParam

exit # TODO: remove this line
# TODO: Add dry run flags to other install scripts

# Install the `loki-distributed`
./loki-distributed/install.sh

# Install the `promtail`
./promtail/install.sh
