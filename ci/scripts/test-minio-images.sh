#!/usr/bin/env bash

# Copyright 2026 RisingWave Labs
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

set -Eeuo pipefail

# Run against a disposable cluster with the RisingWave CRD installed. Convert
# the original manifests with kubectl, but deploy only their MinIO resources.
export E2E_NAMESPACE=minio-image-test
export E2E_RISINGWAVE_NAME=risingwave
export E2E_RISINGWAVE_IMAGE=unused
namespace_created=false

function cleanup() {
	if ${namespace_created}; then
		kubectl delete namespace "${E2E_NAMESPACE}" --ignore-not-found --wait=true
		namespace_created=false
	fi
}

function diagnose() {
	kubectl -n "${E2E_NAMESPACE}" get pods,pvc || true
	kubectl -n "${E2E_NAMESPACE}" describe pod minio-0 || true
	kubectl -n "${E2E_NAMESPACE}" logs minio-0 || true
}

trap cleanup EXIT
trap diagnose ERR

function connect_minio() {
	# Expand credentials inside the container.
	# shellcheck disable=SC2016
	kubectl -n "${E2E_NAMESPACE}" exec minio-0 -- /bin/sh -ec \
		'mc --config-dir /tmp/mc alias set local http://127.0.0.1:9301 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"' >/dev/null
}

function mc() {
	kubectl -n "${E2E_NAMESPACE}" exec -i minio-0 -- \
		/usr/bin/mc --config-dir /tmp/mc "$@"
}

mapfile -t manifests < <(git grep -l 'image: cgr.dev/chainguard/minio@sha256:' -- docs/manifests test/e2e/tests)
(( ${#manifests[@]} > 0 ))

for manifest in "${manifests[@]}"; do
	echo "Testing MinIO in ${manifest}"
	kubectl create namespace "${E2E_NAMESPACE}"
	namespace_created=true
	# Limit substitution to the manifest placeholders.
	# shellcheck disable=SC2016
	objects=$(envsubst '${E2E_NAMESPACE} ${E2E_RISINGWAVE_NAME} ${E2E_RISINGWAVE_IMAGE}' <"${manifest}" |
		kubectl create --dry-run=client --validate=false -f - -o json)
	bucket=$(jq -er '.items[] | select(.kind == "RisingWave") | .spec.stateStore.minio.bucket' <<<"${objects}")
	jq '{apiVersion: "v1", kind: "List", items: [.items[] |
    select(.metadata.name == "minio" or .metadata.name == "minio-credentials") |
    .metadata.namespace = env.E2E_NAMESPACE]}' <<<"${objects}" | kubectl apply -f -
	kubectl -n "${E2E_NAMESPACE}" rollout status statefulset/minio --timeout=180s
	connect_minio
	# The manifest must create the bucket before RisingWave first uses it.
	mc stat "local/${bucket}"
	printf 'minio-image-test' | mc pipe "local/${bucket}/image-test"
	[[ $(mc cat "local/${bucket}/image-test") == minio-image-test ]]
	# The persistent example must also retain objects across pod replacement.
	if jq -e '.items[] | select(.kind == "StatefulSet" and .metadata.name == "minio") |
    .spec.volumeClaimTemplates | length > 0' <<<"${objects}" >/dev/null; then
		kubectl -n "${E2E_NAMESPACE}" delete pod minio-0 --wait=true
		kubectl -n "${E2E_NAMESPACE}" rollout status statefulset/minio --timeout=180s
		connect_minio
		[[ $(mc cat "local/${bucket}/image-test") == minio-image-test ]]
	fi
	cleanup
done
