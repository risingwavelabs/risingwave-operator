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
	"github.com/samber/lo"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit"
	"github.com/risingwavelabs/risingwave-operator/pkg/scaleview"
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

func readRunningReplicas(obj *risingwavev1alpha1.RisingWave, component, group string) int32 {
	pred := func(g risingwavev1alpha1.ComponentGroupReplicasStatus) bool { return g.Name == group }
	switch component {
	case consts.ComponentMeta:
		g, _ := lo.Find(obj.Status.ComponentReplicas.Meta.Groups, pred)
		return g.Running
	case consts.ComponentFrontend:
		g, _ := lo.Find(obj.Status.ComponentReplicas.Frontend.Groups, pred)
		return g.Running
	case consts.ComponentCompactor:
		g, _ := lo.Find(obj.Status.ComponentReplicas.Compactor.Groups, pred)
		return g.Running
	case consts.ComponentCompute:
		g, _ := lo.Find(obj.Status.ComponentReplicas.Compute.Groups, pred)
		return g.Running
	default:
		panic("unexpected")
	}
}

func (mgr *risingWaveScaleViewControllerManagerImpl) SyncGroupReplicasStatusFromRisingWave(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	replicas := int32(0)
	for _, scalePolicy := range mgr.scaleViewObj.Spec.ScalePolicy {
		group := scalePolicy.Group
		runningReplicas := readRunningReplicas(targetObj, mgr.scaleViewObj.Spec.TargetRef.Component, group)
		replicas += runningReplicas
	}
	mgr.scaleViewObj.Status.Replicas = replicas

	return ctrlkit.Continue()
}

func NewRisingWaveScaleViewControllerManagerImpl(client client.Client, scaleViewObj *risingwavev1alpha1.RisingWaveScaleView) RisingWaveScaleViewControllerManagerImpl {
	return &risingWaveScaleViewControllerManagerImpl{
		client:       client,
		scaleViewObj: scaleViewObj,
	}
}
