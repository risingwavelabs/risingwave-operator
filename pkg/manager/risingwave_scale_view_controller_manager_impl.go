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

package manager

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

type risingWaveScaleViewControllerManagerImpl struct {
	client client.Client
}

func (mgr *risingWaveScaleViewControllerManagerImpl) GrabScaleViewLock(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	// TODO implement me
	panic("implement me")
}

func (mgr *risingWaveScaleViewControllerManagerImpl) SyncGroupReplicasToRisingWave(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	// TODO implement me
	panic("implement me")
}

func (mgr *risingWaveScaleViewControllerManagerImpl) SyncGroupReplicasStatusFromRisingWave(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	// TODO implement me
	panic("implement me")
}

func NewRisingWaveScaleViewControllerManagerImpl(client client.Client) RisingWaveScaleViewControllerManagerImpl {
	return &risingWaveScaleViewControllerManagerImpl{
		client: client,
	}
}
