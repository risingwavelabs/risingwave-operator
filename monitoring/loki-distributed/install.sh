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
        echo "This script installs the loki stack stack"
        echo ""
        echo "Usage:"
        echo "$0 [-h] [-d]"
        echo ""
        echo "-h    Show this help message"
        echo "-d    Dry-run. Print what would be done without executing"
    } 1>&2

    exit 1
}

dry=false

while getopts ":dh" o; do
    case "${o}" in
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

dryParam=""
if [[ $dry = true ]]; then 
    echo "Dry-run modus activated in $0"
    dryParam="--dry-run"
fi

helm --namespace monitoring upgrade --install --create-namespace loki grafana/loki-distributed $dryParam
