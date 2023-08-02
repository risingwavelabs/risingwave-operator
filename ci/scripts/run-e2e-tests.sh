#!/usr/bin/env bash

set -euo pipefail

echo "--- Ensure that the kind cluster is deleted"

kind delete cluster --name e2e

echo "--- Running e2e tests"

make e2e-test
