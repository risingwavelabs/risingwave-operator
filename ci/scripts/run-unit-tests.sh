#!/usr/bin/env bash

set -euo pipefail

echo "--- Running unit tests"

go test ./... -coverprofile cover.out.tmp || exit $?
grep -v "_generated.go" <cover.out.tmp >cover.out && rm -f cover.out.tmp

echo "--- Report coverage to codecov.io"
codecov -t "$CODECOV_TOKEN" -f cover.out -F unittests -v