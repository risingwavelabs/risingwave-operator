#!/usr/bin/env bash

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

set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/common/lib.sh"
source "$(dirname "${BASH_SOURCE[0]}")/testenv/lib.sh"
source "$(dirname "${BASH_SOURCE[0]}")/tests/tests.sh"

#######################################
# Set all debug related environment variables to true, and set the logging level to debug.
# Globals
#   TRACE_COMMAND
#   SHOW_COMMAND_OUTPUT
# Arguments
#   None
# Returns
#   0
#######################################
function e2e::turn_on_debug_settings_if_debug_is_true() {
  if [[ -v "DEBUG" && ("${DEBUG}" == "1" || "${DEBUG}" == "true") ]]; then
    logging::set_level debug

    export TRACE_COMMAND=true
    export SHOW_COMMAND_OUTPUT=true

    # Unset the DEBUG to avoid unexpected behaviors, such as the too detailed logs from `kubectl`.
    unset DEBUG
  fi
}

function e2e::list_test_cases() {
  local -a functions
  IFS=$'\n' read -d '' -ra functions <<<"$(compgen -A function | sort)" && unset IFS

  local -a testcases=()
  local testcase
  for f in "${functions[@]}"; do
    if [[ "${f}" =~ test::run:: ]]; then
      testcase=${f#test::run::}

      # Skip tests if prefix is defined.
      if [[ -v "E2E_TEST_CASE_PREFIX" && "${testcase}" != "${E2E_TEST_CASE_PREFIX}"* ]]; then
        continue
      fi

      testcases+=("${testcase}")
    fi
  done

  echo -n "${testcases[*]}"
}

function e2e::test::_ns() {
  local ns="${1/::/-}"
  ns="${ns/_/-}"
}

function e2e::test::pre_run() {
  local ns="$1"
  local tc="$2"

  if ! shell::run kubectl create ns "${ns}"; then
    logging::error "Failed to create the namespace ${ns}!"
    return 1
  fi
}

function e2e::test::post_run() {
  local ns="$1"
  local tc="$2"

  shell::run kubectl delete namespace "${ns}"
}

function e2e::test::run() {
  local ns="$1"
  local tc="$2"

  local tc_func="test::run::${tc}"
  # shellcheck disable=SC2155
  local begin_ts=$(date +%s)

  # Run
  logging::info "Running..."
  local result=0
  ${tc_func} || result=$?

  # shellcheck disable=SC2155
  local end_ts=$(date +%s)
  local elapsed=$((end_ts - begin_ts))
  LOGGING_TAGS+=("cost: $(e2e::util::_print_seconds ${elapsed})")

  if ((result == 0)); then
    logging::info "Passed!"
  else
    logging::error "Failed!"
  fi

  return "${result}"
}

function e2e::test::run_next() {
  local idx="$1"
  local ns="e2e-${idx}"
  local tc="$2"

  local LOGGING_TAGS=(e2e "ns/${ns}" "${tc}")

  # Propagate env vars.
  # shellcheck disable=SC2034
  {
    local E2E_NAMESPACE="${ns}"
    local KUBECTL_NAMESPACE="${ns}"
  }

  # Pre-run
  e2e::test::pre_run "${ns}" "${tc}" || return $?

  # Run
  local result=0
  e2e::test::run "${ns}" "${tc}" || result=$?

  # Post-run
  e2e::test::post_run "${ns}" "${tc}" || :

  return "${result}"
}

function e2e::util::_print_seconds() {
  local hour=$(($1 / 3600))
  local minute=$((($1 % 3600) / 60))
  local second=$(($1 % 60))

  if ((hour > 0)); then
    printf "%dh%02dm%02ds\n" "${hour}" "${minute}" "${second}"
  elif ((minute > 0)); then
    printf "%dm%02ds\n" "${minute}" "${second}"
  else
    printf "%ds\n" "${second}"
  fi
}

function e2e::run() {
  local LOGGING_TAGS=(e2e)

  # List test cases.
  local testcases
  IFS=" " read -r -a testcases <<<"$(e2e::list_test_cases)"
  local total_cnt="${#testcases[@]}"

  logging::infof "Running tests, total %d...\n" "${total_cnt}"

  local cur_cnt=0
  local pass_cnt=0
  local fail_cnt=0
  # shellcheck disable=SC2155
  local begin_ts=$(date +%s)

  for tc in "${testcases[@]}"; do
    if e2e::test::run_next "${cur_cnt}" "${tc}"; then
      ((pass_cnt++))
    else
      ((fail_cnt++))
    fi
    ((cur_cnt++))
  done

  # shellcheck disable=SC2155
  local end_ts=$(date +%s)
  local elapsed=$((end_ts - begin_ts))
  logging::info "Total run time: $(e2e::util::_print_seconds ${elapsed})!"

  if ((fail_cnt > 0)); then
    logging::errorf "Test failed! %d/%d failed!\n" "${fail_cnt}" "${total_cnt}"
    return "${fail_cnt}"
  else
    logging::info "Test passed!"
  fi
}

function e2e::pre_run() {
  testenv::setup || return $?

  shell::run docker pull "${E2E_RISINGWAVE_IMAGE}" || return $?
  testenv::k8s::load_docker_image "${E2E_RISINGWAVE_IMAGE}"
}

function e2e::post_run() {
  testenv::teardown
}

function e2e::main() {
  e2e::turn_on_debug_settings_if_debug_is_true

  while getopts ":p:" opt; do
    case "${opt}" in
    p)
      export E2E_TEST_CASE_PREFIX=${OPTARG}
      logging::warn "Run selected test cases with prefix \"${E2E_TEST_CASE_PREFIX}\"..."
      ;;
    *) ;;
    esac
  done

  # Pre-run, exit if fails.
  e2e::pre_run || return $?

  # Run tests.
  local e2e_result=0
  e2e::run || e2e_result=$?

  # Post-run with best effort.
  e2e::post_run || :

  return "${e2e_result}"
}

e2e::main "$@"
