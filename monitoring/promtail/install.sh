#!/bin/bash
#
# Copyright 2023 RisingWave Labs
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
        echo "This script installs the promtail stack stack"
        echo ""
        echo "Usage:"
        echo "$0 [-h] [-d] [-n <namespace>]"
        echo ""
        echo "-d    Dry-run. Print what would be done without executing"
        echo "-h    Show this help message"
        echo "-n    The namespace in which to install the monitoring stack. Defaults to 'monitoring'"
    } 1>&2

    exit 1
}

dry=false
ns="monitoring"

while getopts ":n:dh" o; do
    case "${o}" in
        d)
            dry=true
            ;;
        h)
            usage
            ;;
        n)
            ns=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

dryParam=""
if [[ $dry = true ]]; then 
    echo "Dry-run modus activated in $0"
    dryParam="--dry-run"
fi

_SCRIPT_BASEDIR=$(dirname "$0")

helm --namespace $ns upgrade --install --create-namespace promtail grafana/promtail -f "${_SCRIPT_BASEDIR}"/loki-promtail-clients.yaml $dryParam
