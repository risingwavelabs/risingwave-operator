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

package v1alpha1

import (
	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// RisingWaveNodeContainer determines the container specs of a RisingWave node.
type RisingWaveNodeContainer struct {
	// Container image name.
	// More info: https://kubernetes.io/docs/concepts/containers/images
	// This field is optional to allow higher level config management to default or override
	// container images in workload controllers like Deployments and StatefulSets.
	// +optional
	Image string `json:"image,omitempty" protobuf:"bytes,2,opt,name=image"`

	// List of sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	// Cannot be updated.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty" protobuf:"bytes,19,rep,name=envFrom"`
	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchMergeKey:"name" patchStrategy:"merge" protobuf:"bytes,7,rep,name=env"`

	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`

	// Pod volumes to mount into the container's filesystem.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty" patchMergeKey:"mountPath" patchStrategy:"merge" protobuf:"bytes,9,rep,name=volumeMounts"`

	// volumeDevices is the list of block devices to be used by the container.
	// +patchMergeKey=devicePath
	// +patchStrategy=merge
	// +optional
	VolumeDevices []corev1.VolumeDevice `json:"volumeDevices,omitempty" patchMergeKey:"devicePath" patchStrategy:"merge" protobuf:"bytes,21,rep,name=volumeDevices"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`

	// SecurityContext defines the security options the container should be run with.
	// If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	// +optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty" protobuf:"bytes,15,opt,name=securityContext"`
}

// RisingWaveNodePodTemplateSpec is a template for a RisingWave's Pod.
type RisingWaveNodePodTemplateSpec struct {
	RisingWaveNodeContainer `json:",inline"`

	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []corev1.Volume `json:"volumes,omitempty" patchMergeKey:"name" patchStrategy:"merge,retainKeys" protobuf:"bytes,1,rep,name=volumes"`

	// Optional duration in seconds the pod may be active on the node relative to
	// StartTime before the system will actively try to mark it failed and kill associated containers.
	// Value must be a positive integer.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty" protobuf:"varint,5,opt,name=activeDeadlineSeconds"`

	// Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
	// Value must be non-negative integer. The value zero indicates stop immediately via
	// the kill signal (no opportunity to shut down).
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 30 seconds.
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty" protobuf:"varint,4,opt,name=terminationGracePeriodSeconds"`

	// Set DNS policy for the pod.
	// Defaults to "ClusterFirst".
	// Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy
	// explicitly to 'ClusterFirstWithHostNet'.
	// +optional
	// +kubebuilder:validation:Enum=ClusterFirst;ClusterFirstWithHostNet;Default;None
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty" protobuf:"bytes,6,opt,name=dnsPolicy,casttype=DNSPolicy"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`

	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`

	// AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.
	// +optional
	AutomountServiceAccountToken *bool `json:"automountServiceAccountToken,omitempty" protobuf:"varint,21,opt,name=automountServiceAccountToken"`

	// Use the host's pid namespace.
	// Optional: Default to false.
	// +k8s:conversion-gen=false
	// +optional
	HostPID bool `json:"hostPID,omitempty" protobuf:"varint,12,opt,name=hostPID"`
	// Use the host's ipc namespace.
	// Optional: Default to false.
	// +k8s:conversion-gen=false
	// +optional
	HostIPC bool `json:"hostIPC,omitempty" protobuf:"varint,13,opt,name=hostIPC"`

	// Share a single process namespace between all of the containers in a pod.
	// When this is set containers will be able to view and signal processes from other containers
	// in the same pod, and the first process in each container will not be assigned PID 1.
	// HostPID and ShareProcessNamespace cannot both be set.
	// Optional: Default to false.
	// +k8s:conversion-gen=false
	// +optional
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty" protobuf:"varint,27,opt,name=shareProcessNamespace"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty" protobuf:"bytes,14,opt,name=securityContext"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	// If specified, these secrets will be passed to individual puller implementations for them to use.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchMergeKey:"name" patchStrategy:"merge" protobuf:"bytes,15,rep,name=imagePullSecrets"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`

	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`

	// HostAliases is an optional list of hosts and IPs that will be injected into the pod's hosts
	// file if specified. This is only valid for non-hostNetwork pods.
	// +optional
	// +patchMergeKey=ip
	// +patchStrategy=merge
	HostAliases []corev1.HostAlias `json:"hostAliases,omitempty" patchMergeKey:"ip" patchStrategy:"merge" protobuf:"bytes,23,rep,name=hostAliases"`

	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`

	// The priority value. Various system components use this field to find the
	// priority of the pod. When Priority Admission Controller is enabled, it
	// prevents users from setting this field. The admission controller populates
	// this field from PriorityClassName.
	// The higher the value, the higher the priority.
	// +optional
	Priority *int32 `json:"priority,omitempty" protobuf:"bytes,25,opt,name=priority"`

	// Specifies the DNS parameters of a pod.
	// Parameters specified here will be merged to the generated DNS
	// configuration based on DNSPolicy.
	// +optional
	DNSConfig *corev1.PodDNSConfig `json:"dnsConfig,omitempty" protobuf:"bytes,26,opt,name=dnsConfig"`

	// RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used
	// to run this pod.  If no RuntimeClass resource matches the named class, the pod will not be run.
	// If unset or empty, the "legacy" RuntimeClass will be used, which is an implicit class with an
	// empty definition that uses the default runtime handler.
	// More info: https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class
	// +optional
	RuntimeClassName *string `json:"runtimeClassName,omitempty" protobuf:"bytes,29,opt,name=runtimeClassName"`

	// PreemptionPolicy is the Policy for preempting pods with lower priority.
	// One of Never, PreemptLowerPriority.
	// Defaults to PreemptLowerPriority if unset.
	// +optional
	PreemptionPolicy *corev1.PreemptionPolicy `json:"preemptionPolicy,omitempty" protobuf:"bytes,31,opt,name=preemptionPolicy"`

	// TopologySpreadConstraints describes how a group of pods ought to spread across topology
	// domains. Scheduler will schedule pods in a way which abides by the constraints.
	// All topologySpreadConstraints are ANDed.
	// +optional
	// +patchMergeKey=topologyKey
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=topologyKey
	// +listMapKey=whenUnsatisfiable
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty" patchMergeKey:"topologyKey" patchStrategy:"merge" protobuf:"bytes,33,opt,name=topologySpreadConstraints"`

	// If true the pod's hostname will be configured as the pod's FQDN, rather than the leaf name (the default).
	// In Linux containers, this means setting the FQDN in the hostname field of the kernel (the nodename field of struct utsname).
	// In Windows containers, this means setting the registry value of hostname for the registry key HKEY_LOCAL_MACHINE\\SYSTEM\\CurrentControlSet\\Services\\Tcpip\\Parameters to FQDN.
	// If a pod does not have FQDN, this has no effect.
	// Default to false.
	// +optional
	SetHostnameAsFQDN *bool `json:"setHostnameAsFQDN,omitempty" protobuf:"varint,35,opt,name=setHostnameAsFQDN"`

	// Specifies the OS of the containers in the pod.
	// Some pod and container fields are restricted if this is set.
	//
	// If the OS field is set to linux, the following fields must be unset:
	// -securityContext.windowsOptions
	//
	// If the OS field is set to windows, following fields must be unset:
	// - spec.hostPID
	// - spec.hostIPC
	// - spec.hostUsers
	// - spec.securityContext.seLinuxOptions
	// - spec.securityContext.seccompProfile
	// - spec.securityContext.fsGroup
	// - spec.securityContext.fsGroupChangePolicy
	// - spec.securityContext.sysctls
	// - spec.shareProcessNamespace
	// - spec.securityContext.runAsUser
	// - spec.securityContext.runAsGroup
	// - spec.securityContext.supplementalGroups
	// - spec.containers[*].securityContext.seLinuxOptions
	// - spec.containers[*].securityContext.seccompProfile
	// - spec.containers[*].securityContext.capabilities
	// - spec.containers[*].securityContext.readOnlyRootFilesystem
	// - spec.containers[*].securityContext.privileged
	// - spec.containers[*].securityContext.allowPrivilegeEscalation
	// - spec.containers[*].securityContext.procMount
	// - spec.containers[*].securityContext.runAsUser
	// - spec.containers[*].securityContext.runAsGroup
	// +optional
	OS *corev1.PodOS `json:"os,omitempty" protobuf:"bytes,36,opt,name=os"`

	// Use the host's user namespace.
	// Optional: Default to true.
	// If set to true or not present, the pod will be run in the host user namespace, useful
	// for when the pod needs a feature only available to the host user namespace, such as
	// loading a kernel module with CAP_SYS_MODULE.
	// When set to false, a new userns is created for the pod. Setting false is useful for
	// mitigating container breakout vulnerabilities even allowing users to run their
	// containers as root without actually having root privileges on the host.
	// This field is alpha-level and is only honored by servers that enable the UserNamespacesSupport feature.
	// +k8s:conversion-gen=false
	// +optional
	HostUsers *bool `json:"hostUsers,omitempty" protobuf:"bytes,37,opt,name=hostUsers"`

	// Additional containers to run in the same Pod. The containers will be appended to the Pod's containers array in order.
	// + optional
	AdditionalContainers []corev1.Container `json:"additionalContainers,omitempty"`
}

// RisingWaveNodePodTemplate determines the Pod specs of a RisingWave node.
type RisingWaveNodePodTemplate struct {
	// PartialObjectMeta tells the operator to add the specified metadata onto the Pod.
	ObjectMeta PartialObjectMeta `json:"metadata,omitempty"`

	// RisingWaveNodePodTemplateSpec determines the Pod spec to start the RisingWave pod.
	Spec RisingWaveNodePodTemplateSpec `json:"spec,omitempty"`
}

// RisingWaveNodeGroupUpgradeStrategyType is the type of upgrade strategies used in RisingWave.
type RisingWaveNodeGroupUpgradeStrategyType string

// Valid values of RisingWaveNodeGroupUpgradeStrategyType.
const (
	RisingWaveUpgradeStrategyTypeRecreate          RisingWaveNodeGroupUpgradeStrategyType = "Recreate"
	RisingWaveUpgradeStrategyTypeRollingUpdate     RisingWaveNodeGroupUpgradeStrategyType = "RollingUpdate"
	RisingWaveUpgradeStrategyTypeInPlaceIfPossible RisingWaveNodeGroupUpgradeStrategyType = "InPlaceIfPossible"
	RisingWaveUpgradeStrategyTypeInPlaceOnly       RisingWaveNodeGroupUpgradeStrategyType = "InPlaceOnly"
)

// RisingWaveNodeGroupRollingUpdate is the spec to define rolling update strategies.
type RisingWaveNodeGroupRollingUpdate struct {
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding down.
	// Defaults to 25%.
	// +optional
	// +kubebuilder:default="25%"
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"bytes,1,opt,name=maxUnavailable"`

	// Partition is the desired number of pods in old revisions.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up by default.
	// It means when partition is set during pods updating, (replicas - partition value) number of pods will be updated.
	// Default value is 0.
	// +optional
	// +kubebuilder:default=0
	Partition *intstr.IntOrString `json:"partition,omitempty"`

	// The maximum number of pods that can be scheduled above the desired replicas during update or specified delete.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up.
	// Defaults to 0.
	// +optional
	// +kubebuilder:default=0
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
}

// RisingWaveNodeGroupUpgradeStrategy is the spec of upgrade strategy used by RisingWave.
type RisingWaveNodeGroupUpgradeStrategy struct {
	// Type of upgrade. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	// +kubebuilder:default=RollingUpdate
	// +kubebuilder:validation:Enum=Recreate;RollingUpdate;InPlaceIfPossible;InPlaceOnly
	Type RisingWaveNodeGroupUpgradeStrategyType `json:"type,omitempty"`

	// Rolling update config params. Present only if DeploymentStrategyType = RollingUpdate.
	// +optional
	RollingUpdate *RisingWaveNodeGroupRollingUpdate `json:"rollingUpdate,omitempty"`

	// InPlaceUpdateStrategy contains strategies for in-place update.
	// +optional
	InPlaceUpdateStrategy *kruisepubs.InPlaceUpdateStrategy `json:"inPlaceUpdateStrategy,omitempty"`
}

// RisingWaveNodeGroup is the definition of a group of RisingWave nodes of the same component.
type RisingWaveNodeGroup struct {
	// Name of the node group.
	// +kubebuilder:default=""
	// +kubebuilder:validation:Pattern="^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// Replicas of Pods in this group.
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas,omitempty"`

	// RestartAt is the time that the Pods under the group should be restarted. Setting a value on this field will
	// trigger a full recreation of the Pods. Defaults to nil.
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Configuration determines the configuration to be used for the RisingWave nodes.
	// +optional
	Configuration *RisingWaveNodeConfiguration `json:"configuration,omitempty"`

	// Upgrade strategy for the components. By default, it is the same as the
	// workload's default strategy that the component is deployed with.
	// Note: the maxSurge will not take effect for the compute component.
	// +optional
	// +patchStrategy=retainKeys
	UpgradeStrategy RisingWaveNodeGroupUpgradeStrategy `json:"upgradeStrategy,omitempty"`

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,9,opt,name=minReadySeconds"`

	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// +optional
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`

	// persistentVolumeClaimRetentionPolicy describes the lifecycle of persistent
	// volume claims created from volumeClaimTemplates. By default, all persistent
	// volume claims are created as needed and retained until manually deleted. This
	// policy allows the lifecycle to be altered, for example by deleting persistent
	// volume claims when their stateful set is deleted, or when their pod is scaled
	// down. This requires the StatefulSetAutoDeletePVC feature gate to be enabled,
	// which is alpha.
	// +optional
	PersistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty" protobuf:"bytes,10,opt,name=persistentVolumeClaimRetentionPolicy"`

	// The maximum time in seconds for a deployment to make progress before it
	// is considered to be failed. The deployment controller will continue to
	// process failed deployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the deployment status. Note that progress will
	// not be estimated during the time a deployment is paused. Defaults to 600s.
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,9,opt,name=progressDeadlineSeconds"`

	// Template tells how the Pod should be started. It is an optional field. If it's empty, then the pod template in
	// the first-level fields under spec will be used.
	// +optional
	Template RisingWaveNodePodTemplate `json:"template,omitempty"`
}

// RisingWaveComponent determines how a RisingWave component is deployed.
type RisingWaveComponent struct {
	// LogLevel controls the log level of the running nodes. It can be in any format that the underlying component supports,
	// e.g., in the RUST_LOG format for Rust programs. Defaults to INFO.
	// +kubebuilder:default=INFO
	LogLevel string `json:"logLevel,omitempty"`

	// DisallowPrintStackTraces determines if the stack traces are allowed to print when panic happens. This options applies
	// to both Rust and Java programs. Defaults to false.
	// +optional
	DisallowPrintStackTraces *bool `json:"disallowPrintStackTraces,omitempty"`

	// NodeGroups of the component deployment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NodeGroups []RisingWaveNodeGroup `json:"nodeGroups,omitempty"`
}

// WorkloadReplicaStatus is a common structure for replica status of some workload.
type WorkloadReplicaStatus struct {
	// Replicas is the declared replicas of the workload.
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the ready replicas of the workload.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// AvailableReplicas is the available replicas of the workload.
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// UpdatedReplicas is the update replicas of the workload.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// UnavailableReplicas is the unavailable replicas of the workload.
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`
}

// RisingWaveNodeGroupStatus is the status of a node group.
type RisingWaveNodeGroupStatus struct {
	// Name of the node group.
	Name string `json:"name"`

	// WorkloadReplicaStatus is the replica status of the node group.
	WorkloadReplicaStatus `json:",inline"`

	// Existence status of the node group.
	Exists bool `json:"exists,omitempty"`
}

// RisingWaveComponentStatus is the status of a component.
type RisingWaveComponentStatus struct {
	// Total is the replica status of the component.
	Total WorkloadReplicaStatus `json:"total,omitempty"`

	// NodeGroups are the status list of all declared node groups of some component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NodeGroups []RisingWaveNodeGroupStatus `json:"nodeGroups,omitempty"`
}
