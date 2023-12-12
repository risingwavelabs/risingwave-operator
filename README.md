# RisingWave Kubernetes Operator

[![Slack](https://badgen.net/badge/Slack/Join%20RisingWave/0abd59?icon=slack)](https://risingwave-community.slack.com/join/shared_invite/zt-1afreobhd-5Npy1oIpUWvDA~Od6zPxTA#/shared-invite/email)
[![Build status](https://badge.buildkite.com/db2c3f749ff1696b9ca22b23990f144133a1f74685e3285ad4.svg)](https://buildkite.com/risingwave-operator/main)
[![codecov](https://codecov.io/gh/risingwavelabs/risingwave-operator/branch/main/graph/badge.svg?token=D08wi9hnt4)](https://codecov.io/gh/risingwavelabs/risingwave-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Description

The RisingWave Kubernetes Operator is a powerful tool designed to facilitate the management and deployment
of [RisingWave, a streaming processing platform written in Rust](https://github.com/risingwavelabs/risingwave). With its
distributed architecture, RisingWave provides a scalable and efficient solution for processing large streams of data in
real-time.

The Kubernetes operator acts as a bridge between the RisingWave platform and the Kubernetes cluster, streamlining the
deployment and management process. It leverages the native capabilities of Kubernetes to automate tasks such as scaling,
monitoring, and fault tolerance, making it easier to operate RisingWave in a Kubernetes environment.

## Table of Contents

- [Description](#description)
- [Compatibility](#compatibility)
- [Installation](#installation)
  - [Install RisingWave Operator](#install-risingwave-operator)
- [Usage](#usage)
  - [Create a RisingWave Cluster](#create-a-risingwave-cluster)
  - [Connect to the RisingWave Cluster](#connect-to-the-risingwave-cluster)
  - [Delete the RisingWave Cluster](#delete-the-risingwave-cluster)
  - [Customize the RisingWave Cluster](#customize-the-risingwave-cluster)
- [Contribution Guidelines](#contribution-guidelines)
- [License](#license)

## Compatibility

RisingWave Operator has been tested and should be working with the following Kubernetes distributions:

- [AWS EKS](https://aws.amazon.com/eks/)
- [GCP GKE](https://cloud.google.com/kubernetes-engine)
- [Azure AKS](https://azure.microsoft.com/en-us/services/kubernetes-service/)
- [Aliyun ACK](https://www.aliyun.com/product/kubernetes)
- [Docker Kubernetes](https://www.docker.com/products/kubernetes)
- [kind](https://kind.sigs.k8s.io/)
- [minukube](https://minikube.sigs.k8s.io/)

If you are using other Kubernetes distributions or encounter problems, please feel free
to [create an issue](https://github.com/risingwavelabs/risingwave-operator/issues/new).

Here is the compatibility matrix:

| RisingWave Operator | RisingWave | Kubernetes |
|---------------------|------------|------------|
| main                | v0.19.0+   | v1.21+     |
| v0.5.0+             | v0.19.0+   | v1.21+     |
| v0.4.1              | v0.18.0+   | v1.21+     |
| v0.3.6              | v0.18.0+   | v1.21+     |


## Installation

To secure the webhook server, you need to install the `cert-manager` first. Please refer to
the [cert-manager installation guide](https://cert-manager.io/docs/installation/kubectl/) for more information.

### Install RisingWave Operator

Install the latest version of RisingWave Operator:

```shell
kubectl apply --server-side -f https://github.com/risingwavelabs/risingwave-operator/releases/latest/download/risingwave-operator.yaml
```

(Optional) Install RisingWave Operator with a specific version:

```shell
# Replace ${VERSION} with the version you want to install, e.g., v0.4.0
kubectl apply --server-side -f https://github.com/risingwavelabs/risingwave-operator/releases/download/${VERSION}/risingwave-operator.yaml
```

(Optional) Install the main branch of RisingWave Operator (not recommended for production environments):

```shell
kubectl apply --server-side -f https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/config/risingwave-operator.yaml
```

> Note: errors might occur if cert-manager is not fully initialized. Don't panic! Simply wait for another minute and
> retry the command above.
>
> > Error from server (InternalError): Internal error occurred: failed calling webhook "webhook.cert-manager.io": failed
> > to call webhook: Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp
> > 10.105.102.32:
> > 443: connect: connection refused
> >
> > Error from server (InternalError): Internal error occurred: failed calling webhook "webhook.cert-manager.io": failed
> > to call webhook: Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp
> > 10.105.102.32:
> > 443: connect: connection refused

Check the installation status:

```shell
# Check the CRDs
$ kubectl get crds | grep risingwavelabs.com
risingwaves.risingwave.risingwavelabs.com              2023-05-23T06:04:00Z
risingwavescaleviews.risingwave.risingwavelabs.com     2023-05-23T06:04:01Z

# Check the controller Pod status
$ kubectl -n risingwave-operator-system get pods
NAME                                                     READY   STATUS    RESTARTS   AGE
risingwave-operator-controller-manager-b5d5f585d-6npn5   2/2     Running   0          60s
```

### Helm chart

You can also use [Helm chart](https://github.com/risingwavelabs/helm-charts/tree/main/charts/risingwave-operator) to install the operator.

## Usage

RisingWave Kubernetes Operator extends the Kubernetes with CRDs (Custom Resource Definitions) to manage RisingWave. That
means all you need to do is to create a RisingWave resource in your Kubernetes cluster, and the RisingWave Kubernetes
Operator will take care of the rest.

The RisingWave resource is a custom resource that defines a RisingWave cluster. You can find more examples in
the [docs/manifests/risingwave](docs/manifests/risingwave) directory. For more details of the APIs, please refer to
the [API reference](docs/general/api.md).

> NOTE: since the project is still under rapid development, the compatibility between different versions of RisingWave
> operator might be broken. We have maintained a stable set of manifests in
> the [docs/manifest/stable](docs/manifests/stable) directory that are ensured to be compatible with the latest released
> version. Please use them if you want to deploy RisingWave in a production environment.

### Create a RisingWave cluster

Follow the steps below to create a RisingWave cluster in your Kubernetes cluster:

```shell
# Download the manifest YAML file.
curl https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/docs/manifests/stable/persistent/minio/risingwave.yaml -o risingwave.yaml

# Apply it to the Kubernetes cluster.
kubectl apply -f risingwave.yaml
```

> Note: the RisingWave cluster will be created in the `default` namespace by default. If you want to create it in
> another namespace, please modify the `metadata.namespace` field in the manifest YAML file or use the `--namespace`
> option.

The RisingWave cluster will be created in a few minutes. You can check the status of the RisingWave cluster by running
the following command:

```shell
kubectl get risingwave
NAME         META STORE   STATE STORE   VERSION   RUNNING   AGE
risingwave   Etcd         MinIO         v1.5.0    True      2m20s
```

> Note: the `META STORE` column indicates the storage backend for the RisingWave metadata. The `STATE STORE` column
> indicates the storage backend for the state store. The `VERSION` column indicates the version of the RisingWave
> cluster.
> The `RUNNING` column indicates whether the RisingWave cluster is running.

You can check the Pods of the RisingWave cluster by running the following command:

```shell
kubectl get pods -l risingwave/name
NAME                                    READY   STATUS    RESTARTS      AGE
risingwave-compactor-5cfcb469c5-gnkrp   1/1     Running   2 (1m ago)    2m35s
risingwave-compute-0                    1/1     Running   2 (1m ago)    2m35s
risingwave-frontend-86c948f4bb-69cld    1/1     Running   2 (1m ago)    2m35s
risingwave-meta-0                       1/1     Running   1 (1m ago)    2m35s
```

### Connect to the RisingWave cluster

The RisingWave cluster is now ready to use. However, it is not accessible from outside the Kubernetes cluster by
default. To connect to the RisingWave cluster, you need to forward the ports of the RisingWave cluster to your local
machine:

```shell
kubectl port-forward svc/risingwave-frontend 4567:service
```

Keep the port forwarding command running in the terminal and open another terminal window. You can now connect to the
RisingWave cluster using the `psql` command line tool. The default username is `root` and the default database name
is `dev`:

```shell
psql -h localhost -p 4567 -d dev -U root
```

Now try to create a table in the database:

```sql
dev=> CREATE TABLE t1 (v1 int);
CREATE_TABLE
```

Then create a materialized view based on the table:

```sql
dev=> CREATE MATERIALIZED VIEW mv1 AS SELECT sum(v1) AS sum_v1 FROM t1;
CREATE_MATERIALIZED_VIEW
```

Insert some data into the table:

```sql
dev=> INSERT INTO t1 VALUES (1), (2), (3);
INSERT 0 3

dev=> FLUSH;
FLUSH
```

Now you can query the materialized view:

```sql
dev=> SELECT * FROM mv1;
sum_v1
--------
      6
(1 row)

```

### Delete the RisingWave cluster

To delete the RisingWave cluster, simply delete the RisingWave resource:

```shell
kubectl delete risingwave risingwave
```

The Pods will be deleted in a few minutes.

> Note: the data in the RisingWave cluster will not be lost after the RisingWave cluster is deleted in this example
> since the etcd and MinIO services are still running. If you would like to terminate them all and purge the data, you
> can run the following commands:
> ```shell
> kubectl delete -f risingwave.yaml   # Delete all resources defined in the risingwave.yaml that you used above.
> kubectl delete pvc -l app=etcd      # Delete the PVCs of etcd.
> kubectl delete pvc -l app=minio     # Delete the PVCs of MinIO.
> ```

### Customize the RisingWave cluster

You can customize the RisingWave cluster by modifying the manifest YAML file. For more details, please refer to the API
reference in the [docs/general/api.md](docs/general/api.md) file.

For customizing the state store backends of RisingWave
cluster, please refer to the [docs/general/state-stores.md](docs/general/state-stores.md) file.

## Contribution Guidelines

We welcome contributions from the community! If you would like to contribute to this project, please follow the
guidelines outlined in the [CONTRIBUTING.md](CONTRIBUTING.md) file.

## License

This project is licensed under the [Apache License 2.0](LICENSE). You can find the full text of the license in
the [LICENSE](LICENSE) file.
