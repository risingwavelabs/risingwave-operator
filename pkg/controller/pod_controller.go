package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MetaPodController reconciles a Pod object.
type MetaPodController struct {
	client.Client
	// Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

const (
	metaLeaderLabel = "meta-is-leader"
)

var c client.Client

func (r *MetaPodController) metaIsLeader(ip string, port int) bool {
	// TODO: send heartbeat to pod
	// If success, then is pod leader
	panic("unimplemented")
}

// TODO: stateful set here?
// Reconcile handles the pods of the meta service. Will add the leader label to the pod.
func (r *MetaPodController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Using a typed object.
	meta_pods := &corev1.PodList{}
	// TODO: change to meta label here
	if err := c.List(context.Background(), meta_pods, client.MatchingLabels{"someLabelKey": "someLabelValue"}); err != nil {
		log.Error(err, "unable to fetch Pod")
		return ctrl.Result{}, err
	}

	if len(meta_pods.Items) == 0 {
		return ctrl.Result{}, nil
	}

	for _, pod := range meta_pods.Items {
		podIp := pod.Status.PodIP
		port := 5690 // TODO: what if this changes?

		// change meta leader label
		if r.metaIsLeader(podIp, port) {
			pod.Labels[metaLeaderLabel] = "true"
		} else {
			pod.Labels[metaLeaderLabel] = "false"
		}

		// update pod in cluster
		if err := r.Update(ctx, &pod); err != nil {
			if apierrors.IsConflict(err) || apierrors.IsNotFound(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			log.Error(err, "unable to update Pod")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetaPodController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

// NewRisingWaveController creates a new RisingWaveController.
func NewPodController(client client.Client) *MetaPodController {
	return &MetaPodController{
		Client: client,
	}
}
