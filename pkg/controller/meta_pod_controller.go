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
	"strings"
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
	"google.golang.org/grpc/credentials/insecure"
)

// Tutorial: https://grpc.io/docs/languages/go/quickstart/
// Generate proto files via make proto

// MetaPodController reconciles meta pods object.
type MetaPodController struct {
	client.Client
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

// metaLeaderStatus sends a MetaMember request to a meta node at ip:port, determining if the node is a leader.
func (mpc *MetaPodController) metaLeaderStatus(ctx context.Context, host string, port uint) labelValue {
	log := log.FromContext(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%v", host, port)

	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Error(err, "Unable to connect to meta pod. Retrying...")
			continue
		}
		defer conn.Close()
		c := pb.NewMetaMemberServiceClient(conn)

		resp, err := c.Members(ctx, &pb.MembersRequest{})
		if err != nil {
			log.Error(err, "Sending MembersRequest failed")
			continue
		}

		for _, member := range resp.Members {
			if member.IsLeader && host == member.Address.Host && port == uint(member.Address.Port) {
				return labelValueLeader
			}
		}
		return labelValueFollower
	}
	return labelValueUnknown
}

// Reconcile handles the pods of the meta service. Will add the metaLeaderLabel to the pods.
func (mpc *MetaPodController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	requeueInterval60s := time.Second * time.Duration(60)
	defaultRequeueResult := ctrl.Result{RequeueAfter: requeueInterval60s}

	// only reconcile when this is related to a meta pod
	if !strings.Contains(req.Name, "meta") {
		return defaultRequeueResult, nil
	}

	log := log.FromContext(ctx)
	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		log.Error(err, "failed to create client")
		return ctrl.Result{Requeue: true}, err
	}

	// get all meta pods
	metaPods := &corev1.PodList{}
	labels := labels.SelectorFromSet(labels.Set{"risingwave/component": "meta"})
	listOptions := client.ListOptions{LabelSelector: labels}
	if err := c.List(context.Background(), metaPods, &listOptions); err != nil {
		log.Error(err, "unable to fetch meta pods")
		return ctrl.Result{Requeue: true}, err
	}

	// Do not requeue, since we do not have any meta pods
	if len(metaPods.Items) == 0 {
		return ctrl.Result{}, nil
	}

	hasUnknown := false
	hasLeader := false
	for _, pod := range metaPods.Items {
		podIp := pod.Status.PodIP
		// FIXME: Do not hardcode port here. Pass in as --arg. Follow-up PR

		port := uint(5690)

		// set meta label
		old_label, ok := pod.Labels[metaLeaderLabel]
		leaderStatus := mpc.metaLeaderStatus(ctx, podIp, port)
		pod.Labels[metaLeaderLabel] = leaderStatus.String()
		hasLeader = hasLeader || leaderStatus == labelValueLeader
		hasUnknown = hasUnknown || leaderStatus == labelValueUnknown

		// only update if something changed
		if ok && old_label == leaderStatus.String() {
			continue
		}

		// update pod in cluster
		if err := mpc.Update(ctx, &pod); err != nil {
			if apierrors.IsConflict(err) || apierrors.IsNotFound(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			log.Error(err, "unable to update Pod")
			return ctrl.Result{Requeue: true}, err
		}
	}

	// requeue if there currently is no leader or meta nodes are in unknown status
	if !hasLeader || hasUnknown {
		return ctrl.Result{Requeue: true}, nil
	}

	return defaultRequeueResult, nil
}

// SetupWithManager sets up the controller with the Manager.
func (mpc *MetaPodController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(mpc)
}

// NewRisingWaveController creates a new RisingWaveController.
func NewMetaPodController(client client.Client) *MetaPodController {
	return &MetaPodController{
		Client: client,
	}
}
