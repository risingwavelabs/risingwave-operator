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

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	pb "github.com/risingwavelabs/risingwave-operator/pkg/controller/proto"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MetaPodController reconciles meta pods object.
type MetaPodController struct {
	client.Client
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

// getMetaRole sends a MetaMember request to a meta node at ip:port, determining if the node is a leader. Returns true if operation was successful.
func (mpc *MetaPodController) getMetaRole(ctx context.Context, host string, port uint, podName string) (string, bool) {
	log := log.FromContext(ctx)
	addr := fmt.Sprintf("%s:%v", host, port)

	newCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(newCtx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, fmt.Sprintf("Unable to connect to meta pod at %s", addr))
		return "", false
	}
	defer conn.Close()
	c := pb.NewMetaMemberServiceClient(conn)

	resp, err := c.Members(ctx, &pb.MembersRequest{})
	if err != nil {
		log.Info(fmt.Sprintf("Sending MembersRequest failed. Assuming meta node is not yet ready. Error was: %s", err.Error()))
		return consts.MetaRoleUnknown, true
	}

	if len(resp.Members) == 0 {
		return consts.MetaRoleUnknown, true
	}

	// Assuming member.Address.Host like risingwave-meta-default-0.risingwave-meta
	for _, member := range resp.Members {
		memberAddress := ""
		if arr := strings.Split(member.Address.Host, "."); len(arr) == 2 {
			memberAddress = arr[0]
		} else {
			log.Error(fmt.Errorf("unexpected member format %s", member.Address.Host), "")
			return consts.MetaRoleUnknown, false
		}
		if member.IsLeader && podName == memberAddress && port == uint(member.Address.Port) {
			return consts.MetaRoleLeader, true
		}
	}
	return consts.MetaRoleFollower, true
}

// getMetaPort returns the service port of the pod. Returns true if operation was successful.
func getMetaPort(pod *corev1.Pod) (uint, bool) {
	for _, container := range pod.Spec.Containers {
		if container.Name == "meta" {
			for _, containerPort := range container.Ports {
				if containerPort.Name == consts.PortService {
					return uint(containerPort.ContainerPort), true
				}
			}
		}
	}
	return 0, false
}

// getPodIP returns the pod IP of the pod. Returns true if operation was successful.
func getPodIP(pod *corev1.Pod) (string, bool) {
	ip := pod.Status.PodIP
	return ip, ip != ""
}

// getOtherMetPods returns all other meta pods of the same RW instance.
func (mpc *MetaPodController) getOtherMetPods(metaPod *corev1.Pod, ctx context.Context) ([]corev1.Pod, error) {
	log := log.FromContext(ctx)

	// get other meta pods of this RW instance
	otherMetaPods := &corev1.PodList{}
	labelSet := map[string]string{consts.LabelRisingWaveComponent: consts.ComponentMeta}
	if rwInstance, ok := metaPod.ObjectMeta.Labels[consts.LabelRisingWaveName]; ok {
		labelSet[consts.LabelRisingWaveName] = rwInstance
	} else {
		log.Error(fmt.Errorf("unable to retrieve risingwave name from pod"), "")
	}
	listOptions := client.ListOptions{LabelSelector: labels.SelectorFromSet(labelSet)}
	err := mpc.Client.List(context.Background(), otherMetaPods, &listOptions)
	return otherMetaPods.Items, err
}

// Reconcile handles the pods of the meta service. Will add the metaLeaderLabel to the pods.
func (mpc *MetaPodController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	defaultRequeue2sResult := ctrl.Result{RequeueAfter: time.Second * 2}
	pendingRequeue10sResult := ctrl.Result{RequeueAfter: time.Second * 10}

	// only reconcile when this is related to a meta pod
	reqPod := &corev1.Pod{}
	mpc.Get(ctx, req.NamespacedName, reqPod)
	rwComponent, ok := reqPod.Labels[consts.LabelRisingWaveComponent]
	if !ok || rwComponent != consts.ComponentMeta {
		return defaultRequeue2sResult, nil
	}

	log := log.FromContext(ctx)

	originalReqPod := reqPod.DeepCopy()
	oldRole := reqPod.Labels[consts.LabelRisingWaveMetaRole]
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	podIP, ok := getPodIP(reqPod)
	if !ok {
		log.Info("error when reconciling this pod. Assuming that pod is pending")
		return pendingRequeue10sResult, nil
	}
	podPort, ok := getMetaPort(reqPod)
	if !ok {
		log.Error(fmt.Errorf("error when reconciling this pod"), "")
		return defaultRequeue2sResult, nil
	}

	newRole, ok := mpc.getMetaRole(timeoutCtx, podIP, podPort, reqPod.Name)
	if !ok {
		return defaultRequeue2sResult, nil
	}
	reqPod.Labels[consts.LabelRisingWaveMetaRole] = newRole

	// only update if something changed
	if oldRole == newRole {
		return defaultRequeue2sResult, nil
	}

	if err := mpc.Patch(ctx, reqPod, client.MergeFrom(originalReqPod)); err != nil {
		log.Error(err, "unable to update this Pod")
		return defaultRequeue2sResult, nil
	}

	// Update other meta components only if we have a change of leadership
	if oldRole != consts.MetaRoleLeader && newRole != consts.MetaRoleLeader {
		return defaultRequeue2sResult, nil
	}

	otherMetaPods, err := mpc.getOtherMetPods(reqPod, ctx)
	if err != nil {
		log.Error(err, "unable to retrieve other meta pods")
		return defaultRequeue2sResult, nil
	}

	for _, pod := range otherMetaPods {
		podIP, ok := getPodIP(&pod)
		if !ok {
			log.Info("error when reconciling other meta pod %s. Assuming that pod is pending", pod.Name)
			return pendingRequeue10sResult, nil
		}
		podPort, ok := getMetaPort(&pod)
		if !ok {
			log.Error(fmt.Errorf("error getting port from other meta pod %s", pod.Name), "")
			return defaultRequeue2sResult, nil
		}

		// set meta label
		originalPod := pod.DeepCopy()
		oldRole := pod.Labels[consts.LabelRisingWaveMetaRole]
		timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		newRole, ok := mpc.getMetaRole(timeoutCtx, podIP, podPort, pod.Name)
		if !ok {
			return defaultRequeue2sResult, nil
		}
		pod.Labels[consts.LabelRisingWaveMetaRole] = newRole

		// only update if something changed
		if oldRole == newRole {
			continue
		}

		if err := mpc.Patch(ctx, &pod, client.MergeFrom(originalPod)); err != nil {
			log.Error(err, "unable to update Pod")
			continue
		}
		log.Info("updated meta pod", "pod", pod.Name)
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
