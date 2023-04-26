#!/usr/bin/env bash

set -euo pipefail

docker run --rm -v "$(pwd):/workspace" -w /workspace -v /var/run/docker.sock:/var/run/docker.sock --network host --entrypoint bash "${CI_BASE_IMAGE}" "$@"