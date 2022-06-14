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

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceMonitorContextKey struct{}

func ContextWithServiceMonitorFlag(ctx context.Context, useServiceMonitor bool) context.Context {
	return context.WithValue(ctx, ServiceMonitorContextKey{}, useServiceMonitor)
}

func ServiceMonitorFlagFromContext(ctx context.Context) bool {
	if v, ok := ctx.Value(ServiceMonitorContextKey{}).(bool); ok {
		return v
	}
	return false
}

func GenerateServiceMonitor(serviceName string, portName string, rw *v1alpha1.RisingWave) *prometheusv1.ServiceMonitor {

	var sm = prometheusv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-sm", serviceName),
			Namespace: rw.Namespace,
			Labels: map[string]string{
				"Name": fmt.Sprintf("%s-sm", serviceName),
			},
		},
		Spec: prometheusv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					ServiceNameKey: serviceName,
					UIDKey:         string(rw.UID),
				},
			},
			NamespaceSelector: prometheusv1.NamespaceSelector{
				MatchNames: []string{rw.Namespace},
			},
			Endpoints: []prometheusv1.Endpoint{
				{
					Port: portName,
				},
			},
		},
	}

	return &sm
}

func DeleteServiceMonitor(ctx context.Context, c client.Client, serviceName string, rw *v1alpha1.RisingWave) error {
	err := DeleteObjectByObjectKey(ctx, c, types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      fmt.Sprintf("%s-sm", serviceName),
	}, &prometheusv1.ServiceMonitor{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

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

func CompactorNodeComponentName(name string) string {
	return fmt.Sprintf("%s-compactor-node", name)
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

func StoreParam(rw *v1alpha1.RisingWave) string {
	storage := rw.Spec.ObjectStorage
	switch {
	case storage.S3 != nil:
		var bucket = *storage.S3.Bucket
		return fmt.Sprintf("hummock+s3://%s", bucket)
	case storage.Memory:
		return "in-memory"
	case storage.MinIO != nil:
		return fmt.Sprintf("hummock+minio://hummock:12345678@%s:%d/hummock001", MinIOComponentName(rw.Name), v1alpha1.MinIOServerPort)
	default:
		return fmt.Sprint("no-support")
	}
}
