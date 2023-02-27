package factory

import (
	"time"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type predicate[T client.Object, k any] struct {
	// Name of test/predicate
	Name string

	// Predicate function thare takes in the object,test case and returns a boolean.
	Fn func(obj T, testcase k) bool
}

func GetDeploymentPredicates() []predicate[*appsv1.Deployment, deploymentTestcase] {
	return []predicate[*appsv1.Deployment, deploymentTestcase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return HasLabels(obj, ComponentGroupLabels(tc.risingwave, tc.component, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return HasAnnotations(obj, ComponentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return MapContains(obj.Spec.Template.Labels, PodSelector(tc.risingwave, tc.component, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				if tc.restartAt != nil {
					return MapContains(obj.Spec.Template.Annotations, map[string]string{
						consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
					})
				} else {
					_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
					return !ok
				}
			},
		},
		{
			Name: "pod-template-works",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return MatchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				for _, s := range tc.group.ImagePullSecrets {
					if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
						return false
					}
				}
				return true
			},
		},
		{
			Name: "resources-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.Strategy, appsv1.DeploymentStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.Strategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestcase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

func GetSTSPredicates() []predicate[*appsv1.StatefulSet, stsTestcase] {
	return []predicate[*appsv1.StatefulSet, stsTestcase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return HasLabels(obj, ComponentGroupLabels(tc.risingwave, consts.ComponentCompute, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return HasAnnotations(obj, ComponentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return MapContains(obj.Spec.Template.Labels, PodSelector(tc.risingwave, consts.ComponentCompute, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				if tc.restartAt != nil {
					return MapContains(obj.Spec.Template.Annotations, map[string]string{
						consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
					})
				} else {
					_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
					return !ok
				}
			},
		},
		{
			Name: "pod-template-works",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return MatchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				for _, s := range tc.group.ImagePullSecrets {
					if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
						return false
					}
				}
				return true
			},
		},
		{
			Name: "resources-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, appsv1.StatefulSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
		{
			Name: "check-volume-mounts",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].VolumeMounts, tc.group.VolumeMounts, func(t *corev1.VolumeMount) string { return t.MountPath }, deepEqual[corev1.VolumeMount])
			},
		},
		{
			Name: "first-container-must-have-probes",
			Fn: func(obj *appsv1.StatefulSet, tc stsTestcase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

func GetClonesetPredicates() []predicate[*kruiseappsv1alpha1.CloneSet, clonesetTestcase] {
	return []predicate[*kruiseappsv1alpha1.CloneSet, clonesetTestcase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return HasLabels(obj, ComponentGroupLabels(tc.risingwave, tc.component, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return HasAnnotations(obj, ComponentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return MapContains(obj.Spec.Template.Labels, PodSelector(tc.risingwave, tc.component, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				if tc.restartAt != nil {
					return MapContains(obj.Spec.Template.Annotations, map[string]string{
						consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
					})
				} else {
					_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
					return !ok
				}
			},
		},
		{
			Name: "pod-template-works",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return MatchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				for _, s := range tc.group.ImagePullSecrets {
					if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
						return false
					}
				}
				return true
			},
		},
		{
			Name: "resources-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, kruiseappsv1alpha1.CloneSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc clonesetTestcase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

func GetAdvancedSTSPredicates() []predicate[*kruiseappsv1beta1.StatefulSet, advancedSTSTestcase] {
	return []predicate[*kruiseappsv1beta1.StatefulSet, advancedSTSTestcase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return HasLabels(obj, ComponentGroupLabels(tc.risingwave, consts.ComponentCompute, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return HasAnnotations(obj, ComponentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return MapContains(obj.Spec.Template.Labels, PodSelector(tc.risingwave, consts.ComponentCompute, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				if tc.restartAt != nil {
					return MapContains(obj.Spec.Template.Annotations, map[string]string{
						consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
					})
				} else {
					_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
					return !ok
				}
			},
		},
		{
			Name: "pod-template-works",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return MatchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				for _, s := range tc.group.ImagePullSecrets {
					if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
						return false
					}
				}
				return true
			},
		},
		{
			Name: "resources-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, kruiseappsv1beta1.StatefulSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
		{
			Name: "check-volume-mounts",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].VolumeMounts, tc.group.VolumeMounts, func(t *corev1.VolumeMount) string { return t.MountPath }, deepEqual[corev1.VolumeMount])
			},
		},
		{
			Name: "first-container-must-have-probes",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc advancedSTSTestcase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}
