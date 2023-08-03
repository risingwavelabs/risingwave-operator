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

${__E2E_SOURCE_COMMON_LOGGING_SH__:=false} && return 0 || __E2E_SOURCE_COMMON_LOGGING_SH__=true

source "$(dirname "${BASH_SOURCE[0]}")/shell.sh"

# Bash version check: must be >= 4 to has associated array supported!
shell::assert_minimum_bash_version 4 4

# Associate array for color name and ansi code mappings. (Readonly)
readonly -A _LOGGING_ANSI_COLORS=(
	["red"]="\033[00;31m"
	["green"]="\033[00;32m"
	["yellow"]="\033[00;33m"
	["blue"]="\033[00;34m"
	["magenta"]="\033[00;35m"
	["purple"]="\033[00;36m"
	["light-gray"]="\033[00;37m"
	["light-red"]="\033[01;31m"
	["light-green"]="\033[01;32m"
	["light-yellow"]="\033[01;33m"
	["light-blue"]="\033[01;34m"
	["light-magenta"]="\033[01;35m"
	["light-purple"]="\033[01;36m"
	["white"]="\033[01;37m"
	["escape"]="\033[0m"
)

#######################################
# Utility function for wrapping texts with ansi color codes.
# Globals
#   _LOGGING_ANSI_COLORS
# Arguments
#   Color name, e.g., red
#   Variable length strings
# Outputs
#   STDOUT
# Return
#   0
#######################################
function color::ansi() {
	echo -en "${_LOGGING_ANSI_COLORS[$1]}"
	echo -en "${@:2}"
	echo -en "${_LOGGING_ANSI_COLORS[escape]}"
}

#######################################
# Utility function for wrapping texts with green codes.
# Globals
#   _LOGGING_ANSI_COLORS
# Arguments
#   Variable length strings
# Outputs
#   STDOUT
# Return
#   0
#######################################
function color::ansi::green() {
	color::ansi green "$@"
}

#######################################
# Utility function for wrapping texts with red codes.
# Globals
#   _LOGGING_ANSI_COLORS
# Arguments
#   Variable length strings
# Outputs
#   STDOUT
# Return
#   0
#######################################
function color::ansi::red() {
	color::ansi red "$@"
}

#######################################
# Utility function for wrapping texts with blue codes.
# Globals
#   _LOGGING_ANSI_COLORS
# Arguments
#   Variable length strings
# Outputs
#   STDOUT
# Return
#   0
#######################################
function color::ansi::blue() {
	color::ansi blue "$@"
}

#######################################
# Utility function for wrapping texts with yellow codes.
# Globals
#   _LOGGING_ANSI_COLORS
# Arguments
#   Variable length strings
# Outputs
#   STDOUT
# Return
#   0
#######################################
function color::ansi::yellow() {
	color::ansi yellow "$@"
}

#######################################
# Utility function for wrapping texts with magenta codes.
# Globals
#   _LOGGING_ANSI_COLORS
# Arguments
#   Variable length strings
# Outputs
#   STDOUT
# Return
#   0
#######################################
function color::ansi::magenta() {
	color::ansi magenta "$@"
}

#######################################
# Utility function for wrapping texts with purple codes.
# Globals
#   _LOGGING_ANSI_COLORS
# Arguments
#   Variable length strings
# Outputs
#   STDOUT
# Return
#   0
#######################################
function color::ansi::purple() {
	color::ansi purple "$@"
}

# Constant integer values of log levels. (Readonly)
readonly _LOGGING_LEVEL_DISABLED=0
readonly _LOGGING_LEVEL_ERROR=1
readonly _LOGGING_LEVEL_WARN=2
readonly _LOGGING_LEVEL_INFO=3
readonly _LOGGING_LEVEL_DEBUG=4

# Associate array for log levels. (Readonly)
readonly -A _LOGGING_LEVELS=(
	["off"]=${_LOGGING_LEVEL_DISABLED}
	["error"]=${_LOGGING_LEVEL_ERROR}
	["warn"]=${_LOGGING_LEVEL_WARN}
	["info"]=${_LOGGING_LEVEL_INFO}
	["debug"]=${_LOGGING_LEVEL_DEBUG}
)

