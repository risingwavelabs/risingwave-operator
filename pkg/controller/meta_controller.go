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

	pb "github.com/risingwavelabs/risingwave-operator/pkg/controller/proto"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// MetaPodController reconciles a Pod object.
type MetaPodController struct {
	client.Client
	// Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

const (
	metaLeaderLabel = "risingwave/meta-role"
)

type labelValue int

const (
	labelValueLeader labelValue = iota
	labelValueFollower
	labelValueUnknown
)

func (l *labelValue) String() string {
	switch *l {
	case labelValueLeader:
		return "leader"
	case labelValueFollower:
		return "follower"
	case labelValueUnknown:
		return "unknown"
	}
	return "UnknownLabelCode"
}

// metaLeaderStatus sends a heartbeat to a meta node at ip:port. If meta responds it is the leader.
func (r *MetaPodController) metaLeaderStatus(ctx context.Context, ip string, port uint) labelValue {
	addr := fmt.Sprintf("%s:%v", ip, port)

	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("Connecting against %s", addr))

	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		// TODO: Do we need to make this connection secure?
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Info("Unable to connect: %s. Retrying...", err.Error())
			continue
		}
		defer conn.Close()
		c := pb.NewHeartbeatServiceClient(conn)

		// TODO: use timeout in this function
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err = c.Heartbeat(ctx, &pb.HeartbeatRequest{})
		if err == nil {
			// TODO: will this ever happen?
			log.Info("no error") // TODO: remove this
			return labelValueLeader
		}

		log.Info(fmt.Sprintf("Err code was %v: %s", status.Code(err), err.Error())) // TODO: remove line
		switch status.Code(err) {
		case codes.OK:
			return labelValueLeader
		case codes.Unimplemented:
			return labelValueFollower
		}
		// TODO: retry on unavailable?
	}
	return labelValueUnknown
}

/*
- Tutorial: https://grpc.io/docs/languages/go/quickstart/
- Generate proto files via

	protoc --go_out=. --go_opt=paths=source_relative \
	    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	    proto/meta.proto
*/

// TODO: rename r into c -> controller
// Reconcile handles the pods of the meta service. Will add the metaLeaderLabel to the pods.
func (r *MetaPodController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// TODO: only reconcile if the pod is actually a meta pod

	log := log.FromContext(ctx)

	// Using a typed object.
	meta_pods := &corev1.PodList{}
	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		log.Error(err, "failed to create client")
		return ctrl.Result{}, err
	}

	// TODO: filter does not work
	labels := labels.SelectorFromSet(labels.Set{"risingwave/component": "meta"})
	listOptions := client.ListOptions{
		LabelSelector: labels,
	}

	if err := c.List(context.Background(), meta_pods, &listOptions); err != nil {
		log.Error(err, "unable to fetch Pod")
		return ctrl.Result{}, err
	}

	if len(meta_pods.Items) == 0 {
		return ctrl.Result{}, nil
	}

	for _, pod := range meta_pods.Items {
		log.Info(fmt.Sprintf("pod is %s/%s", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name))

		podIp := pod.Status.PodIP

		// TODO: am i really talking to the correct port? yes
		port := uint(5690) // What if this changes?

		// change meta leader label
		old_label, ok := pod.Labels[metaLeaderLabel]
		leaderStatus := r.metaLeaderStatus(ctx, podIp, port)
		pod.Labels[metaLeaderLabel] = leaderStatus.String()

		// only update if something changed
		if ok && old_label == leaderStatus.String() {
			log.Info(fmt.Sprintf("skipping update on %s/%s", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name))
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
