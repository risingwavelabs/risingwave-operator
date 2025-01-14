# Using IAM for Service Account to Configure the S3 State Store Backend

## Overview

This document describes how to use IAM for Service Account to configure the S3 state store backend.

## Prerequisites

1. The RisingWave should be running on the AWS EKS cluster.

## Prepare

1. Follow the instructions [here](https://docs.aws.amazon.com/AmazonS3/latest/userguide/creating-bucket.html) to create
   a S3 bucket.
2. Follow the instructions [here](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
   to create an IAM role for service account. These two steps must be completed:
    1. [Creating an IAM OIDC provider for your cluster](https://docs.aws.amazon.com/eks/latest/userguide/enable-iam-roles-for-service-accounts.html)
    2. [Configuring a Kubernetes service account to assume an IAM role](https://docs.aws.amazon.com/eks/latest/userguide/associate-service-account-role.html)
3. Attach the IAM role
   with [AmazonS3FullAccess](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/AmazonS3FullAccess.html).
    1. You can customize the IAM policy to limit the access to the S3 bucket. Here
       is [a guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_resource.html) to limit
       the resources in policy.

## Configure

Assume that

1. The service account name is `risingwave-s3`.
2. The S3 bucket name is `risingwave`.
3. The AWS region is `us-east-1`.

Use the following YAML to configure the S3 state store backend:

> Note that meta store is commented out. If you want to use etcd as the meta store, please uncomment the meta store and
> make sure it's configured correctly.

```yaml
apiVersion: risingwave.risingwavelabs.com/v1alpha1
kind: RisingWave
metadata:
  name: risingwave
spec:
  image: risingwavelabs/risingwave:v2.1.2
  #  metaStore:
  #    etcd:
  #      endpoint: etcd:2388
  stateStore:
    dataDirectory: hummock001
    s3:
      bucket: risingwave
      region: us-east-1
      credentials:
        useServiceAccount: true   # Use IAM for Service Account
  components:
    meta:
      nodeGroups:
      - replicas: 1
        name: ''
        template:
          spec:
            serviceAccountName: risingwave-s3  # Use the service account
    compactor:
      nodeGroups:
      - replicas: 1
        name: ''
        template:
          spec:
            serviceAccountName: risingwave-s3 # Use the service account
    frontend:
      nodeGroups:
      - replicas: 1
        name: ''
        template:
          spec:
            serviceAccountName: risingwave-s3 # Use the service account
    compute:
      nodeGroups:
      - replicas: 1
        name: ''
        template:
          spec:
            serviceAccountName: risingwave-s3 # Use the service account
```