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

package manager

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type ComponentManager interface {
	// CreateService ...
	CreateService(context.Context, client.Client, *v1alpha1.RisingWave) error

	// UpdateService will update service if service spec changed
	//if no change, do nothing
	UpdateService(context.Context, client.Client, *v1alpha1.RisingWave) (bool, error)

	// DeleteService ...
	DeleteService(context.Context, client.Client, *v1alpha1.RisingWave) error

	// CheckService will do health check. if not OK, should ensure service by EnsureService
	CheckService(context.Context, client.Client, *v1alpha1.RisingWave) (bool, error)

	// EnsureService block util the service is ready, or failed
	//if failed. return the error
	EnsureService(context.Context, client.Client, *v1alpha1.RisingWave) error

	Name() string
}
