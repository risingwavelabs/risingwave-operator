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

${__E2E_SOURCE_COMMON_AWS_SH__:=false} && return 0 || __E2E_SOURCE_COMMON_AWS_SH__=true

source "$(dirname "${BASH_SOURCE[0]}")/shell.sh"
source "$(dirname "${BASH_SOURCE[0]}")/logging.sh"

#######################################
# Utility function for using the AWSCLI_PROFILE variable.
# Globals:
#   AWSCLI_PROFILE
# Arguments:
#   Variable sized arguments.
# Returns
#   Code that awscli returns.
#######################################
function awscli::aws() {
  local -a extra_args=()
  [[ -v "AWSCLI_PROFILE" && -n "${AWSCLI_PROFILE}" ]] && extra_args+=("--profile" "${AWSCLI_PROFILE}")
  aws "${extra_args[@]}" "$@"
}

#######################################
# Utility function for getting a profile name by hashing the access key and secret access key.
# Globals:
#   None
# Arguments
#   Access Key
#   Secret Access Key
# Outputs
#   STDOUT
# Returns
#   0 if supported, non-zero otherwise.
#######################################
function awscli::profile::_hash() {
  # shellcheck disable=SC2155
  local os=$(uname -s)
  case "${os}" in
  Linux)
    md5sum <<<"$1$2" | awk '{print $1}'
    ;;
  Darwin)
    md5 <<<"$1$2"
    ;;
  *)
    logging:error "Unsupported platform ${os}"
    return 1
    ;;
  esac
}

#######################################
# Utility function for configuring or adding a profile. If it succeeds, it will set the global variable
# AWSCLI_PROFILE to the current profile.
# Globals:
#   AWSCLI_PROFILE
# Arguments
#   Region
#   Access Key
#   Secret Access Key
# Returns
#   Code that awscli returns.
#######################################
function awscli::configure() {
  (($# == 3)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$2" ]] || { echo >&2 "access key must be provided" && return 1; }
  [[ -n "$3" ]] || { echo >&2 "secret access key must be provided" && return 1; }

  local region=$1
  local access_key=$2
  local secret_access_key=$3
  local profile

  profile=$(awscli::profile::_hash "${access_key}" "${secret_access_key}")

  [[ -n ${region} ]] && aws configure --profile "${profile}" set region "${region}"

  common::run aws configure --profile "${profile}" set aws_access_key_id "${access_key}"
  common::run aws configure --profile "${profile}" set aws_secret_access_key "${secret_access_key}"

  # Set the global var AWSCLI_PROFILE.
  export AWSCLI_PROFILE="${profile}"
}

#######################################
# Utility function for checking if a profile of specified access key/secret access key exists.
# Globals:
#   AWSCLI_PROFILE
# Arguments
#   Access Key
#   Secret Access Key
# Returns
#   0 if true, non-zero otherwise.
#######################################
function awscli::configure::profile_exists() {
  (($# == 2)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$1" ]] || { echo >&2 "access key must be provided" && return 1; }
  [[ -n "$2" ]] || { echo >&2 "secret access key must be provided" && return 1; }

  local access_key=$1
  local secret_access_key=$2
  local profile

  profile=$(awscli::profile::_hash "${access_key}" "${secret_access_key}")

  aws configure list-profiles | grep -w -q "${profile}"
}

#######################################
# Utility function for using a profile of specified access key/secret access key exists.
# Globals:
#   None
# Arguments
#   Access Key
#   Secret Access Key
# Returns
#   0 if exists, non-zero otherwise.
#######################################
function awscli::configure::use_profile() {
  (($# == 2)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$1" ]] || { echo >&2 "access key must be provided" && return 1; }
  [[ -n "$2" ]] || { echo >&2 "secret access key must be provided" && return 1; }

  local access_key=$1
  local secret_access_key=$2
  local profile

  profile=$(awscli::profile::_hash "${access_key}" "${secret_access_key}")

  if aws configure list-profiles | grep -w -q "${profile}"; then
    # Set the global var AWSCLI_PROFILE.
    export AWSCLI_PROFILE="${profile}"
  else
    return 1
  fi
}

#######################################
# Check if a bucket exists in the specified region.
# Globals:
#   AWSCLI_PROFILE
# Arguments:
#   Bucket name.
# Returns
#   0 if the bucket exists, non-zero else or on error.
#######################################
function awscli::s3api::bucket_exists() {
  (($# == 1)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$1" ]] || { echo >&2 "bucket must be provided" && return 1; }

  local bucket=$1

  common::run awscli::aws s3api head-bucket --bucket "${bucket}"
}

#######################################
# Create a bucket in the specified region.
# Globals:
#   AWSCLI_PROFILE
# Arguments:
#   Region.
#   Bucket name.
# Returns
#   0 if the bucket is created, non-zero on error.
#######################################
function awscli::s3api::create_bucket() {
  (($# == 2)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$1" && -n "$2" ]] || { echo >&2 "either region or bucket must be provided" && return 1; }

  local region=$1
  local bucket=$2

  local -a location_constraint_arg=()
  if [[ "${region}" != "us-east-1" ]]; then
    location_constraint_arg=(--create-bucket-configuration "LocationConstraint=${region}")
  fi

  common::run awscli::aws s3api create-bucket --bucket "${bucket}" --region "${region}" "${location_constraint_arg[@]}"
}

#######################################
# Delete a bucket in the specified region.
# Globals:
#   AWSCLI_PROFILE
#   AWSCLI_S3API_FORCE_DELETE_BUCKET
# Arguments:
#   Bucket name.
# Returns
#   0 if the bucket is deleted, non-zero on error.
#######################################
function awscli::s3api::delete_bucket() {
  (($# == 1)) || { echo >&2 "not enough arguments" && return 1; }
  [[ -n "$1" ]] || { echo >&2 "bucket must be provided" && return 1; }

  local bucket=$1

  if [[ "${AWSCLI_S3API_FORCE_DELETE_BUCKET}" == "true" ]]; then
    common::run awscli::aws s3 rb s3://"${bucket}" --force
  else
    common::run awscli::aws s3api delete-bucket --bucket "${bucket}"
  fi
}
