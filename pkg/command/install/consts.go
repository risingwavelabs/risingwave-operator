/*
 * Copyright 2022 Singularity Data
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package install

const (
	TimeOut           = 5
	OperatorNamespace = "risingwave-operator-system"
	OperatorName      = "risingwave-operator-controller-manager"

	TemDir = "/tmp/kubectl-rw"

	RisingWaveUrl = "https://github.com/risingwavelabs/risingwave-operator/releases/download/v0.1.1/risingwave-operator.yaml"

	CertManagerUrl = "https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml"
)
