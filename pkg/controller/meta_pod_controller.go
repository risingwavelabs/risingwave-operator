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
		log.Error(err, fmt.Sprintf("Unable to connect to meta pod at %s", addr))
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
	return 0, fmt.Errorf("unable to retrieve the service port from pod")
}

func getPodIP(ctx context.Context, pod *corev1.Pod) (string, error) {
	ip := pod.Status.PodIP
	if ip == "" {
		return "", fmt.Errorf("pod IP is not yet available")
	}
	return ip, nil
}

// Reconcile handles the pods of the meta service. Will add the metaLeaderLabel to the pods.
func (mpc *MetaPodController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	requeueInterval2s := time.Second * time.Duration(2)
	defaultRequeue2sResult := ctrl.Result{RequeueAfter: requeueInterval2s}

	// only reconcile when this is related to a meta pod
	reqPod := &corev1.Pod{}
	mpc.Get(ctx, req.NamespacedName, reqPod)
	rwComponent, ok := reqPod.Labels[consts.LabelRisingWaveComponent]
	if !ok || rwComponent != consts.ComponentMeta {
		return defaultRequeue2sResult, nil
	}

	log := log.FromContext(ctx)

	originalReqPod := reqPod.DeepCopy()
	oldRole, ok := reqPod.Labels[consts.LabelRisingWaveMetaRole]
	if !ok {
		oldRole = consts.MetaRoleUnknown
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	podIP, err := getPodIP(ctx, reqPod)
	if err != nil {
		log.Error(err, "error when reconciling this pod")
		return defaultRequeue2sResult, nil
	}
	podPort, err := getMetaPort(reqPod)
	if err != nil {
		log.Error(err, "error when reconciling this pod")
		return defaultRequeue2sResult, nil
	}

	newRole, err := mpc.metaLeaderStatus(timeoutCtx, podIP, podPort)
	if err != nil {
		return ctrl.Result{Requeue: true}, nil
	}
	reqPod.Labels[consts.LabelRisingWaveMetaRole] = newRole

	// only update if something changed
	if ok && oldRole == newRole {
		return defaultRequeue2sResult, nil
	}

	// Update other meta components only if we have a leadership change
	if !(oldRole == consts.MetaRoleLeader || newRole == consts.MetaRoleLeader) {
		if err := mpc.Patch(ctx, reqPod, client.MergeFrom(originalReqPod)); err != nil {
			log.Error(err, "unable to update Pod")
			return ctrl.Result{Requeue: true}, err
		}
		return defaultRequeue2sResult, nil
	}

	// get other meta pods of this RW instance
	otherMetaPods := &corev1.PodList{}
	labelSet := map[string]string{consts.LabelRisingWaveComponent: consts.ComponentMeta}
	if rwInstance, ok := reqPod.ObjectMeta.Labels[consts.LabelRisingWaveName]; ok {
		labelSet[consts.LabelRisingWaveName] = rwInstance
	} else {
		log.Error(fmt.Errorf("unable to retrieve risingwave name from pod"), "")
	}
	listOptions := client.ListOptions{LabelSelector: labels.SelectorFromSet(labelSet)}
	if err := mpc.Client.List(context.Background(), otherMetaPods, &listOptions); err != nil {
		log.Error(err, "unable to fetch meta pods")
		return ctrl.Result{Requeue: true}, err
	}

	for _, pod := range otherMetaPods.Items {
		podIP, err := getPodIP(ctx, &pod)
		if err != nil {
			log.Error(err, "error when reconciling other meta pod %s", pod.Name)
			return defaultRequeue2sResult, nil
		}
		podPort, err := getMetaPort(&pod)
		if err != nil {
			log.Error(err, fmt.Sprintf("Error getting port from other meta pod %s", pod.Name))
			return defaultRequeue2sResult, err
		}

		// set meta label
		originalPod := pod.DeepCopy()
		oldRole, ok := pod.Labels[consts.LabelRisingWaveMetaRole]
		timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		newRole, err := mpc.metaLeaderStatus(timeoutCtx, podIP, podPort)
		if err != nil {
			return defaultRequeue2sResult, nil
		}
		pod.Labels[consts.LabelRisingWaveMetaRole] = newRole

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

	// update requested pod at the end. We want that early requeueing leads to an update to all meta pods if needed.
	if err := mpc.Patch(ctx, reqPod, client.MergeFrom(originalReqPod)); err != nil {
		log.Error(err, "unable to update Pod")
		return ctrl.Result{Requeue: true}, err
	}

	return defaultRequeue2sResult, nil
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

// Test if pod is pending, then we do not try to connect against it
