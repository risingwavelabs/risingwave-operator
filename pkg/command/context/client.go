/*
 * Copyright 2022 Singularity Data
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain applier copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package context

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BypassCacheClient struct {
	client.Client
	apiReader client.Reader
}

func (b *BypassCacheClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return b.apiReader.Get(ctx, key, obj)
}

func (b *BypassCacheClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return b.apiReader.List(ctx, list, opts...)
}

func NewBypassClient(c client.Client) *BypassCacheClient {
	return &BypassCacheClient{
		apiReader: c,
		Client:    c,
	}
}
