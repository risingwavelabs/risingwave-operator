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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/risingwavelabs/ctrlkit"

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	pb "github.com/risingwavelabs/risingwave-operator/pkg/controller/proto"
	"github.com/risingwavelabs/risingwave-operator/pkg/factory/envs"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MetaPodRoleLabeler reconciles meta pods object.
type MetaPodRoleLabeler struct {
	client.Client
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch

// getMetaRole sends a gRPC request to the meta node at host:port to tell its role from the response. The endpoint is used
// to identify the meta node. If the node isn't found in the response, an unknown will be returned.
func (mpl *MetaPodRoleLabeler) getMetaRole(ctx context.Context, host string, port uint, endpoint string) (string, error) {
	addr := fmt.Sprintf("%s:%v", host, port)
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", fmt.Errorf("unable to connect: %w", err)
	}
	defer conn.Close()

	metaClient := pb.NewMetaMemberServiceClient(conn)

	resp, err := metaClient.Members(ctx, &pb.MembersRequest{})
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	for _, member := range resp.Members {
		if member.Address.Host == endpoint && member.Address.Port == int32(port) {
			return lo.If(member.IsLeader, consts.MetaRoleLeader).Else(consts.MetaRoleFollower), nil
		}
	}

	logger := log.FromContext(ctx)
	logger.Info("No role recognized from the current member list!", "members", resp.Members, "address", fmt.Sprintf("%s:%d", host, port), "endpoint", endpoint)

	return consts.MetaRoleUnknown, nil
}

func (mpl *MetaPodRoleLabeler) isRisingWaveMetaPod(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}

	if _, ok := pod.Labels[consts.LabelRisingWaveName]; !ok {
		return false
	}
	if pod.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
		return false
	}

	if utils.GetContainerFromPod(pod, "meta") == nil {
		return false
	}

	return true
}

func (mpl *MetaPodRoleLabeler) getEndpointFromArgs(pod *corev1.Pod, args []string) string {
	endpoint := ""

	// Get the subsequent value of "--host" or "--advertise-addr".
	for i := range args {
		if i == len(args)-1 {
			break
		}

		arg := args[i]
		switch {
		case arg == "--host":
			endpoint = args[i+1]
		case arg == "--advertise-addr":
			endpoint = strings.Split(args[i+1], ":")[0]
		case strings.HasPrefix(arg, "--host="):
			endpoint = arg[len("--host="):]
		case strings.HasPrefix(arg, "--advertise-addr="):
			endpoint = strings.Split(arg[len("--advertise-addr="):], ":")[0]
		}

		if len(endpoint) > 0 {
			break
		}
	}

	if len(endpoint) == 0 {
		return ""
	}

	endpoint = strings.ReplaceAll(endpoint, "$(POD_IP)", pod.Status.PodIP)
	endpoint = strings.ReplaceAll(endpoint, "$(POD_NAME)", pod.Name)
	endpoint = strings.ReplaceAll(endpoint, "$(POD_NAMESPACE)", pod.Namespace)

	return endpoint
}

func (mpl *MetaPodRoleLabeler) getEndpointFromEnvVars(pod *corev1.Pod, envVars []corev1.EnvVar) string {
	endpoint := ""

	// Get the value of RW_ADVERTISE_ADDR.
	for _, envVar := range envVars {
		if envVar.Name == envs.RWAdvertiseAddr {
			endpoint = strings.Split(envVar.Value, ":")[0]
			break
		}
	}

	if len(endpoint) == 0 {
		return ""
	}

	endpoint = strings.ReplaceAll(endpoint, "$(POD_IP)", pod.Status.PodIP)
	endpoint = strings.ReplaceAll(endpoint, "$(POD_NAME)", pod.Name)
	endpoint = strings.ReplaceAll(endpoint, "$(POD_NAMESPACE)", pod.Namespace)

	return endpoint
}

