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

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	pb "github.com/risingwavelabs/risingwave-operator/pkg/controller/proto"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"google.golang.org/grpc"
)

// MetaPodController reconciles meta pods object.
type MetaPodController struct {
	client.Client
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

// metaLeaderStatus sends a MetaMember request to a meta node at ip:port, determining if the node is a leader.
func (mpc *MetaPodController) metaLeaderStatus(ctx context.Context, host string, port uint) (string, error) {
	log := log.FromContext(ctx)
	addr := fmt.Sprintf("%s:%v", host, port)

	newCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(newCtx, addr)
	if err != nil {
		log.Error(err, "Unable to connect to meta pod. Retrying...")
		return "", err
	}
	defer conn.Close()
	c := pb.NewMetaMemberServiceClient(conn)

	resp, err := c.Members(ctx, &pb.MembersRequest{})
	if err != nil {
		log.Info(fmt.Sprintf("Sending MembersRequest failed. Assuming meta node is not yet ready. Error was: %s", err.Error()))
		return consts.MetaRoleUnknown, nil
	}

	if len(resp.Members) == 0 {
		return consts.MetaRoleUnknown, nil
	}

	for _, member := range resp.Members {
		if member.IsLeader && host == member.Address.Host && port == uint(member.Address.Port) {
			return consts.MetaRoleLeader, nil
		}
	}
	return consts.MetaRoleFollower, nil
}

// getMetaPort returns the service port of the pod.
func getMetaPort(pod *corev1.Pod) (uint, error) {
	for _, container := range pod.Spec.Containers {
		if container.Name == "meta" {
			for _, containerPort := range container.Ports {
				if containerPort.Name == consts.PortService {
					return uint(containerPort.ContainerPort), nil
				}
			}
		}
	}
	return 0, fmt.Errorf("unable to retrieve the service port from pod %s", pod.ObjectMeta.Name)
}

// Reconcile handles the pods of the meta service. Will add the metaLeaderLabel to the pods.
func (mpc *MetaPodController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	requeueInterval10s := time.Second * time.Duration(10)
	defaultRequeue10sResult := ctrl.Result{RequeueAfter: requeueInterval10s}

	// only reconcile when this is related to a meta pod
	reqPod := &corev1.Pod{}
	mpc.Get(ctx, req.NamespacedName, reqPod)
	rwComponent, ok := reqPod.Labels[consts.LabelRisingWaveComponent]
	if !ok || rwComponent != consts.ComponentMeta {
		return defaultRequeue10sResult, nil
	}

	log := log.FromContext(ctx)

	// get all meta pods
	metaPods := &corev1.PodList{}
	labels := labels.SelectorFromSet(labels.Set{consts.LabelRisingWaveComponent: consts.ComponentMeta})
	listOptions := client.ListOptions{LabelSelector: labels}
	if err := mpc.Client.List(context.Background(), metaPods, &listOptions); err != nil {
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
		port, err := getMetaPort(&pod)
		if err != nil {
			log.Error(err, "Error. Retrying...")
			return defaultRequeue10sResult, err
		}

		// set meta label
		originalPod := pod.DeepCopy()
		oldRole, ok := pod.Labels[consts.LabelRisingWaveMetaRole]
		timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		newRole, err := mpc.metaLeaderStatus(timeoutCtx, podIp, port)
		if err != nil {
			return ctrl.Result{Requeue: true}, nil
		}
		pod.Labels[consts.LabelRisingWaveMetaRole] = newRole
		hasLeader = hasLeader || newRole == consts.MetaRoleLeader
		hasUnknown = hasUnknown || newRole == consts.MetaRoleUnknown

		// only update if something changed
		if ok && oldRole == newRole {
			continue
		}

		// update pod in cluster
		if err := mpc.Patch(ctx, &pod, client.MergeFrom(originalPod)); err != nil {
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

	return defaultRequeue10sResult, nil
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
