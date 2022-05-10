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

package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"

	types2 "github.com/singularity-data/risingwave-operator/pkg/s3/types"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	v1 "k8s.io/api/core/v1"

	"github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client *awss3.Client
	cfg    aws.Config
}

var _ types2.Client = &S3{}

func NewS3(opt types2.CreateOption) (types2.Client, error) {
	secret, ok := opt.(v1.Secret)
	if !ok {
		return nil, fmt.Errorf("create option cannot convert to secret")
	}
	c := &S3{}

	var provider = &SecretProvider{
		Secret: secret,
	}

	rg, err := provider.Region()
	if err != nil {
		return nil, err
	}

	var credConfig = config.WithCredentialsProvider(provider)

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithDefaultRegion(rg), credConfig)
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := awss3.NewFromConfig(cfg)
	c.client = client

	return c, nil
}

func (s *S3) CreateBucket(ctx context.Context, name string) error {
	input := &awss3.CreateBucketInput{
		ACL: types.BucketCannedACLPrivate,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(s.cfg.Region),
		},
		Bucket: &name,
	}

	_, err := s.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("create bucket %s failed, %w", name, err)
	}
	return nil
}

func (s *S3) DeleteBucket(ctx context.Context, name string) error {
	input := &awss3.DeleteBucketInput{
		Bucket: &name,
	}
	_, err := s.client.DeleteBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("delete bucket %s failed, %w", name, err)
	}
	return nil
}
