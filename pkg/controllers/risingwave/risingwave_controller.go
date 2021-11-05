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

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/manger"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// processFunc need return ready flag and error
// if return ture, means continue to sync, only when Phase==Ready, return true
// if return false, means break this sync and put back into queue and reconcile
type processFunc func(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error)

// Reconciler reconciles a RisingWave object
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
// TODO(user): Modify the Reconcile function to compare the state specified by
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

	// if rw.condition nil, mark as Initializing
	if len(rw.Status.Condition) == 0 {
		rw.Status.Condition = []v1alpha1.RisingWaveCondition{
			{
				Type:    v1alpha1.Initializing,
				Status:  metav1.ConditionTrue,
				Message: "Initializing",

				LastTransitionTime: metav1.Now(),
			},
		}

		err = r.updateStatus(ctx, rw)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
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

	// mark running and update rw status
	if markRunningCondition(rw) {
		err = r.updateStatus(ctx, rw)
		if err != nil {
			log.Error(err, "update risingwave status failed")
			return ctrl.Result{Requeue: true}, err
		}
	}

	return ctrl.Result{}, nil
}

// syncMetaService do meta node create,update,health check
func (r *Reconciler) syncMetaService(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	log := logger.FromContext(ctx)

	if rw.Status.MetaNode.Phase == v1alpha1.ComponentReady {
		return true, nil
	}

	log.Info("Begin sync meta service", "phase", rw.Status.MetaNode.Phase)

	isReady, err := r.syncComponent(ctx, rw, manger.NewMetaMetaNodeManager(), rw.Status.MetaNode.Phase)
	if err != nil {
		return false, err
	}

	if isReady {
		rw.Status.MetaNode.Phase = v1alpha1.ComponentReady
	} else {
		rw.Status.MetaNode.Phase = v1alpha1.ComponentInitializing
	}

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

// syncObjectStorage do object-storage(MinIO,S3,etc...) create,update,health check
func (r *Reconciler) syncObjectStorage(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	var phash = rw.Status.ObjectStorage.Phase
	if phash == v1alpha1.ComponentReady {
		return true, nil
	}

	log := logger.FromContext(ctx)
	log.Info("Begin sync object service", "phase", phash)

	// update storage type
	switch {
	case rw.Spec.ObjectStorage.MinIO != nil:
		rw.Status.ObjectStorage.StorageType = v1alpha1.MinIOType
	case rw.Spec.ObjectStorage.S3:
		rw.Status.ObjectStorage.StorageType = v1alpha1.S3Type
	case rw.Spec.ObjectStorage.Memory:
		rw.Status.ObjectStorage.StorageType = v1alpha1.MemoryType
	default:
		rw.Status.ObjectStorage.StorageType = v1alpha1.UnknownType
	}

	var ready bool
	var err error

	// ensure minIO service
	if rw.Spec.ObjectStorage.MinIO != nil {
		ready, err = r.syncComponent(ctx, rw, manger.NewMinIOManager(), phash)
		if err != nil {
			return false, err
		}
	} else {
		// TODO: support other type
		ready = true
	}

	if ready {
		rw.Status.ObjectStorage.Phase = v1alpha1.ComponentReady
	} else {
		rw.Status.ObjectStorage.Phase = v1alpha1.ComponentInitializing
	}

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

// syncComputeNode do compute-node create,update,health check
func (r *Reconciler) syncComputeNode(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	if rw.Status.ComputeNode.Phase == v1alpha1.ComponentReady {
		return true, nil
	}

	ready, err := r.syncComponent(ctx, rw, manger.NewComputeNodeManager(), rw.Status.ComputeNode.Phase)
	if err != nil {
		return false, err
	}

	if ready {
		rw.Status.ComputeNode.Phase = v1alpha1.ComponentReady
	} else {
		rw.Status.ComputeNode.Phase = v1alpha1.ComponentInitializing
	}

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (r *Reconciler) syncFrontend(ctx context.Context, rw *v1alpha1.RisingWave) (bool, error) {
	if rw.Status.Frontend.Phase == v1alpha1.ComponentReady {
		return true, nil
	}

	ready, err := r.syncComponent(ctx, rw, manger.NewFrontendManager(), rw.Status.Frontend.Phase)
	if err != nil {
		return false, err
	}

	if ready {
		rw.Status.Frontend.Phase = v1alpha1.ComponentReady
	} else {
		rw.Status.Frontend.Phase = v1alpha1.ComponentInitializing
	}

	err = r.updateStatus(ctx, rw)
	if err != nil {
		return false, err
	}

	return false, nil
}

// syncComponent will do creation or update
// return ready flag and error
func (r *Reconciler) syncComponent(
	ctx context.Context,
	rw *v1alpha1.RisingWave,
	m manger.ComponentManager,
	phase v1alpha1.ComponentPhase,
) (bool, error) {
	log := logger.FromContext(ctx).WithValues("component", m.Name())

	log.V(1).Info("Begin to sync component service")
	var start = metav1.Now()

	defer func() {
		dur := metav1.Now().Sub(start.Time).Milliseconds()
		log.V(1).Info("Complete to sync component service", "duration", dur)
	}()

	if len(phase) == 0 {
		log.V(1).Info("Need to create component service")
		err := m.CreateService(ctx, r.Client, rw)
		if err != nil {
			return false, fmt.Errorf("create service failed, %w", err)
		}

		// return not ready, should wait and check
		return false, nil
	}

	// check if changed and update if changed
	changed, err := m.UpdateService(ctx, r.Client, rw)
	if err != nil {
		return false, fmt.Errorf("update service failed, %w", err)
	}

	// if changed, return false and wait service ready
	// TODO: support component upgrade
	if changed {
		log.Info("RisingWave has been changed, need to update and wait ready")
		return false, nil
	}

	log.V(1).Info("Wait component service ready")
	err = m.EnsureService(ctx, r.Client, rw)
	if err != nil {
		return false, fmt.Errorf("enservice service failed, %w", err)
	}

	// return ready
	return true, nil
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
	err = r.deleteComponent(ctx, rw, manger.NewMetaMetaNodeManager())
	if err != nil {
		log.Error(err, "Delete meta failed")
		return
	}
	fSet.Delete(v1alpha1.MetaNodeFinalizer)

	// delete object storage
	if rw.Status.ObjectStorage.StorageType == v1alpha1.MinIOType {
		err = r.deleteComponent(ctx, rw, manger.NewMinIOManager())
		if err != nil {
			log.Error(err, "Delete minIO failed")
			return
		}
	}
	fSet.Delete(v1alpha1.ObjectStorageFinalizer)

	// delete computeNode
	err = r.deleteComponent(ctx, rw, manger.NewComputeNodeManager())
	if err != nil {
		log.Error(err, "Delete compute node failed")
		return
	}
	fSet.Delete(v1alpha1.ComputeNodeFinalizer)

	// delete frontend
	err = r.deleteComponent(ctx, rw, manger.NewFrontendManager())
	if err != nil {
		log.Error(err, "Delete frontend failed")
		return
	}
	fSet.Delete(v1alpha1.FrontendFinalizer)

	return nil
}

func (r *Reconciler) deleteComponent(ctx context.Context, rw *v1alpha1.RisingWave, m manger.ComponentManager) error {
	log := logger.FromContext(ctx).WithValues("component", m.Name())

	log.V(1).Info("Begin to delete risingwave component")
	return m.DeleteService(ctx, r.Client, rw)
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
		logger.FromContext(ctx).V(1).Info("update status, but no stats update, return")
		return nil
	}

	newR.Status = *rw.Status.DeepCopy()
	err = r.Status().Update(ctx, &newR)
	if err != nil {
		return fmt.Errorf("update failed, err %w", err)
	}

	return nil
}
