/*
 * Copyright 2023 RisingWave Labs
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
	"time"

	"github.com/go-logr/logr"
	"github.com/risingwavelabs/ctrlkit"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/scaleview"
)

type risingWaveScaleViewControllerManagerImpl struct {
	client              client.Client
	scaleView           *risingwavev1alpha1.RisingWaveScaleView
	scaleViewStatusCopy *risingwavev1alpha1.RisingWaveScaleViewStatus
}

func (mgr *risingWaveScaleViewControllerManagerImpl) isStatusChanged() bool {
	return !equality.Semantic.DeepEqual(&mgr.scaleView.Status, mgr.scaleViewStatusCopy)
}

func (mgr *risingWaveScaleViewControllerManagerImpl) isTargetObjMatched(targetObj *risingwavev1alpha1.RisingWave) bool {
	return targetObj != nil && targetObj.UID == mgr.scaleView.Spec.TargetRef.UID
}

// UpdateScaleViewStatus implements theRisingWaveScaleViewControllerManagerImpl.
func (mgr *risingWaveScaleViewControllerManagerImpl) UpdateScaleViewStatus(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
	if mgr.isStatusChanged() {
		err := mgr.client.Status().Update(ctx, mgr.scaleView)
		return ctrlkit.RequeueIfErrorAndWrap("unable to update status of risingwavescaleview", err)
	}
	return ctrlkit.Continue()
}

// GrabOrUpdateScaleViewLock implements RisingWaveScaleViewControllerManagerImpl.
func (mgr *risingWaveScaleViewControllerManagerImpl) GrabOrUpdateScaleViewLock(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	if !mgr.isTargetObjMatched(targetObj) {
		if targetObj != nil {
			logger.Info("Object's uid doesn't match", "expect", mgr.scaleView.Spec.TargetRef.UID, "actual", targetObj.UID)
		}

		return ctrlkit.Continue()
	}

	lockMgr := object.NewScaleViewLockManager(targetObj)

	// Update the lock in memory.
	updated, err := lockMgr.GrabOrUpdateScaleViewLockFor(mgr.scaleView)
	if err != nil {
		return ctrlkit.RequeueIfErrorAndWrap("unable to grab or update lock", err)
	}

	// If updated, try updating the remote. Here the update is an atomic operation which means it either succeeds updating
	// the current captured object or fails with a conflict error.
	if updated {
		logger.Info("Lock grabbed(updated) in memory! Try updating the remote...")
		if err := mgr.client.Status().Update(ctx, targetObj); err != nil {
			return ctrlkit.RequeueIfErrorAndWrap("unable to update the status of RisingWave", err)
		}
		mgr.scaleView.Status.Locked = true
		return ctrlkit.RequeueImmediately()
	}
	return ctrlkit.Continue()
}

// SyncGroupReplicasToRisingWave implementsRisingWaveScaleViewControllerManagerImpl.
func (mgr *risingWaveScaleViewControllerManagerImpl) SyncGroupReplicasToRisingWave(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	if !mgr.isTargetObjMatched(targetObj) {
		if targetObj != nil {
			logger.Info("Object's uid doesn't match", "expect", mgr.scaleView.Spec.TargetRef.UID, "actual", targetObj.UID)
		}

		return ctrlkit.Continue()
	}

	targetObjOriginal := targetObj.DeepCopy()

	lockObj := object.NewScaleViewLockManager(targetObj).GetScaleViewLock(mgr.scaleView)
	if lockObj == nil || lockObj.Generation != mgr.scaleView.Generation {
		logger.Info("Lock is outdated, retry...")
		return ctrlkit.RequeueAfter(5 * time.Millisecond)
	}

	helper := scaleview.NewRisingWaveScaleViewHelper(targetObj, mgr.scaleView.Spec.TargetRef.Component)

	changed := false
	for _, group := range lockObj.GroupLocks {
		updated := helper.WriteReplicas(group.Name, group.Replicas)
		changed = changed || updated
	}

	if changed {
		logger.Info("Syncing the replicas changes...")

		err := mgr.client.Patch(ctx, targetObj, client.MergeFrom(targetObjOriginal))
		if err != nil {
			return ctrlkit.RequeueIfErrorAndWrap("unable to update the replicas of RisingWave", err)
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
	case consts.ComponentStandalone:
		panic("not supported")
	default:
		panic("unexpected")
	}
}

// SyncGroupReplicasStatusFromRisingWave implementsRisingWaveScaleViewControllerManagerImpl.
func (mgr *risingWaveScaleViewControllerManagerImpl) SyncGroupReplicasStatusFromRisingWave(ctx context.Context, logger logr.Logger, targetObj *risingwavev1alpha1.RisingWave) (ctrl.Result, error) {
	if !mgr.isTargetObjMatched(targetObj) {
		mgr.scaleView.Status.Replicas = ptr.To(int32(0))
		mgr.scaleView.Status.Locked = false
		return ctrlkit.Continue()
	}
	replicas := int32(0)
	for _, scalePolicy := range mgr.scaleView.Spec.ScalePolicy {
		group := scalePolicy.Group
		runningReplicas := readRunningReplicas(targetObj, mgr.scaleView.Spec.TargetRef.Component, group)
		replicas += runningReplicas
	}
	mgr.scaleView.Status.Replicas = ptr.To(replicas)
	return ctrlkit.Continue()
}

// NewRisingWaveScaleViewControllerManagerImpl creates an object that implements the RisingWaveScaleViewControllerManagerImpl.
func NewRisingWaveScaleViewControllerManagerImpl(client client.Client, scaleView *risingwavev1alpha1.RisingWaveScaleView) RisingWaveScaleViewControllerManagerImpl {
	return &risingWaveScaleViewControllerManagerImpl{
		client:              client,
		scaleView:           scaleView,
		scaleViewStatusCopy: scaleView.Status.DeepCopy(),
	}
}
