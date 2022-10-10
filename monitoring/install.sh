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
        echo "$0 [-r] [-k <aws_access_key>] [-s <aws_secret_key>]"
        echo ""
        echo "-r    Enable prometheus remote write (AWS). Requires that -k and -s are set"
        echo "-k    AWS access key"
        echo "-s    AWS secret key"
    } 1>&2

    exit 1
}


r=false

while getopts ":k:s:r" o; do
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
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

echo "k = ${k}"
echo "s = ${s}"
echo "r = ${r}"

# We require credentials, if we use prometheus remote write
if [[ $r = true ]]; then
    if [ -z "${s}" ] || [ -z "${k}" ]; then
        usage
    fi
fi

# TODO use k and s if r is set
# pass to kube-prometheus-stack/install.sh
exit 1

_SCRIPT_BASEDIR=$(dirname "$0")

cd "${_SCRIPT_BASEDIR}" || exit

# Set up the helm repos and update
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install the `kube-prometheus-stack`
./kube-prometheus-stack/install.sh # pass parameters here

# Install the `loki-distributed`
./loki-distributed/install.sh

# Install the `promtail`

./promtail/install.sh