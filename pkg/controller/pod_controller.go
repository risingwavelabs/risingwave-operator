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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MetaPodController reconciles a Pod object.
type MetaPodController struct {
	client.Client
	// Scheme *runtime.Scheme
}

// TODO: Rename this file

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

const (
	metaLeaderLabel = "meta-is-leader"
)

var c client.Client

func (r *MetaPodController) metaIsLeader(ip string, port uint) bool {
	// TODO: send heartbeat to pod
	// If success, then is pod leader

	addr := fmt.Sprintf("%s:%v", ip, port)
	log.Log.Info("connecting against %s", addr) // TODO: remove this line

	// TODO: how do I make this connection secure?
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Log.Info("did not connect: %v", err)
		// TODO: retry request
		return false
	}
	defer conn.Close()

	panic("unimplemented")

}

/*
- I do NOT need to use stream, since I just send one heartbeat message

- Generate proto files via
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/meta.proto
*/

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
		port := uint(5690) // What if this changes?

		old_label, ok := pod.Labels[metaLeaderLabel]

		// change meta leader label
		if r.metaIsLeader(podIp, port) {
			pod.Labels[metaLeaderLabel] = "true"
		} else {
			pod.Labels[metaLeaderLabel] = "false"
		}

		// only update if something changed
		if ok && old_label == pod.Labels[metaLeaderLabel] {
			continue
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
