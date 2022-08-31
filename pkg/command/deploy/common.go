/*
 * Copyright 2022 Singularity Data
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

package deploy

import "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"

type GroupReplicas struct {
	Compute   []ReplicaInfo
	Frontend  []ReplicaInfo
	Compactor []ReplicaInfo
	Meta      []ReplicaInfo
}

type ReplicaInfo struct {
	GroupName string
	Replicas  int32
}

const (
	ReplicaAnnotation       = "replicas.old"
	GlobalReplicaAnnotation = "replicas.global.old"
)

// Check if instance has already been stopped.
func doesReplicaAnnotationExist(instance *v1alpha1.RisingWave) bool {
	if _, ok := instance.Annotations[ReplicaAnnotation]; !ok {
		return false
	}

	if _, ok := instance.Annotations[GlobalReplicaAnnotation]; !ok {
		return false
	}

	return true
}
