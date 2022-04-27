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
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MetaNodeComponentName(name string) string {
	return fmt.Sprintf("%s-meta-node", name)
}

func ComputeNodeComponentName(name string) string {
	return fmt.Sprintf("%s-compute-node", name)
}

func FrontendComponentName(name string) string {
	return fmt.Sprintf("%s-frontend", name)
}

func MinIOComponentName(name string) string {
	return fmt.Sprintf("%s-minio", name)
}

func CreateOrUpdateObject(ctx context.Context, c client.Client, o client.Object) error {
	var existing = o.DeepCopyObject().(client.Object)
	err := c.Get(ctx, client.ObjectKeyFromObject(o), existing)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Create(ctx, o)
		}
		return fmt.Errorf("get existed object failed, %w", err)
	}

	o.SetResourceVersion(existing.GetResourceVersion())
	o.SetCreationTimestamp(existing.GetCreationTimestamp())
	o.SetGeneration(existing.GetGeneration())

	return c.Update(ctx, o)
}

func CreateIfNotFound(ctx context.Context, c client.Client, o client.Object) error {
	var existing = o.DeepCopyObject().(client.Object)
	err := c.Get(ctx, client.ObjectKeyFromObject(o), existing)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Create(ctx, o)
		}
		return fmt.Errorf("get existed object failed, %w", err)
	}
	return nil
}

func DeleteObject(ctx context.Context, c client.Client, o client.Object) error {
	err := c.Delete(ctx, o)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func DeleteObjectByObjectKey(ctx context.Context, c client.Client, key client.ObjectKey, existing client.Object) error {
	err := c.Get(ctx, key, existing)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	err = DeleteObject(ctx, c, existing)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}
