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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNamespacedName returns the NamespacedName of the given object.
func GetNamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

// IsDeleted returns true when object's deletion timestamp isn't null.
func IsDeleted(obj client.Object) bool {
	return obj != nil && !obj.GetDeletionTimestamp().IsZero()
}

// GetVersionFromImage return the version from spec.global.image.
func GetVersionFromImage(image string) string {
	if image == "" {
		return ""
	}

	lastRepoIdx := strings.LastIndex(image, "/")
	lastTagIdx := strings.LastIndex(image[lastRepoIdx+1:], ":")
	if lastTagIdx < 0 {
		return "latest"
	} else {
		return image[lastRepoIdx+lastTagIdx+2:]
	}
}

// GetContainerFromPod gets a pointer to the container with the same name. Nil is returned when
// the container isn't found.
func GetContainerFromPod(pod *corev1.Pod, name string) *corev1.Container {
	if pod == nil {
		return nil
	}

	for _, container := range pod.Spec.Containers {
		if container.Name == name {
			return &container
		}
	}

	return nil
}

// GetPortFromContainer gets the specified port from the container. If the port's not found,
// it returns a false.
func GetPortFromContainer(container *corev1.Container, name string) (int32, bool) {
	if container == nil {
		return 0, false
	}

	for _, containerPort := range container.Ports {
		if containerPort.Name == name {
			return containerPort.ContainerPort, true
		}
	}

	return 0, false
}
