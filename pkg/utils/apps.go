/*
 * Copyright 2023 RisingWave Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

// IsDeploymentRolledOut returns true when the Deployment is rolled out.
func IsDeploymentRolledOut(deploy *appsv1.Deployment) bool {
	if deploy == nil {
		return false
	}
	if deploy.Status.ObservedGeneration < deploy.Generation {
		return false
	}
	for _, cond := range deploy.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing {
			if cond.Reason == "ProgressDeadlineExceeded" {
				return false
			}
		}
	}
	if deploy.Spec.Replicas != nil && deploy.Status.UpdatedReplicas < *deploy.Spec.Replicas {
		return false
	}
	if deploy.Status.Replicas > deploy.Status.UpdatedReplicas {
		return false
	}
	if deploy.Status.AvailableReplicas < deploy.Status.UpdatedReplicas {
		return false
	}
	return true
}

// IsStatefulSetRolledOut returns true when the StatefulSet is rolled out.
func IsStatefulSetRolledOut(statefulSet *appsv1.StatefulSet) bool {
	if statefulSet == nil {
		return false
	}
	if statefulSet.Status.ObservedGeneration < statefulSet.Generation {
		return false
	}
	if statefulSet.Spec.Replicas != nil && statefulSet.Status.UpdatedReplicas < *statefulSet.Spec.Replicas {
		return false
	}
	if statefulSet.Status.Replicas > statefulSet.Status.UpdatedReplicas {
		return false
	}
	if statefulSet.Status.AvailableReplicas < statefulSet.Status.UpdatedReplicas {
		return false
	}
	return true
}

// IsCloneSetRolledOut returns true when the OpenKruise CloneSet is rolled out.
func IsCloneSetRolledOut(cloneSet *kruiseappsv1alpha1.CloneSet) bool {
	if cloneSet == nil {
		return false
	}
	if cloneSet.Status.ObservedGeneration < cloneSet.Generation {
		return false
	}
	for _, cond := range cloneSet.Status.Conditions {
		if cond.Type == kruiseappsv1alpha1.CloneSetConditionFailedUpdate || cond.Type == kruiseappsv1alpha1.CloneSetConditionFailedScale {
			return false
		}
	}
	if cloneSet.Spec.Replicas != nil && cloneSet.Status.UpdatedReplicas < *cloneSet.Spec.Replicas {
		return false
	}
	if cloneSet.Status.Replicas > cloneSet.Status.UpdatedReplicas {
		return false
	}
	if cloneSet.Status.AvailableReplicas < cloneSet.Status.UpdatedReplicas {
		return false
	}
	return true
}

// IsAdvancedStatefulSetRolledOut returns true when the OpenKruise StatefulSet is rolled out.
func IsAdvancedStatefulSetRolledOut(statefulSet *kruiseappsv1beta1.StatefulSet) bool {
	if statefulSet == nil {
		return false
	}
	if statefulSet.Status.ObservedGeneration < statefulSet.Generation {
		return false
	}

	for _, cond := range statefulSet.Status.Conditions {
		if cond.Type == kruiseappsv1beta1.FailedCreatePod || cond.Type == kruiseappsv1beta1.FailedUpdatePod {
			return false
		}
	}
	if statefulSet.Spec.Replicas != nil && statefulSet.Status.UpdatedReplicas < *statefulSet.Spec.Replicas {
		return false
	}
	if statefulSet.Status.Replicas > statefulSet.Status.UpdatedReplicas {
		return false
	}
	if statefulSet.Status.AvailableReplicas < statefulSet.Status.UpdatedReplicas {
		return false
	}
	return true
}
