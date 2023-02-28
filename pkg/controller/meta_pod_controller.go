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

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		if podName == memberAddress && port == uint(member.Address.Port) {
			if member.IsLeader {
				return consts.MetaRoleLeader, true
			}
			return consts.MetaRoleFollower, true
		}
	}
	return consts.MetaRoleUnknown, true
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

// getLeaderMetaPods returns all other meta pods of the same RW instance. Returns true if operation was successful.
func (mpc *MetaPodController) getLeaderMetaPods(metaPod *corev1.Pod, ctx context.Context) ([]corev1.Pod, bool) {
	log := log.FromContext(ctx)

	rwInstance, ok := metaPod.ObjectMeta.Labels[consts.LabelRisingWaveName]
	if !ok {
		log.Error(fmt.Errorf("unable to retrieve risingwave name from pod"), "")
		return []corev1.Pod{}, false
	}

	// get pods that are marked as leader from this instance
	otherLeaderPods := &corev1.PodList{}
	labelSet := map[string]string{
		consts.LabelRisingWaveComponent: consts.ComponentMeta,
		consts.LabelRisingWaveMetaRole:  consts.MetaRoleLeader,
		consts.LabelRisingWaveName:      rwInstance,
	}
	listOptions := client.ListOptions{LabelSelector: labels.SelectorFromSet(labelSet)}
	err := mpc.Client.List(context.Background(), otherLeaderPods, &listOptions)
	leaders := otherLeaderPods.Items
	if len(leaders) > 1 {
		log.Error(fmt.Errorf("multiple pods are marked as leaders: %v", leaders), "")
	}
	return leaders, err != nil
}

// Reconcile handles the pods of the meta service. Will add the metaLeaderLabel to the pods.
func (mpc *MetaPodController) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, e error) {
	defaultRequeue2sResult := ctrl.Result{RequeueAfter: time.Second * 2}

	// only reconcile when this is related to a meta pod
	reqPod := &corev1.Pod{}
	mpc.Get(ctx, req.NamespacedName, reqPod)
	if reqPod.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
		return ctrl.Result{}, nil
	}

	log := log.FromContext(ctx)

	originalReqPod := reqPod.DeepCopy()
	oldRole := reqPod.Labels[consts.LabelRisingWaveMetaRole]
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	podIP, okPodIP := getPodIP(reqPod)
	podPort, okMetaPort := getMetaPort(reqPod)
	if !okPodIP || !okMetaPort {
		log.Info("Ignoring this pod")
		return ctrl.Result{}, nil
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

	// We need to defer the update, since we are updating all leader pods below
	defer func() {
		if err := mpc.Patch(ctx, reqPod, client.MergeFrom(originalReqPod)); err != nil {
			log.Error(err, "unable to update this Pod")
			res = defaultRequeue2sResult
			e = nil
		}
	}()

	// Update other meta components only if we have a change of leadership
	if newRole == consts.MetaRoleLeader {
		return defaultRequeue2sResult, nil
	}

	otherMetaPods, ok := mpc.getLeaderMetaPods(reqPod, ctx)
	if !ok {
		return defaultRequeue2sResult, nil
	}

	for _, pod := range otherMetaPods {
		podIP, okPodIP := getPodIP(&pod)
		podPort, okMetaPort := getMetaPort(&pod)
		if !okPodIP || !okMetaPort {
			log.Info("Ignoring pod %s", pod.Name)
			return ctrl.Result{}, nil
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
