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

_SCRIPT_BASEDIR=$(dirname "$0")

cd "${_SCRIPT_BASEDIR}" || exit

# Set up the helm repos and update
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install the `kube-prometheus-stack`
./kube-prometheus-stack/install.sh

# Install the `loki-distributed`
./loki-distributed/install.sh

# Install the `promtail`
./promtail/install.sh
