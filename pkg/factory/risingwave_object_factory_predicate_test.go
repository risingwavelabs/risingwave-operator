// Copyright 2023 The fold Authors.
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

package factory

import (
	"strconv"
	"time"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

type predicate[T kubeObject, k testCaseType] struct {
	// Name of test/predicate.
	Name string

	// Predicate function that takes in the object, test case and returns a boolean.
	Fn func(obj T, testcase k) bool
}

// This function returns the base predicates used for the deployment objects.
func deploymentPredicates() []predicate[*appsv1.Deployment, deploymentTestCase] {
	return []predicate[*appsv1.Deployment, deploymentTestCase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return hasLabels(obj, componentGroupLabels(tc.risingwave, tc.component, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return hasAnnotations(obj, componentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return mapContains(obj.Spec.Template.Labels, podSelector(tc.risingwave, tc.component, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				if tc.restartAt != nil {
					return mapContains(obj.Spec.Template.Annotations, map[string]string{
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
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return matchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
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
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.Strategy, appsv1.DeploymentStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.Strategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *appsv1.Deployment, tc deploymentTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

// This function returns the predicates used for to compare deployment objects for the compactor component.
// It inherits from the base deployment predicates and further additional predicates can be added for compactor.
func compactorDeploymentPredicates() []predicate[*appsv1.Deployment, deploymentTestCase] {
	return deploymentPredicates()
}

// This function returns the predicates used for to compare deployment objects for the Frontend component.
// It inherits from the base deployment predicates and further additional predicates can be added for Frontend.
func frontendDeploymentPredicates() []predicate[*appsv1.Deployment, deploymentTestCase] {
	return deploymentPredicates()
}

// This function returns the predicates used for the meta statefulset predicates.
func metaStatefulSetPredicates() []predicate[*appsv1.StatefulSet, metaStatefulSetTestCase] {
	return []predicate[*appsv1.StatefulSet, metaStatefulSetTestCase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return hasLabels(obj, componentGroupLabels(tc.risingwave, tc.component, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return hasAnnotations(obj, componentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return mapContains(obj.Spec.Template.Labels, podSelector(tc.risingwave, tc.component, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				if tc.restartAt != nil {
					return mapContains(obj.Spec.Template.Annotations, map[string]string{
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
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return matchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
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
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, appsv1.StatefulSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *appsv1.StatefulSet, tc metaStatefulSetTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

// This function returns the predicates used for the base statefulset predicates.
func getSTSPredicates() []predicate[*appsv1.StatefulSet, computeStatefulSetTestCase] {
	return []predicate[*appsv1.StatefulSet, computeStatefulSetTestCase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return hasLabels(obj, componentGroupLabels(tc.risingwave, consts.ComponentCompute, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return hasAnnotations(obj, componentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return mapContains(obj.Spec.Template.Labels, podSelector(tc.risingwave, consts.ComponentCompute, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				if tc.restartAt != nil {
					return mapContains(obj.Spec.Template.Annotations, map[string]string{
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
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return matchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
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
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, appsv1.StatefulSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
		{
			Name: "check-volume-mounts",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].VolumeMounts, tc.group.VolumeMounts, func(t *corev1.VolumeMount) string { return t.MountPath }, deepEqual[corev1.VolumeMount])
			},
		},
		{
			Name: "first-container-must-have-probes",
			Fn: func(obj *appsv1.StatefulSet, tc computeStatefulSetTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

// This function returns the predicates used for to compare stateful objects for the compute component.
// It inherits from the base statefulset predicates and further additional predicates can be added for compute.
func computeStatefulSetPredicates() []predicate[*appsv1.StatefulSet, computeStatefulSetTestCase] {
	return getSTSPredicates()
}

// This function returns the base predicates used for the CloneSet objects.
func getCloneSetPredicates() []predicate[*kruiseappsv1alpha1.CloneSet, cloneSetTestCase] {
	return []predicate[*kruiseappsv1alpha1.CloneSet, cloneSetTestCase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return hasLabels(obj, componentGroupLabels(tc.risingwave, tc.component, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return hasAnnotations(obj, componentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return mapContains(obj.Spec.Template.Labels, podSelector(tc.risingwave, tc.component, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				if tc.restartAt != nil {
					return mapContains(obj.Spec.Template.Annotations, map[string]string{
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
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return matchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
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
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, kruiseappsv1alpha1.CloneSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *kruiseappsv1alpha1.CloneSet, tc cloneSetTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

// This function returns the predicates used for to compare CloneSet objects for the compactor component.
// It inherits from the base CloneSet predicates and further additional predicates can be added for compactor.
func compactorCloneSetPredicates() []predicate[*kruiseappsv1alpha1.CloneSet, cloneSetTestCase] {
	return getCloneSetPredicates()
}

// This function returns the predicates used for to compare CloneSet objects for the frontend component.
// It inherits from the base CloneSet predicates and further additional predicates can be added for frontend.
func frontendCloneSetPredicates() []predicate[*kruiseappsv1alpha1.CloneSet, cloneSetTestCase] {
	return getCloneSetPredicates()
}

// This function returns the predicates used for the base advanced statefulset predicates.
func getAdvancedSTSPredicates() []predicate[*kruiseappsv1beta1.StatefulSet, computeAdvancedSTSTestCase] {
	return []predicate[*kruiseappsv1beta1.StatefulSet, computeAdvancedSTSTestCase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return hasLabels(obj, componentGroupLabels(tc.risingwave, consts.ComponentCompute, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return hasAnnotations(obj, componentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return mapContains(obj.Spec.Template.Labels, podSelector(tc.risingwave, consts.ComponentCompute, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				if tc.restartAt != nil {
					return mapContains(obj.Spec.Template.Annotations, map[string]string{
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
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return matchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
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
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, kruiseappsv1beta1.StatefulSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
		{
			Name: "check-volume-mounts",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].VolumeMounts, tc.group.VolumeMounts, func(t *corev1.VolumeMount) string { return t.MountPath }, deepEqual[corev1.VolumeMount])
			},
		},
		{
			Name: "first-container-must-have-probes",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc computeAdvancedSTSTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

// This function returns the predicates used to compare Advanced STS objects for the compute component.
// It inherits from the base Advanced STS predicates and further additional predicates can be added for compute.
func computeAdvancedSTSPredicates() []predicate[*kruiseappsv1beta1.StatefulSet, computeAdvancedSTSTestCase] {
	return getAdvancedSTSPredicates()
}

// This function returns the predicates used for the meta statefulset predicates.
func metaAdvancedSTSPredicates() []predicate[*kruiseappsv1beta1.StatefulSet, metaAdvancedSTSTestCase] {
	return []predicate[*kruiseappsv1beta1.StatefulSet, metaAdvancedSTSTestCase]{
		{
			Name: "namespace-equals",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "labels-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return hasLabels(obj, componentGroupLabels(tc.risingwave, tc.component, &tc.group.Name, true), true)
			},
		},
		{
			Name: "annotations-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return hasAnnotations(obj, componentGroupAnnotations(tc.risingwave, &tc.group.Name), true)
			},
		},
		{
			Name: "replicas-equal",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return *obj.Spec.Replicas == tc.group.Replicas
			},
		},

		{
			Name: "pod-template-labels-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return mapContains(obj.Spec.Template.Labels, podSelector(tc.risingwave, tc.component, &tc.group.Name))
			},
		},
		{
			Name: "pod-template-annotations-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				if tc.restartAt != nil {
					return mapContains(obj.Spec.Template.Annotations, map[string]string{
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
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				if tc.group.PodTemplate != nil {
					temp := tc.podTemplate[*tc.group.PodTemplate].Template
					return matchesPodTemplate(&obj.Spec.Template, &temp)
				} else {
					return true
				}
			},
		},
		{
			Name: "image-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
			},
		},
		{
			Name: "image-pull-policy-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
			},
		},
		{
			Name: "image-pull-secrets-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
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
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "node-selector-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
			},
		},
		{
			Name: "tolerations-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Tolerations, tc.group.Tolerations)
			},
		},
		{
			Name: "priority-class-name-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return obj.Spec.Template.Spec.PriorityClassName == tc.group.PriorityClassName
			},
		},
		{
			Name: "security-context-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.SecurityContext, tc.group.SecurityContext)
			},
		},
		{
			Name: "dns-config-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.DNSConfig, tc.group.DNSConfig)
			},
		},
		{
			Name: "termination-grace-period-seconds-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				if tc.group.TerminationGracePeriodSeconds != nil {
					return *obj.Spec.Template.Spec.TerminationGracePeriodSeconds == *tc.group.TerminationGracePeriodSeconds
				} else {
					return true
				}
			},
		},
		{
			Name: "upgrade-strategy-match",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				if tc.expectedUpgradeStrategy == nil {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, kruiseappsv1beta1.StatefulSetUpdateStrategy{})
				} else {
					return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
				}
			},
		}, {
			Name: "first-container-must-have-probes",
			Fn: func(obj *kruiseappsv1beta1.StatefulSet, tc metaAdvancedSTSTestCase) bool {
				container := &obj.Spec.Template.Spec.Containers[0]
				return container.LivenessProbe != nil && container.ReadinessProbe != nil
			},
		},
	}
}

func servicesPredicates() []predicate[*corev1.Service, servicesTestCase] {
	return []predicate[*corev1.Service, servicesTestCase]{
		{
			Name: "controlled-by-risingwave",
			Fn: func(obj *corev1.Service, testcase servicesTestCase) bool {
				return controlledBy(testcase.risingwave, obj)
			},
		},
		{
			Name: "namespace-equals",
			Fn: func(obj *corev1.Service, testcase servicesTestCase) bool {
				return obj.Namespace == testcase.risingwave.Namespace
			},
		},
		{
			Name: "ports-equal",
			Fn: func(obj *corev1.Service, testcase servicesTestCase) bool {
				return hasTCPServicePorts(obj, testcase.ports)
			},
		},
		{
			Name: "service-type-match",
			Fn: func(obj *corev1.Service, testcase servicesTestCase) bool {
				return isServiceType(obj, testcase.expectServiceType)
			},
		},
		{
			Name: "service-labels-match",
			Fn: func(obj *corev1.Service, testcase servicesTestCase) bool {
				return hasLabels(obj, componentLabels(testcase.risingwave, testcase.component, true), true)
			},
		},
		{
			Name: "selector-equals",
			Fn: func(obj *corev1.Service, testcase servicesTestCase) bool {
				return hasServiceSelector(obj, podSelector(testcase.risingwave, testcase.component, nil))
			},
		},
	}
}

func serviceMetadataPredicates() []predicate[*corev1.Service, serviceMetadataTestCase] {
	return []predicate[*corev1.Service, serviceMetadataTestCase]{
		{
			Name: "service-labels-match",
			Fn: func(obj *corev1.Service, testcase serviceMetadataTestCase) bool {
				return hasLabels(obj, componentLabels(testcase.risingwave, testcase.component, true), true)
			},
		},
		{
			Name: "service-annotations-match",
			Fn: func(obj *corev1.Service, testcase serviceMetadataTestCase) bool {
				return hasAnnotations(obj, componentAnnotations(testcase.risingwave, testcase.component), true)
			},
		},
	}
}

func objectStorageStatefulsetPredicates() []predicate[*appsv1.StatefulSet, objectStoragesTestCase] {
	return []predicate[*appsv1.StatefulSet, objectStoragesTestCase]{
		{
			Name: "hummock-args-match",
			Fn: func(obj *appsv1.StatefulSet, tc objectStoragesTestCase) bool {
				return lo.Contains(obj.Spec.Template.Spec.Containers[0].Args, tc.hummockArg)
			},
		},
		{
			Name: "env-vars-contains",
			Fn: func(obj *appsv1.StatefulSet, tc objectStoragesTestCase) bool {
				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
			},
		},
	}
}

func configMapPredicates() []predicate[*corev1.ConfigMap, configMapTestCase] {
	return []predicate[*corev1.ConfigMap, configMapTestCase]{
		{
			Name: "controlled-by-risingwave",
			Fn: func(obj *corev1.ConfigMap, tc configMapTestCase) bool {
				return controlledBy(tc.risingwave, obj)
			},
		},
		{
			Name: "namespace-equals",
			Fn: func(obj *corev1.ConfigMap, tc configMapTestCase) bool {
				return obj.Namespace == tc.risingwave.Namespace
			},
		},
		{
			Name: "configmap-labels-match",
			Fn: func(obj *corev1.ConfigMap, tc configMapTestCase) bool {
				return hasLabels(obj, componentLabels(tc.risingwave, consts.ComponentConfig, false), true)
			},
		},
		{
			Name: "configmap-data-match",
			Fn: func(obj *corev1.ConfigMap, tc configMapTestCase) bool {
				return mapEquals(obj.Data, map[string]string{
					risingWaveConfigMapKey: lo.If(tc.configVal == "", "").Else(tc.configVal),
				})
			},
		},
	}
}

func metaStoragePredicates() []predicate[*appsv1.StatefulSet, metaStorageTestCase] {
	return []predicate[*appsv1.StatefulSet, metaStorageTestCase]{
		{
			Name: "args-match",
			Fn: func(obj *appsv1.StatefulSet, tc metaStorageTestCase) bool {
				return containsStringSlice(obj.Spec.Template.Spec.Containers[0].Args, tc.args)
			},
		},
		{
			Name: "env-vars-contains",
			Fn: func(obj *appsv1.StatefulSet, tc metaStorageTestCase) bool {
				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
			},
		},
	}
}

func serviceMonitorPredicates() []predicate[*prometheusv1.ServiceMonitor, baseTestCase] {
	return []predicate[*prometheusv1.ServiceMonitor, baseTestCase]{
		{
			Name: "owned",
			Fn: func(obj *prometheusv1.ServiceMonitor, tc baseTestCase) bool {
				return controlledBy(tc.risingwave, obj)
			},
		},
		{
			Name: "has-labels",
			Fn: func(obj *prometheusv1.ServiceMonitor, tc baseTestCase) bool {
				return hasLabels(obj, map[string]string{
					consts.LabelRisingWaveName:       tc.risingwave.Name,
					consts.LabelRisingWaveGeneration: strconv.FormatInt(tc.risingwave.Generation, 10),
				}, true)
			},
		},
		{
			Name: "target-labels",
			Fn: func(obj *prometheusv1.ServiceMonitor, tc baseTestCase) bool {
				return listContains(obj.Spec.TargetLabels, []string{
					consts.LabelRisingWaveName,
					consts.LabelRisingWaveComponent,
					consts.LabelRisingWaveGroup,
				})
			},
		},
		{
			Name: "scrape-port-metrics",
			Fn: func(obj *prometheusv1.ServiceMonitor, tc baseTestCase) bool {
				return len(obj.Spec.Endpoints) > 0 && obj.Spec.Endpoints[0].Port == consts.PortMetrics
			},
		},
	}

}
