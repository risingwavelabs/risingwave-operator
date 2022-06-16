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

package object

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type RisingWaveManager struct {
	client           client.Client
	risingwave       *risingwavev1alpha1.RisingWave
	risingwaveStatus *risingwavev1alpha1.RisingWaveStatus
}

func (mgr *RisingWaveManager) RisingWave() *risingwavev1alpha1.RisingWave {
	return mgr.risingwave
}

func (mgr *RisingWaveManager) UpdateRisingWaveStatus(ctx context.Context) error {
	// Do nothing if not changed.
	if equality.Semantic.DeepEqual(mgr.risingwaveStatus, &mgr.risingwave.Status) {
		return nil
	}

	return mgr.client.Status().Update(ctx, mgr.risingwave)
}

func NewRisingWaveManager(client client.Client, risingwave *risingwavev1alpha1.RisingWave) *RisingWaveManager {
	return &RisingWaveManager{
		client:           client,
		risingwave:       risingwave,
		risingwaveStatus: risingwave.Status.DeepCopy(),
	}
}
