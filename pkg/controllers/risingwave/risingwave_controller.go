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
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/hook"
	"github.com/singularity-data/risingwave-operator/pkg/manager"
)

// processFunc need return ready flag and error
// if return ture, means continue to sync, only when Phase==Ready, return true
// if return false, means break this sync and put back into queue and reconcile.
type processFunc func(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error)

// Reconciler reconciles a RisingWave object.
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme

	process []processFunc
}

func NewReconciler(
	c client.Client,
	s *runtime.Scheme) *Reconciler {
	r := &Reconciler{
		Client: c,
		Scheme: s,
	}
	r.process = []processFunc{
		r.syncMetaService,
		r.syncObjectStorage,
		r.syncComputeNode,
		r.syncCompactorNode,
		r.syncFrontend,
	}
	return r
}

// SetupWithManager sets up the controller with the ComponentManager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RisingWave{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=risingwave.singularity-data.com,resources=risingwaves,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=risingwave.singularity-data.com,resources=risingwaves/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=risingwave.singularity-data.com,resources=risingwaves/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// Modify the Reconcile function to compare the state specified by
// the RisingWave object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithValues("ns", req.Namespace, "name", req.Name)

	var obj v1alpha1.RisingWave
	// fetch from cache
	err := r.Get(ctx, req.NamespacedName, &obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{Requeue: false}, nil
		}
		return ctrl.Result{Requeue: false}, err
	}

	rw := obj.DeepCopy()
	if rw == nil {
		return ctrl.Result{}, nil
	}

	// check if deleted
	if rw.DeletionTimestamp != nil {
		// delete resource
		log.V(1).Info("Begin to delete risingwave")
		err := r.doDeletion(ctx, rw)
		if err != nil && !errors.IsNotFound(err) {
			log.Error(err, "delete risingwave failed")
			return ctrl.Result{Requeue: true}, err
		}

		return ctrl.Result{}, nil
	}

	// validate spec should ensure no nil pointer
	err = validate(rw)
	if err != nil {
		log.Error(err, "risingwave validate failed")
		return ctrl.Result{Requeue: false}, err
	}

	// if rw.condition nil, mark rw as Initializing
	if len(rw.Status.Condition) == 0 {
		return r.markRisingWaveInitializing(ctx, rw)
	}

	for _, f := range r.process {
		ready, err := f(ctx, rw)
		if err != nil {
			log.Error(err, "sync risingwave component failed", "process_name", reflect.TypeOf(f).Name())
			return ctrl.Result{Requeue: true}, err
		}
		if !ready {
			return ctrl.Result{}, nil
		}
	}

	// mark rw as running
	return r.markRisingWaveRunning(ctx, rw)
}

