#!/usr/bin/env bash

set -euo pipefail

exit_code=0

function mark_fail() {
  echo $1
  exit_code=$2
}

echo "--- Running spell checks"

make spellcheck || mark_fail "Spell check failed" $?

echo "--- Running shell checks"

make shellcheck || mark_fail "Shell check failed" $?

echo "--- Running go mod tidy"

function go_mod_tidy_check() {
  local exit_code=0
  local go_mod_tidy_output
  go_mod_tidy_output=$(go mod tidy -v 2>&1) || exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    echo "go mod tidy failed:"
    echo "$go_mod_tidy_output"
    return 1
  fi
  if echo "$go_mod_tidy_output" | grep -q -v "go: downloading"; then
    echo "go mod tidy modified go.mod or go.sum:"
    echo "$go_mod_tidy_output"
    return 1
  fi
  return 0
}

go_mod_tidy_check || mark_fail "go mod tidy failed" $?

echo "--- Running go mod vendor"

go mod vendor || mark_fail "go mod vendor failed" $?

echo "--- Running go fmt"

function go_fmt_check() {
  local exit_code=0
  local go_fmt_output
  go_fmt_output=$(go fmt ./... 2>&1) || exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    echo "go fmt failed:"
    echo "$go_fmt_output"
    return 1
  fi
  if [[ -n "$go_fmt_output" ]]; then
    echo "go fmt modified files:"
    echo "$go_fmt_output"
    return 1
  fi
  return 0
}

go_fmt_check || mark_fail "go fmt failed" $?

echo "--- Generating manifests"

make manifests || mark_fail "Manifest generation failed" $?

echo "--- Generating YAML files"

make generate-yaml generate-test-yaml || mark_fail "YAML generation failed" $?

function yaml_change_check() {
  if git status --porcelain | grep -q 'config/risingwave-operator.*.yaml'; then
    echo "YAML files changed, please run 'make generate-yaml generate-test-yaml' and commit the changes"
    return 1
  fi
}

yaml_change_check || mark_fail "YAML files changed" $?

echo "--- Generating docs"

make generate-docs || mark_fail "Docs generation failed" $?

function docs_change_check() {
  if git status --porcelain | grep -q docs/; then
    echo "Docs changed, please run 'make generate-docs' and commit the changes"
    return 1
  fi
}
docs_change_check || mark_fail "Docs changed" $?

echo "--- Running lint checks"

make lint || mark_fail "Lint failed" $?

exit $exit_code