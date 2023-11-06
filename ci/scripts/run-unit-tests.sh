#!/usr/bin/env bash

set -euo pipefail

if [[ "$CI_ENV" = "1" ]]; then
  echo "--- Mark /workspace as safe directory (VCS)"
  git config --global --add safe.directory /workspace
fi

echo "--- Running unit tests"

go test ./... -coverprofile cover.out.tmp || exit $?
grep -v "_generated.go" <cover.out.tmp >cover.out && rm -f cover.out.tmp

echo "--- Report coverage to codecov.io"
codecov -t "$CI_CODECOV_TOKEN" -f cover.out -F unittests -v