// syncMetaService do meta node create,update,health check.
func (r *Reconciler) syncMetaService(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	log := logger.FromContext(ctx)

	var event = hook.GenLifeCycleEvent(rw.Status.MetaNode.Phase, *rw.Spec.MetaNode.Replicas, rw.Status.MetaNode.Replicas)
	if event.Type == hook.SkipType {
		return true, nil
	}
	log.Info("Begin sync meta service", "phase", rw.Status.MetaNode.Phase)

	componentPhase, err := r.syncComponent(ctx, rw, manager.NewMetaMetaNodeManager(), event, hook.LifeCycleOption{
		PostReadyFunc: func() error {
			rw.Status.MetaNode.Replicas = *rw.Spec.MetaNode.Replicas
			return nil
		},
	})
	if err != nil {
		return false, err
	}

	rw.Status.MetaNode.Phase = componentPhase

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

// syncObjectStorage do object-storage(MinIO,S3,etc...) create,update,health check.
func (r *Reconciler) syncObjectStorage(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	var phase = rw.Status.ObjectStorage.Phase

	// if not use minIO. return when phase==Ready
	if phase == v1alpha1.ComponentReady && rw.Spec.ObjectStorage.MinIO == nil {
		return true, nil
	}

	// update storage type
	switch {
	case rw.Spec.ObjectStorage.MinIO != nil:
		rw.Status.ObjectStorage.StorageType = v1alpha1.MinIOType
		// ensure minIO service
		var realPodCount int32 = 0
		if rw.Status.ObjectStorage.MinIOStatus != nil {
			realPodCount = rw.Status.ObjectStorage.MinIOStatus.Replicas
		}
		var event = hook.GenLifeCycleEvent(rw.Status.ObjectStorage.Phase, *rw.Spec.ObjectStorage.MinIO.Replicas, realPodCount)
		if event.Type == hook.SkipType {
			return true, nil
		}

		componentPhase, err := r.syncComponent(ctx, rw, manager.NewMinIOManager(), event, hook.LifeCycleOption{
			PostReadyFunc: func() error {
				if rw.Status.ObjectStorage.MinIOStatus == nil {
					rw.Status.ObjectStorage.MinIOStatus = &v1alpha1.MinIOStatus{}
				}
				rw.Status.ObjectStorage.MinIOStatus.Replicas = *rw.Spec.ObjectStorage.MinIO.Replicas
				return nil
			},
		})
		if err != nil {
			return false, err
		}
		rw.Status.ObjectStorage.Phase = componentPhase
	case rw.Spec.ObjectStorage.S3 != nil:
		rw.Status.ObjectStorage.StorageType = v1alpha1.S3Type
		rw.Status.ObjectStorage.Phase = v1alpha1.ComponentReady
	case rw.Spec.ObjectStorage.Memory:
		rw.Status.ObjectStorage.StorageType = v1alpha1.MemoryType
		rw.Status.ObjectStorage.Phase = v1alpha1.ComponentReady
	default:
		rw.Status.ObjectStorage.StorageType = v1alpha1.UnknownType
	}

	err := r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

// syncComputeNode do compute-node create,update,health check.
func (r *Reconciler) syncComputeNode(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	var event = hook.GenLifeCycleEvent(rw.Status.ComputeNode.Phase, *rw.Spec.ComputeNode.Replicas, rw.Status.ComputeNode.Replicas)
	if event.Type == hook.SkipType {
		return true, nil
	}

	componentPhase, err := r.syncComponent(ctx, rw, manager.NewComputeNodeManager(), event, hook.LifeCycleOption{
		PostReadyFunc: func() error {
			rw.Status.ComputeNode.Replicas = *rw.Spec.ComputeNode.Replicas
			return nil
		},
	})
	if err != nil {
		return false, err
	}
	rw.Status.ComputeNode.Phase = componentPhase

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

// syncCompactorNode do compactor-node create,update,health check.
func (r *Reconciler) syncCompactorNode(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	var event = hook.GenLifeCycleEvent(rw.Status.CompactorNode.Phase, *rw.Spec.CompactorNode.Replicas, rw.Status.CompactorNode.Replicas)
	if event.Type == hook.SkipType {
		return true, nil
	}

	componentPhase, err := r.syncComponent(ctx, rw, manager.NewCompactorNodeManager(), event, hook.LifeCycleOption{
		PostReadyFunc: func() error {
			rw.Status.CompactorNode.Replicas = *rw.Spec.CompactorNode.Replicas
			return nil
		},
	})
	if err != nil {
		return false, err
	}
	rw.Status.CompactorNode.Phase = componentPhase

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (r *Reconciler) syncFrontend(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	var event = hook.GenLifeCycleEvent(rw.Status.Frontend.Phase, *rw.Spec.Frontend.Replicas, rw.Status.Frontend.Replicas)
	if event.Type == hook.SkipType {
		return true, nil
	}

	componentPhase, err := r.syncComponent(ctx, rw, manager.NewFrontendManager(), event, hook.LifeCycleOption{
		PostReadyFunc: func() error {
			rw.Status.Frontend.Replicas = *rw.Spec.Frontend.Replicas
			return nil
		},
	})
	if err != nil {
		return false, err
	}

	rw.Status.Frontend.Phase = componentPhase

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

// syncComponent will do creation or update
// return componentPhase and error.
func (r *Reconciler) syncComponent(
	ctx context.Context,
	rw *v1alpha1.RisingWave,
	mgr manager.ComponentManager,
	event hook.LifeCycleEvent,
	lifeHook hook.LifeCycleOption,
) (v1alpha1.ComponentPhase, error) {
	log := logger.FromContext(ctx).WithValues("component", mgr.Name())

	var start = metav1.Now()
	defer func() {
		dur := metav1.Now().Sub(start.Time).Milliseconds()
		log.V(1).Info("Complete to sync component service", "duration(ms)", dur)
	}()

	switch event.Type {
	case hook.CreateType:
		log.Info("Need to create component service")
		err := mgr.CreateService(ctx, r.Client, rw)
		if err != nil {
			return v1alpha1.ComponentFailed, fmt.Errorf("create service failed, %w", err)
		}

		// return not ready, should wait and check
		return v1alpha1.ComponentInitializing, nil
	case hook.ScaleUpType, hook.ScaleDownType:

		// do post ready lifecycle hook
		if lifeHook.PreUpdateFunc != nil {
			err := lifeHook.PreUpdateFunc()
			if err != nil {
				return v1alpha1.ComponentFailed, fmt.Errorf("pre update hook failed, %w", err)
			}
		}

		// check if changed and update if changed
		changed, err := mgr.UpdateService(ctx, r.Client, rw)
		if err != nil {
			return v1alpha1.ComponentFailed, fmt.Errorf("update service failed, %w", err)
		}

		// if changed, return scaling state and wait service ready
		if changed {
			log.Info("RisingWave has been changed, need to update and wait ready")
			// do post ready lifecycle hook
			if lifeHook.PostUpdateFunc != nil {
				err := lifeHook.PostUpdateFunc()
				if err != nil {
					return v1alpha1.ComponentFailed, fmt.Errorf("post update hook failed, %w", err)
				}
			}
			return v1alpha1.ComponentScaling, nil
		} else {
			return v1alpha1.ComponentReady, nil
		}
	case hook.UpgradeType:
		// TODO: support component upgrade
		log.Info("no support to event type", "type", event.Type)

		return v1alpha1.ComponentUpgrading, nil
	case hook.HealthCheckType:
		log.Info("Wait component service ready")

		err := mgr.EnsureService(ctx, r.Client, rw)
		if err != nil {
			return v1alpha1.ComponentFailed, fmt.Errorf("enservice service failed, %w", err)
		}

		// do post ready lifecycle hook
		if lifeHook.PostReadyFunc != nil {
			err = lifeHook.PostReadyFunc()
			if err != nil {
				return v1alpha1.ComponentFailed, fmt.Errorf("post ready hook failed, %w", err)
			}
		}

		// return ready phase
		return v1alpha1.ComponentReady, nil
	default:
		log.Error(fmt.Errorf("no support to event type"), "type", event.Type)
		return v1alpha1.ComponentUnknown, nil
	}
}

func (r *Reconciler) doDeletion(ctx context.Context, rw *v1alpha1.RisingWave) (err error) {
	log := logger.FromContext(ctx)

	fSet := sets.NewString(rw.Finalizers...)

	// do update
	defer func() {
		// update finalizers means ready to delete from api-server
		rw.Finalizers = fSet.List()

		updateErr := r.Client.Update(ctx, rw)
		switch {
		case errors.IsNotFound(updateErr):
			log.V(1).Info("Update rw, but not existed")
			return
		case updateErr != nil:
			log.Error(updateErr, "Update risingwave finalizers failed")
			err = fmt.Errorf("update risingwave finalizers failed, %w", updateErr)
		default:
			return
		}
	}()

	//TODO: do deletion here

	// delete meta service
	err = r.deleteComponent(ctx, rw, manager.NewMetaMetaNodeManager())
	if err != nil {
		log.Error(err, "Delete meta failed")
		return
	}
	fSet.Delete(v1alpha1.MetaNodeFinalizer)

	// delete object storage
	if rw.Status.ObjectStorage.StorageType == v1alpha1.MinIOType {
		err = r.deleteComponent(ctx, rw, manager.NewMinIOManager())
		if err != nil {
			log.Error(err, "Delete minIO failed")
			return
		}
	}
	fSet.Delete(v1alpha1.ObjectStorageFinalizer)

	// delete computeNode
	err = r.deleteComponent(ctx, rw, manager.NewComputeNodeManager())
	if err != nil {
		log.Error(err, "Delete compute node failed")
		return
	}
	fSet.Delete(v1alpha1.ComputeNodeFinalizer)

	// delete frontend
	err = r.deleteComponent(ctx, rw, manager.NewFrontendManager())
	if err != nil {
		log.Error(err, "Delete frontend failed")
		return
	}
	fSet.Delete(v1alpha1.FrontendFinalizer)

	return nil
}

func (r *Reconciler) deleteComponent(ctx context.Context, rw *v1alpha1.RisingWave, m manager.ComponentManager) error {
	log := logger.FromContext(ctx).WithValues("component", m.Name())

	log.V(1).Info("Begin to delete risingwave component")
	return m.DeleteService(ctx, r.Client, rw)
}