# Variables to control the logging behaviors.
_LOGGING_LEVEL=_LOGGING_LEVEL_INFO
_LOGGING_ERROR_TO_STDERR=true

#######################################
# Utility function for getting the log level in integer.
# Globals
#   _LOGGING_LEVEL
# Arguments
#   Log level in either string or index.
# Outputs
#   STDOUT
#   STDERR
# Return
#   0 if the log level is found, non-zero otherwise.
#######################################
function logging::_level_int_val() {
	if [[ $1 =~ ^[0-9]+$ ]]; then
		# Is a number
		echo "$1"
	else
		# Is a string, use ${var,,} to convert the string to lowercase.
		local level=${1,,}
		if [[ -v "_LOGGING_LEVELS[${level}]" ]]; then
			local val=${_LOGGING_LEVELS[${level}]}
			echo "${val}"
		else
			echo >&2 "unknown log level: ${1}"
			return 1
		fi
	fi
}

#######################################
# Utility function for checking if the log level is enabled.
# Globals
#   _LOGGING_LEVEL
# Arguments
#   Target log level in integer.
# Return
#   0 if true, non-zero otherwise.
#######################################
function logging::_is_level_enabled() {
	((_LOGGING_LEVEL >= $1))
}

#######################################
# Utility function for setting the log level.
# Globals
#   _LOGGING_LEVEL
# Arguments
#   Log level in either string or index.
# Return
#   0 if the log level is found and set, non-zero otherwise.
#######################################
function logging::set_level() {
	(($# == 1)) || { echo >&2 "not enough arguments" && return 1; }
	[[ -n $1 ]] || { echo >&2 "level must be provided" && return 1; }

	_LOGGING_LEVEL=$(logging::_level_int_val "$1")
}

#######################################
# Utility function for enabling the logging. By default, the log level is set to INFO.
# Globals
#   _LOGGING_LEVEL
# Arguments
#   None
# Return
#   0
#######################################
function logging::enable() {
	logging::set_level info
}

#######################################
# Utility function for disabling the logging.
# Globals
#   _LOGGING_LEVEL
# Arguments
#   None
# Return
#   0
#######################################
function logging::disable() {
	logging::set_level off
}

#######################################
# Utility function for getting the additional tags.
# Globals
#   LOGGING_TAGS
# Arguments
#   None
# Outputs
#   STDOUT
# Return
#   0
#######################################
function logging::_additional_tags() {
	if [[ -v "LOGGING_TAGS" ]]; then
		# If LOGGING_TAGS isn't an array, convert it to an array.
		if [[ ! "$(declare -p LOGGING_TAGS)" =~ "declare -a" ]]; then
			# shellcheck disable=SC2206
			LOGGING_TAGS=(${LOGGING_TAGS})
		fi

		local quote_tags
		quote_tags=("${LOGGING_TAGS[@]/#/[}")
		echo " ${quote_tags[*]/%/]}"
	fi
}

#######################################
# Utility function for printing messages in log level INFO.
# Globals
#   _LOGGING_LEVEL
# Arguments
#   Variable length strings.
# Outputs
#   STDOUT
# Return
#   0
#######################################
function logging::info() {
	if logging::_is_level_enabled _LOGGING_LEVEL_INFO; then
		echo "$(color::ansi::blue "[INFO]")$(logging::_additional_tags)" "$@"
	fi
}

#######################################
# Utility function for formatting and printing messages in log level INFO.
# Globals
#   _LOGGING_LEVEL
#   LOGGING_TAGS
# Arguments
#   Format string.
#   Variable length strings.
# Outputs
#   STDOUT
# Return
#   0
#######################################
function logging::infof() {
	# shellcheck disable=SC2059
	if logging::_is_level_enabled _LOGGING_LEVEL_INFO; then
		printf "$(color::ansi::blue "[INFO]")$(logging::_additional_tags) $1" "${@:2}"
	fi
}

#######################################
# Utility function for printing messages in log level WARN.
# Globals
#   _LOGGING_LEVEL
#   LOGGING_TAGS
# Arguments
#   Variable length strings.
# Outputs
#   STDOUT
# Return
#   0
#######################################
function logging::warn() {
	if logging::_is_level_enabled _LOGGING_LEVEL_WARN; then
		echo "$(color::ansi::yellow "[WARN]")$(logging::_additional_tags)" "$@"
	fi
}

#######################################
# Utility function for formatting and printing messages in log level WARN.
# Globals
#   _LOGGING_LEVEL
#   LOGGING_TAGS
# Arguments
#   Format string.
#   Variable length strings.
# Outputs
#   STDOUT
# Return
#   0
#######################################
function logging::warnf() {
	# shellcheck disable=SC2059
	if logging::_is_level_enabled _LOGGING_LEVEL_WARN; then
		printf "$(color::ansi::yellow "[WARN]")$(logging::_additional_tags) $1" "${@:2}"
	fi
}

#######################################
# Utility function for printing messages in log level ERROR.
# Globals
#   _LOGGING_LEVEL
#   _LOGGING_ERROR_TO_STDERR, decides if the output is streamed to stderr.
#   LOGGING_TAGS
# Arguments
#   Variable length strings.
# Outputs
#   STDERR if _LOGGING_ERROR_TO_STDERR is true (by default), STDOUT otherwise.
# Return
#   0
#######################################
function logging::error() {
	if logging::_is_level_enabled _LOGGING_LEVEL_ERROR; then
		if [[ ${_LOGGING_ERROR_TO_STDERR} == true ]]; then
			echo >&2 "$(color::ansi::red "[ERRO]")$(logging::_additional_tags)" "$@"
		else
			echo "$(color::ansi::red "[ERRO]")$(logging::_additional_tags)" "$@"
		fi
	fi
}

#######################################
# Utility function for formatting and printing messages in log level ERROR.
# Globals
#   _LOGGING_LEVEL
#   LOGGING_TAGS
# Arguments
#   Format string.
#   _LOGGING_ERROR_TO_STDERR, decides if the output is streamed to stderr.
# Outputs
#   STDERR if _LOGGING_ERROR_TO_STDERR is true (by default), STDOUT otherwise.
# Return
#   0
#######################################
function logging::errorf() {
	# shellcheck disable=SC2059
	if logging::_is_level_enabled _LOGGING_LEVEL_ERROR; then
		if [[ ${_LOGGING_ERROR_TO_STDERR} == true ]]; then
			printf >&2 "$(color::ansi::red "[ERRO]")$(logging::_additional_tags) $1" "${@:2}"
		else
			printf "$(color::ansi::red "[ERRO]")$(logging::_additional_tags) $1" "${@:2}"
		fi
	fi
}

#######################################
# Utility function for printing messages in log level DEBUG. The output is always streamed to stderr.
# Globals
#   _LOGGING_LEVEL
#   LOGGING_TAGS
# Arguments
#   Variable length strings.
# Outputs
#   STDERR
# Return
#   0
#######################################
function logging::debug() {
	if logging::_is_level_enabled _LOGGING_LEVEL_DEBUG; then
		echo >&2 "[DEBU]$(logging::_additional_tags)" "$@"
	fi
}

#######################################
# Utility function for formatting and printing messages in log level DEBUG. The output is always streamed to stderr.
# Globals
#   _LOGGING_LEVEL
#   LOGGING_TAGS
# Arguments
#   Format string.
#   Variable length strings.
# Outputs
#   STDERR
# Return
#   0
#######################################
function logging::debugf() {
	# shellcheck disable=SC2059
	if logging::_is_level_enabled _LOGGING_LEVEL_DEBUG; then
		printf >&2 "[DEBU]$(logging::_additional_tags) $1" "${@:2}"
	fi
}
