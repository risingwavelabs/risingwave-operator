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

package webhook

import (
	"context"
	"testing"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/consts"
	"github.com/singularity-data/risingwave-operator/pkg/testutils"
)

func Test_RisingWaveMutatingWebhook_Default(t *testing.T) {
	risingwave := &risingwavev1alpha1.RisingWave{}

	NewRisingWaveMutatingWebhook().Default(context.Background(), risingwave)

	if !testutils.DeepEqual(risingwave, &risingwavev1alpha1.RisingWave{
		Spec: risingwavev1alpha1.RisingWaveSpec{
			Components: risingwavev1alpha1.RisingWaveComponentsSpec{
				Meta: risingwavev1alpha1.RisingWaveComponentMeta{
					Ports: risingwavev1alpha1.RisingWaveComponentMetaPorts{
						RisingWaveComponentCommonPorts: risingwavev1alpha1.RisingWaveComponentCommonPorts{
							ServicePort: consts.DefaultMetaServicePort,
							MetricsPort: consts.DefaultMetaMetricsPort,
						},
						DashboardPort: consts.DefaultMetaDashboardPort,
					},
				},
				Frontend: risingwavev1alpha1.RisingWaveComponentFrontend{
					Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultFrontendServicePort,
						MetricsPort: consts.DefaultFrontendMetricsPort,
					},
				},
				Compute: risingwavev1alpha1.RisingWaveComponentCompute{
					Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultComputeServicePort,
						MetricsPort: consts.DefaultComputeMetricsPort,
					},
				},
				Compactor: risingwavev1alpha1.RisingWaveComponentCompactor{
					Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultCompactorServicePort,
						MetricsPort: consts.DefaultCompactorMetricsPort,
					},
				},
			},
		},
	}) {
		t.Fail()
	}
}
