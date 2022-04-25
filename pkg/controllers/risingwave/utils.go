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

package risingwave

import (
	"context"
	"fmt"
	v1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/thoas/go-funk"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

type SynType string

func validate(rw *v1alpha1.RisingWave) error {
	if rw.Spec.MetaNode == nil {
		return fmt.Errorf("meta node spec nil")
	}

	if rw.Spec.ObjectStorage == nil {
		return fmt.Errorf("object storage spec nil")
	}

	if rw.Spec.ComputeNode == nil {
		return fmt.Errorf("compute node spec nil")
	}

	if rw.Spec.ComputeNode == nil {
		return fmt.Errorf("frontend spec nil")
	}
	return nil
}

func conditionsToMap(conditions []v1alpha1.RisingWaveCondition) map[v1alpha1.RisingWaveType]v1alpha1.RisingWaveCondition {
	obj := funk.ToMap(conditions, "Type")
	m := obj.(map[v1alpha1.RisingWaveType]v1alpha1.RisingWaveCondition)
	return m
}

func conditionMapToArray(m map[v1alpha1.RisingWaveType]v1alpha1.RisingWaveCondition) []v1alpha1.RisingWaveCondition {
	r := funk.Map(m, func(_ v1alpha1.RisingWaveType, v v1alpha1.RisingWaveCondition) v1alpha1.RisingWaveCondition {
		return v
	})
	return r.([]v1alpha1.RisingWaveCondition)
}

// markRisingWaveRunning will update condition if necessary,
// and update if necessary
func (r *Reconciler) markRisingWaveRunning(ctx context.Context, rw *v1alpha1.RisingWave) (ctrl.Result, error) {
	var needUpdate = false
	m := conditionsToMap(rw.Status.Condition)
	v, e := m[v1alpha1.Running]
	if !e || v.Status == metav1.ConditionFalse {
		m[v1alpha1.Running] = v1alpha1.RisingWaveCondition{
			Type:               v1alpha1.Running,
			LastTransitionTime: metav1.Now(),
			Status:             metav1.ConditionTrue,
			Message:            "Running",
		}
		rw.Status.Condition = conditionMapToArray(m)
		needUpdate = true
	}

	rw.Status.Condition = conditionMapToArray(m)
	if !needUpdate {
		return ctrl.Result{}, nil
	}
	err := r.updateStatus(ctx, rw)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

// markRisingWaveInitializing will mark rw as Initializing and update status
func (r *Reconciler) markRisingWaveInitializing(ctx context.Context, rw *v1alpha1.RisingWave) (ctrl.Result, error) {
	rw.Status.Condition = []v1alpha1.RisingWaveCondition{
		{
			Type:    v1alpha1.Initializing,
			Status:  metav1.ConditionTrue,
			Message: "Initializing",

			LastTransitionTime: metav1.Now(),
		},
	}

	err := r.updateStatus(ctx, rw)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

// updateStatus will update status. If conflict error, will get the latest and retry
func (r *Reconciler) updateStatus(ctx context.Context, rw *v1alpha1.RisingWave) error {
	var newR v1alpha1.RisingWave
	// fetch from cache
	err := r.Get(ctx, types.NamespacedName{Namespace: rw.Namespace, Name: rw.Name}, &newR)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if reflect.DeepEqual(newR.Status, rw.Status) {
		logger.FromContext(ctx).V(1).Info("update status, but no status update, return")
		return nil
	}

	newR.Status = *rw.Status.DeepCopy()
	err = r.Status().Update(ctx, &newR)
	if err != nil {
		return fmt.Errorf("update failed, err %w", err)
	}

	return nil
}
