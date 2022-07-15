# RisingWave Operator

## Introduction

The RisingWave operator is a deployment and management system of the [RisingWave streaming database](https://github.com/singularity-data/risingwave) that runs on top of Kubernetes. It provides functionalities like provisioning, upgrading, scaling and destroying the `RisingWave` instances inside the Kubernetes cluster. It models the deployment and management progress with the concepts provided in Kubernetes and organizes them in a way called [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). Thus we can just declare what kind of `RisingWave` instances we want and create them as objects in the Kubernetes. The RisingWave operator will always make sure that they are finally there.

The operator also contains several custom resources, as listed below:

- `risingwave.singularity-data.com/v1alpha1`
  - `RisingWave`
  - `RisingWavePodTemplate`

## Quick Start

### Installation

First, you need to install the `cert-manager` in the cluster before installing the `risingwave-operator`.

The default static configuration cert-manager can be installed as follows:

```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
```

More information on this install cert-manager method [can be found here](https://cert-manager.io/docs/installation/#default-static-install).

Then, you can install the `risingwave-operator` with the following command:

```shell
kubectl apply -f https://raw.githubusercontent.com/singularity-data/risingwave-operator/main/config/risingwave-operator.yaml
```

To check if the installation is successful, you can run the following commands to check if the Pods are running.

```shell
kubectl -n cert-manager get pods
kubectl -n risingwave-operator-system get pods
```

### First RisingWave Instance

Now you can deploy a RisingWave instance with in-memory storage with the following command (be careful about the node arch):

```shell
# It runs on the Linux/amd64 platform. If you want to run on Linux/arm64, you need to run the command below.
kubectl apply -f https://raw.githubusercontent.com/singularity-data/risingwave-operator/main/examples/risingwave-in-memory.yaml

# Linux/arm64
curl https://raw.githubusercontent.com/singularity-data/risingwave-operator/main/examples/risingwave-in-memory.yaml | sed -e 's/ghcr.io\/singularity-data\/risingwave/public.ecr.aws\/x5u3w5h6\/risingwave-arm/g' | kubectl apply -f -
```

Check the running status of RisingWave with the following command:

```shell
kubectl get risingwave
```

The expected output is like this:

```shell
NAME                    RUNNING   STORAGE(META)   STORAGE(OBJECT)   AGE
risingwave-in-memory    True      Memory          Memory            30s
```

### Connect & Query

#### ClusterIP

By default, the operator will create a service for the frontend component, with the type of `ClusterIP` if not specified. It is not accessible from the outside. So we will create a standalone Pod of PostgreSQL inside the Kubernetes, which runs an infinite loop so that we can attach to it.

You can create one by following the commands below, or you can just do it yourself:

```shell
kubectl apply -f examples/psql/psql-console.yaml
```

And then you will find a Pod named `psql-console` running in the Kubernetes, and you can attach to it to execute commands inside the container with the following command:

```shell
kubectl exec -it psql-console bash
```

Finally, we can get access to the RisingWave with the `psql` command inside the Pod:

```shell
psql -h risingwave-in-memory -p 4567 -d dev -U root
```

#### NodePort

If you want to connect to the RisingWave from the nodes (e.g., EC2) in the Kubernetes, you can set the service type to `NodePort`, and run the following commands on the node:

```shell
export RISINGWAVE_NAME=risingwave-in-memory
export RISINGWAVE_NAMESPACE=default
export RISINGWAVE_HOST=`kubectl -n ${RISINGWAVE_NAMESPACE} get node -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'`
export RISINGWAVE_PORT=`kubectl -n ${RISINGWAVE_NAMESPACE} get svc -l risingwave/name=${RISINGWAVE_NAME},risingwave/component=frontend -o jsonpath='{.items[0].spec.ports[0].nodePort}'`

psql -h ${RISINGWAVE_HOST} -p ${RISINGWAVE_PORT} -d dev -U root
```

```yamlex
# ...
spec:
  global:
    serviceType: NodePort
# ...
```

#### LoadBalancer

For EKS/GCP and some other Kubernetes provided by cloud vendors, we can expose the Service to the public network with a load balancer on the cloud. We can simply achieve this by setting the service type to `LoadBalancer`, by setting the following field:

```yamlex
# ...
spec:
  global:
    serviceType: LoadBalancer
# ...
```

And then you can connect to the RisingWave with the following command:

```shell
export RISINGWAVE_NAME=risingwave-in-memory
export RISINGWAVE_NAMESPACE=default
export RISINGWAVE_HOST=`kubectl -n ${RISINGWAVE_NAMESPACE} get svc -l risingwave/name=${RISINGWAVE_NAME},risingwave/component=frontend -o jsonpath='{.items[0].status.loadBalancer.ingress[0].ip}'`
export RISINGWAVE_PORT=`kubectl -n ${RISINGWAVE_NAMESPACE} get svc -l risingwave/name=${RISINGWAVE_NAME},risingwave/component=frontend -o jsonpath='{.items[0].spec.ports[0].port}'`

psql -h ${RISINGWAVE_HOST} -p ${RISINGWAVE_PORT} -d dev -U root
```

## Storages

### Memory

Currently, memory storage is supported for test usage only. We highly discourage you use the memory storage for other purposes. For now, you can enable the memory metadata and object storage with the following configs:

```yamlex
#...
spec:
  storages:
    meta:
      memory: true
    object:
      memory: true
#...
```

### etcd (meta)

We recommend using the [etcd](https://etcd.io/)](https://etcd.io/) to store the metadata. You can specify the connection information of the `etcd` you'd like to use like the following:

```yamlex
#...
spec:
  storages:
    meta:
      etcd: 
        endpoint: risingwave-etcd:2388
        secret: etcd-credentials      # optional, empty means no authentication 
#...
```

Check the [examples/risingwave-etcd-minio.yaml](./examples/risingwave-etcd-minio.yaml) for how to provision a simple RisingWave with an `etcd` instance as the metadata storage.

### MinIO

We support using MinIO as the object storage. Check the [examples/risingwave-etcd-minio.yaml](./examples/risingwave-etcd-minio.yaml) for details. The YAML structure is like the following:

```yamlex
#...
spec:
  storages:
    object:
      minio:
        secret: minio-credentials
        endpoint: minio-endpoint:2388
        bucket: hummock001
#...
```

### S3

We support using AWS S3 as the object storage. Follow the steps below and check the [examples/risingwave-etcd-s3.yaml](./examples/risingwave-etcd-s3.yaml) for details:

First, you need to create a `Secret` with the name `s3-credentials`:

```shell
kubectl create secret generic s3-credentials --from-literal AccessKeyID=${ACCESS_KEY} --from-literal SecretAccessKey=${SECRET_ACCESS_KEY} --from-literal Region=${AWS_REGION}
```

Then, you need to create a `bucket` on the console, e.g., `hummock001`.

Finally, you can specify S3 as the object storage in YAML, like the following:

```yamlex
#...
spec:
  storages:
    object:
      s3:
        secret: s3-credentials
        bucket: hummock001
#...
```

## License

The RisingWave operator is developed under the Apache License 2.0. Please refer to [LICENSE](LICENSE) for more information.

## Contributing

Thanks for your interest in contributing to the project! Please refer to the [Contribution and Development Guidelines](CONTRIBUTING.md) for more information.
