#!/usr/bin/env bash

set -euo pipefail

if [[ "$CI_ENV" = "1" ]]; then
  echo "--- Mark /workspace as safe directory (VCS)"
  git config --global --add safe.directory /workspace
fi

echo "--- Ensure that the kind cluster is deleted"

kind delete cluster --name e2e

echo "--- Running e2e tests"

make e2e-test
