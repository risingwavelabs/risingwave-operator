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

${__E2E_SOURCE_TESTS_TESTS_SH__:=false} && return 0 || __E2E_SOURCE_TESTS_TESTS_SH__=true

export E2E_RISINGWAVE_NAME="${E2E_RISINGWAVE_NAME:=e2e}"

if [[ -v "RW_VERSION" ]]; then
  E2E_RISINGWAVE_IMAGE="ghcr.io/risingwavelabs/risingwave:${RW_VERSION}"
fi
export E2E_RISINGWAVE_IMAGE="${E2E_RISINGWAVE_IMAGE:=risingwavelabs/risingwave:v2.6.3}"

source "$(dirname "${BASH_SOURCE[0]}")/risingwave/tests.sh"
