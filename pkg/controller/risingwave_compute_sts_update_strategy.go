// Copyright 2023 RisingWave Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/risingwavelabs/ctrlkit"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/metrics"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

// RisingWaveComputeSTSUpdateStrategy implements an optimized update strategy for the StatefulSets that controls the compute nodes.
type RisingWaveComputeSTSUpdateStrategy struct {
	client client.Client
}

// Reconcile implements reconcile.Reconciler.
func (s *RisingWaveComputeSTSUpdateStrategy) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	// StatefulSet that was updated.
	var sts appsv1.StatefulSet

	err := s.client.Get(ctx, request.NamespacedName, &sts)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(10).Info("Not found, abort")
			return ctrlkit.NoRequeue()
		}
		logger.Error(err, "Failed to get StatefulSet")
		return ctrlkit.RequeueIfErrorAndWrap("unable to get statefulset", err)
	}

	// Controller manager hasn't observed the change. Wait for another second and retry.
	if sts.Generation != sts.Status.ObservedGeneration {
		return ctrlkit.RequeueAfter(time.Second)
	}

	// Return immediately if the StatefulSet isn't updating.
	if sts.Status.UpdatedReplicas == sts.Status.Replicas {
		return ctrlkit.NoRequeue()
	}

	labelSelector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		logger.Error(err, "Failed to get label selector")
		return ctrlkit.RequeueIfErrorAndWrap("unable to get label selector", err)
	}

	// Pods of the StatefulSet.
	var podList corev1.PodList
	err = s.client.List(ctx, &podList, client.InNamespace(request.Namespace), client.MatchingLabelsSelector{Selector: labelSelector})
	if err != nil {
		logger.Error(err, "Failed to list Pods")
		return ctrlkit.RequeueIfErrorAndWrap("unable to list pods", err)
	}

	// A Pod is considered up-to-date if and only if its revision equals to the update revision.
	// Note this could be a false positive when the observed StatefulSet is outdated (has been updated)
	// and the observed Pod is created by the new StatefulSet. AFAIK, there's no way to guarantee it
	// because there's no atomic operation that can be leveraged to get the mapping relations between
	// the generations and the Pods.
	isPodUpToDate := func(pod *corev1.Pod) bool {
		controllerRevisionHash := pod.GetLabels()[appsv1.StatefulSetRevisionLabel]
		return controllerRevisionHash == sts.Status.UpdateRevision
	}

	// Delete pods that is outdated immediately.
	toDeletePods := lo.Filter(podList.Items, func(pod corev1.Pod, _ int) bool {
		return pod.DeletionTimestamp.IsZero() &&
			metav1.GetControllerOfNoCopy(&pod).UID == sts.UID &&
			!isPodUpToDate(&pod)
	})
	if len(toDeletePods) == 0 {
		return ctrlkit.NoRequeue()
	}

	logger.V(1).Info("Deleting pods...", "pods", utils.MapObjectsToNames[corev1.Pod, *corev1.Pod](toDeletePods))
	for _, pod := range toDeletePods {
		err := s.client.Delete(ctx, &pod, client.Preconditions{UID: &pod.UID})
		if client.IgnoreNotFound(err) != nil {
			logger.V(10).Error(err, "Failed to delete pod", "pod", pod.Name)
			return ctrlkit.RequeueIfErrorAndWrap("unable to delete pod", err)
		}
	}

	return ctrlkit.NoRequeue()
}

// SetupWithManager registers itself with the given manager.
func (s *RisingWaveComputeSTSUpdateStrategy) SetupWithManager(mgr ctrl.Manager) error {
	gvk, err := apiutil.GVKForObject(&risingwavev1alpha1.RisingWave{}, s.client.Scheme())
	if err != nil {
		return fmt.Errorf("unable to find gvk for RisingWave: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("RisingWaveComputeSTSUpdateStrategy").
		Watches(
			&appsv1.StatefulSet{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				if object == nil {
					return nil
				}

				// Owned by RisingWave.
				ownerRef := metav1.GetControllerOfNoCopy(object)
				if ownerRef == nil {
					return nil
				}
				if ownerRef.APIVersion != gvk.GroupVersion().String() || ownerRef.Kind != gvk.Kind {
					return nil
				}

				// Has certain labels.
				if object.GetLabels()[consts.LabelRisingWaveName] == "" ||
					object.GetLabels()[consts.LabelRisingWaveComponent] != consts.ComponentCompute {
					return nil
				}

				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Namespace: object.GetNamespace(),
							Name:      object.GetName(),
						},
					},
				}
			}),
			builder.WithPredicates(utils.UpdateEventFilter, predicate.GenerationChangedPredicate{}),
		).
		Complete(metrics.NewControllerMetricsRecorder(s, "RisingWaveComputeSTSUpdateStrategy", lo.Must(apiutil.GVKForObject(&appsv1.StatefulSet{}, s.client.Scheme()))))
}

// NewRisingWaveComputeSTSUpdateStrategy creates a new RisingWaveComputeSTSUpdateStrategy.
func NewRisingWaveComputeSTSUpdateStrategy(client client.Client) *RisingWaveComputeSTSUpdateStrategy {
	return &RisingWaveComputeSTSUpdateStrategy{
		client: client,
	}
}
