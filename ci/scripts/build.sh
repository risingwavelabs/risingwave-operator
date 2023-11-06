#!/usr/bin/env bash

set -euo pipefail

if [[ "$CI_ENV" = "1" ]]; then
  echo "--- Mark /workspace as safe directory (VCS)"
  git config --global --add safe.directory /workspace
fi

echo "--- Build binaries"

make build