func (mpl *MetaPodRoleLabeler) syncRoleLabelForSinglePod(ctx context.Context, pod *corev1.Pod) (string, error) {
	// Extract information from the Pod.
	if !mpl.isRisingWaveMetaPod(pod) {
		return "", errors.New("not a meta pod")
	}
	if pod.Status.PodIP == "" {
		return "", errors.New("not running")
	}
	metaContainer := utils.GetContainerFromPod(pod, "meta")
	if metaContainer == nil {
		return "", errors.New("meta container not found")
	}
	svcPort, ok := utils.GetPortFromContainer(metaContainer, consts.PortService)
	if !ok {
		return "", errors.New("service port not found")
	}
	endpoint := mpl.getEndpointFromArgs(pod, metaContainer.Args)
	if len(endpoint) == 0 {
		if endpoint = mpl.getEndpointFromEnvVars(pod, metaContainer.Env); len(endpoint) == 0 {
			return "", errors.New("endpoint not found")
		}
	}

	logger := log.FromContext(ctx).WithValues("pod", pod.Name)

	// Send a gRPC request and get the current role.
	role, err := func() (string, error) {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return mpl.getMetaRole(ctx, pod.Status.PodIP, uint(svcPort), endpoint)
	}()
	if err != nil {
		logger.Info("Failed to get the current role from the meta Pod.", "error", err)
		// Use an unknown role.
		role = consts.MetaRoleUnknown
	}

	// Update the role if it changes.
	beforeRole := pod.Labels[consts.LabelRisingWaveMetaRole]
	if beforeRole != role {
		originalPod := pod.DeepCopy()
		pod.Labels[consts.LabelRisingWaveMetaRole] = role
		err := mpl.Patch(ctx, pod, client.StrategicMergeFrom(originalPod))
		return role, err
	}

	return role, nil
}

func (mpl *MetaPodRoleLabeler) syncRoleLabels(ctx context.Context, pod *corev1.Pod) {
	logger := log.FromContext(ctx)

	beforeRole := pod.Labels[consts.LabelRisingWaveMetaRole]
	role, err := mpl.syncRoleLabelForSinglePod(ctx, pod)
	if err != nil {
		logger.Info("Failed to sync the meta role label.", "pod", pod.Name, "error", err)
		return
	}

	if beforeRole != consts.MetaRoleLeader && role == consts.MetaRoleLeader {
		risingwaveName := pod.Labels[consts.LabelRisingWaveName]
		currentPod := pod.Name

		var leaderPodList corev1.PodList
		err := mpl.List(ctx, &leaderPodList, client.InNamespace(pod.Namespace), client.MatchingLabels{
			consts.LabelRisingWaveName:      risingwaveName,
			consts.LabelRisingWaveComponent: consts.ComponentMeta,
			consts.LabelRisingWaveMetaRole:  consts.MetaRoleLeader,
		})
		if err != nil {
			logger.Info("Failed to list leader Pods.", "pod", pod.Name, "error", err)
		}

		leaderPods := lo.Filter(leaderPodList.Items, func(item corev1.Pod, _ int) bool {
			return item.Name != currentPod
		})

		// Do our best.
		for _, pod := range leaderPods {
			if _, err := mpl.syncRoleLabelForSinglePod(ctx, &pod); err != nil {
				logger.Info("Failed to sync the meta role label.", "pod", pod.Name, "error", err)
			}
		}
	}
}

// Reconcile handles the pods of the meta service. Will add the metaLeaderLabel to the pods.
func (mpl *MetaPodRoleLabeler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, e error) {
	var pod corev1.Pod
	if err := mpl.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrlkit.RequeueIfError(client.IgnoreNotFound(err))
	}

	// If Pod is deleted or not running, then no need to sync.
	if utils.IsDeleted(&pod) || !utils.IsPodRunning(&pod) {
		return ctrlkit.NoRequeue()
	}

	// Ignore non-meta or non-running Pods.
	if pod.Status.PodIP == "" || !mpl.isRisingWaveMetaPod(&pod) {
		return ctrlkit.NoRequeue()
	}

	// Sync the label for the current Pod. If the current Pod is the new leader, then
	// aggressively sync the labels for all leader Pods.
	mpl.syncRoleLabels(ctx, &pod)

	// Sync every 2 seconds.
	return ctrlkit.RequeueAfter(2 * time.Second)
}

// SetupWithManager sets up the controller with the Manager.
func (mpl *MetaPodRoleLabeler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("meta-pod-role-labeler").
		For(&corev1.Pod{}).
		Complete(mpl)
}

// NewMetaPodRoleLabeler creates a new MetaPodRoleLabeler.
func NewMetaPodRoleLabeler(client client.Client) *MetaPodRoleLabeler {
	return &MetaPodRoleLabeler{
		Client: client,
	}
}
