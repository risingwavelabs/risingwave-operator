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

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/s3/types"
)

type S3Manager struct {
	Provider string

	client types.Client

	secret v1.Secret
}

func (m S3Manager) Name() string {
	return fmt.Sprintf("%s-%s", S3Name, m.Provider)
}

func (m S3Manager) CreateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	log := logger.FromContext(ctx).WithValues("provider", m.Provider)
	log.V(1).Info("Begin to create S3 bucket")
	if m.client == nil {
		f, err := types.GetClientFun(m.Provider)
		if err != nil {
			return fmt.Errorf("get client create func failed, %w", err)
		}
		c, err := f(m.secret)
		if err != nil {
			return fmt.Errorf("create client failed, %w", err)
		}
		m.client = c
	}

	// TODO: maybe generate bucket by another way
	var bucket = string(rw.UID)
	err := m.client.CreateBucket(ctx, bucket)
	if err != nil {
		return fmt.Errorf("create bucket failed, %w", err)
	}
	rw.Status.ObjectStorage.S3 = &v1alpha1.S3Status{
		Bucket: bucket,
	}
	return nil
}

func (m S3Manager) UpdateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	return false, nil
}

func (m S3Manager) DeleteService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	log := logger.FromContext(ctx).WithValues("provider", m.Provider)
	log.V(1).Info("Begin to delete S3 bucket")

	var bucket = string(rw.UID)
	err := m.client.DeleteBucket(ctx, bucket)
	if err != nil {
		return fmt.Errorf("delete bucket failed, %w", err)
	}
	rw.Status.ObjectStorage.S3 = &v1alpha1.S3Status{}
	return nil
}

func (m S3Manager) CheckService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	if rw.Status.ObjectStorage.S3 != nil && len(rw.Status.ObjectStorage.S3.Bucket) != 0 {
		return true, nil
	}
	return false, nil
}

func (m S3Manager) EnsureService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	return nil
}

func NewS3Manager(provider string, secret v1.Secret) *S3Manager {
	return &S3Manager{
		secret:   secret,
		Provider: provider,
	}
}
