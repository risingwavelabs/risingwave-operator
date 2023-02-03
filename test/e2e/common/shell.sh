# Copyright 2023 RisingWave Labs
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

${__E2E_SOURCE_COMMON_SHELL_SH__:=false} && return 0 || __E2E_SOURCE_COMMON_SHELL_SH__=true

#######################################
# Helper function to assert the minimum bash version.
# Globals
#   None
# Arguments
#   Major version (inclusive).
#   Minor version (inclusive, optional).
#   Patch version (inclusive, optional).
# Outputs
#   STDERR, when assertion fails.
# Returns
#   0 if bash version meets requirements, non-zero if not.
#######################################
function shell::assert_minimum_bash_version() {
  (($# >= 1)) || { echo >&2 "not enough arguments" && return 1; }

  local major=$1
  local minor=0
  local patch=0

  (($# < 2)) || minor=$2
  (($# < 3)) || patch=$3

  if ((BASH_VERSINFO[0] < major)) ||
    ((BASH_VERSINFO[0] == major && BASH_VERSINFO[1] < minor)) ||
    ((BASH_VERSINFO[0] == major && BASH_VERSINFO[2] == minor || BASH_VERSINFO[2] < patch)); then
    echo >&2 "The minimum bash version required is ${major}.${minor}.${patch}, but current is ${BASH_VERSION}!"
    return 1
  fi
}

#######################################
# Utility function for checking if command exists.
# Globals
#   PATH, optional
# Arguments
#   Command name
# Returns
#   0 if exists, non-zero if not.
#######################################
function shell::command_exists() {
  (($# == 1)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$1" ]] || { echo >&2 "command name must be provided" && return 1; }

  command -v "$1" >/dev/null 2>&1
}

#######################################
# Utility function for running commands with a debug log and verbose control.
# Globals
#   TRACE_COMMAND
#   SHOW_COMMAND_OUTPUT
# Arguments
#   Command to run and arguments.
# Returns
#   Code returns from the command.
#######################################
function shell::run() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(command)

  if [[ "${TRACE_COMMAND:=false}" == "true" ]]; then
    logging::debug "$*"
  fi

  local exit_code=0

  {
    if [[ ${SHOW_COMMAND_OUTPUT:=false} == "true" ]]; then
      # shellcheck disable=SC2294
      eval "$@"
    else
      # shellcheck disable=SC2294
      eval "$@" >/dev/null 2>&1
    fi
  } || exit_code=$?

  if [[ "${TRACE_COMMAND:=false}" == "true" ]]; then
    logging::debug "$*, exit code: ${exit_code}"
  fi

  return ${exit_code}
}

# Global variables used for capturing stdout and stderr of command run by `shell::run_and_capture_outputs`.
CAPTURED_STDOUT=""
CAPTURED_STDERR=""
CAPTURED_EXIT_CODE=0

#######################################
# Utility function for running commands and capture its stdout and stderr separately.
# The solution comes from the following answer on the stackoverflow:
# https://stackoverflow.com/questions/11027679/capture-stdout-and-stderr-into-different-variables
# Note
#   To use this function concurrently in jobs, please make sure to wrap it in a subshell so that
#   the variables won't conflict with each other.
#
#   We recommend you to define local variables inside the invoker function avoid conflicts.
# Globals
#   CAPTURED_STDOUT
#   CAPTURED_STDERR
#   CAPTURED_EXIT_CODE
# Arguments
#   Command to run and arguments.
# Returns
#   Code returns from the command.
#######################################
function shell::run_and_capture_outputs() {
  # shellcheck disable=SC2034
  {
    IFS=$'\n' read -r -d '' CAPTURED_STDERR
    IFS=$'\n' read -r -d '' CAPTURED_STDOUT
    IFS=$'\n' read -r -d '' CAPTURED_EXIT_CODE
  } < <((printf '\0%s\0%d\0' "$("$@")" "${?}" 1>&2) 2>&1)

  return "${CAPTURED_EXIT_CODE}"
}

#######################################
# Run the command in the background.
# Globals
#   BACKGROUND_PIDS
# Arguments
#   Variable sized strings
# Outputs
#   Depends on the command that executes
# Returns
#   0
#######################################
function shell::spawn() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(background command)

  if [[ "${TRACE_COMMAND:=false}" == "true" ]]; then
    logging::debug "$*"
  fi

  # shellcheck disable=SC2294
  eval "$@" &

  local pid=$!
  # shellcheck disable=SC2034
  BACKGROUND_PIDS["${pid}"]="$*"

  if [[ "${TRACE_COMMAND:=false}" == "true" ]]; then
    logging::debug "$*, pid: ${pid}"
  fi
}

#######################################
# Wait for the background jobs to complete one by one. If any of the background jobs returns non-zero code,
# the function breaks and returns with that code.
# Globals
#   BACKGROUND_PIDS
# Arguments
#   None
# Outputs
#   None
# Returns
#   0 if all succeeds, the first non-zero exit code otherwise.
#######################################
function shell::wait() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(background command)

  local cmd
  local exit_code=0

  for pid in "${!BACKGROUND_PIDS[@]}"; do
    cmd=${BACKGROUND_PIDS[${pid}]}
    exit_code=0
    wait "${pid}" || exit_code=$?

    if [[ "${TRACE_COMMAND:=false}" == "true" ]]; then
      logging::debug "${cmd[*]}, pid: ${pid}, exit code: ${exit_code}"
    fi

    ((exit_code == 0)) || return "${exit_code}"
  done
}

#######################################
# Wait for all the background jobs to complete. It returns the last non-zero code it met or 0.
# Globals
#   BACKGROUND_PIDS
# Arguments
#   None
# Outputs
#   None
# Returns
#   0 if all succeeds, the last non-zero exit code otherwise.
#######################################
function shell::wait_all() {
  # shellcheck disable=SC2034
  local LOGGING_TAGS=(background command)

  local cmd
  local return_code=0
  local exit_code=0

  for pid in "${!BACKGROUND_PIDS[@]}"; do
    cmd=${BACKGROUND_PIDS[${pid}]}

    exit_code=0
    wait "${pid}" || exit_code=$?
    ((exit_code == 0)) || return_code=${exit_code}

    if [[ "${TRACE_COMMAND:=false}" == "true" ]]; then
      logging::debug "${cmd[*]}, pid: ${pid}, exit code: ${exit_code}"
    fi
  done

  return "${return_code}"
}

#######################################
# Calculate the MD5 hash of the given content.
# Globals
#   None
# Arguments
#   The value to hash.
# Outputs
#   MD5 hash value.
# Returns
#   0 if succeeds, non-zero on errors.
#######################################
function shell::md5() {
  # shellcheck disable=SC2155
  local os=$(uname -s)
  case "${os}" in
  Linux)
    md5sum <<<"$1" | awk '{print $1}'
    ;;
  Darwin)
    md5 <<<"$1"
    ;;
  *)
    logging:error "Unsupported platform ${os}"
    return 1
    ;;
  esac
}
