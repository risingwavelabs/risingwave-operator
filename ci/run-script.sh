#!/usr/bin/env bash

set -euo pipefail

mkdir -p ci/run && env | grep "BUILDKITE\|CI\|AWS"> ci/run/env.list

docker run --rm --userns=host --privileged --env-file ci/run/env.list -v "$(pwd):/workspace" -w /workspace -v /var/run/docker.sock:/var/run/docker.sock --network host --entrypoint bash "${CI_BASE_IMAGE}" "$@"