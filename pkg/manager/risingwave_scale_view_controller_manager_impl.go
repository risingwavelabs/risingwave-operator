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
	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit"
	"github.com/risingwavelabs/risingwave-operator/pkg/scaleview"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type risingWaveScaleViewControllerManagerImpl struct {
	client       client.Client
	scaleViewObj *risingwavev1alpha1.RisingWaveScaleView
}

func (mgr *risingWaveScaleViewControllerManagerImpl) HandleScaleViewFinalizer(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	// TODO implement me
	panic("implement me")
}

func (mgr *risingWaveScaleViewControllerManagerImpl) GrabScaleViewLock(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	// TODO implement me
	panic("implement me")
}

func (mgr *risingWaveScaleViewControllerManagerImpl) SyncGroupReplicasToRisingWave(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	scaleViewManager := scaleview.NewComponentGroupReplicasManager(targetObj, mgr.scaleViewObj.Spec.TargetRef.Component)

	hasUpdates := false
	for _, group := range mgr.scaleViewObj.Spec.ScalePolicy {
		updated := scaleViewManager.WriteReplicas(group.Group, group.Replicas)
		if updated {
			hasUpdates = true
		}
	}

	if hasUpdates {
		err := mgr.client.Update(ctx, targetObj)
		if err != nil {
			return ctrlkit.RequeueIfError(err)
		}
	}

	return ctrlkit.Continue()
}

func (mgr *risingWaveScaleViewControllerManagerImpl) SyncGroupReplicasStatusFromRisingWave(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	
}

func NewRisingWaveScaleViewControllerManagerImpl(client client.Client, scaleViewObj *risingwavev1alpha1.RisingWaveScaleView) RisingWaveScaleViewControllerManagerImpl {
	return &risingWaveScaleViewControllerManagerImpl{
		client:       client,
		scaleViewObj: scaleViewObj,
	}
}
