# RisingWave Operator

[![Slack](https://badgen.net/badge/Slack/Join%20RisingWave/0abd59?icon=slack)](https://join.slack.com/t/risingwave-community/shared_invite/zt-120rft0mr-d8uGk3d~NZiZAQWPnElOfw)
![Build](https://github.com/risingwavelabs/risingwave-operator/actions/workflows/e2e.yaml/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/risingwavelabs/risingwave-operator/branch/main/graph/badge.svg?token=D08wi9hnt4)](https://codecov.io/gh/risingwavelabs/risingwave-operator)

## Introduction

The RisingWave operator is a deployment and management system of
the [RisingWave streaming database](https://github.com/risingwavelabs/risingwave) that runs on top of Kubernetes. It
provides functionalities like provisioning, upgrading, scaling and destroying the `RisingWave` instances inside the
Kubernetes cluster. It models the deployment and management progress with the concepts provided in Kubernetes and
organizes them in a way called [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).
Thus, we can just declare what kind of `RisingWave` instances we want and create them as objects in the Kubernetes. The
RisingWave operator will always make sure that they are finally there.

The operator also contains several custom resources. Refer to the [API docs](./docs/general/api.md) for more details.

## Quick Start

### Installation

First, you need to install the `cert-manager` in the cluster before installing the `risingwave-operator`.

The default static configuration cert-manager can be installed as follows:

```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.yaml
```

More information on the installation of
the `cert-manager` [can be found here](https://cert-manager.io/docs/installation/#default-static-install).

Afterward, you can install the `risingwave-operator` with the following command:

```shell
kubectl apply --server-side -f https://github.com/risingwavelabs/risingwave-operator/releases/latest/download/risingwave-operator.yaml
```

To check if the installation is successful, you can run the following commands to check if the Pods are all running.

```shell
kubectl -n cert-manager get pods
kubectl -n risingwave-operator-system get pods
```

> NOTE: Don't worry if you encountered the following errors:
>
> ```text
> Error from server (InternalError): error when creating "https://github.com/risingwavelabs/risingwave-operator/releases/latest/download/risingwave-operator.yaml": Internal error occurred: failed calling webhook "webhook.cert-manager.io": Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp 10.111.35.75:443: connect: connection refused
> Error from server (InternalError): error when creating "https://github.com/risingwavelabs/risingwave-operator/releases/latest/download/risingwave-operator.yaml": Internal error occurred: failed calling webhook "webhook.cert-manager.io": Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp 10.111.35.75:443: connect: connection refused
> ```
>
> The errors are because of the initializing `cert-manager`. Just hold tight and wait for another minute, and re-apply
> the `risingwave-operator.yaml` above.

### First RisingWave Instance

Now you can deploy a RisingWave instance with in-memory storage with the following command (be careful about the node
arch):

```shell
kubectl apply -f https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/docs/manifests/risingwave/risingwave-in-memory.yaml
```

Check the running status of RisingWave with the following command:

```shell
kubectl get risingwave
```

The expected output is like this:

```plain
NAME                    RUNNING   STORAGE(META)   STORAGE(OBJECT)   AGE
risingwave-in-memory    True      Memory          Memory            30s
```

> If you find the `risingwave-in-memory` `RUNNING` filed is `false`, please type `kubectl get pods`. If you find some pods status is `ImagePullBackOff` rather than `Running`, the problem may result from your network.


### Connect & Query

#### ClusterIP

By default, the operator will create a service for the frontend component, with the type of `ClusterIP` if not
specified. It is not accessible from the outside. So we will create a standalone Pod of PostgreSQL inside the
Kubernetes, which runs an infinite loop so that we can attach to it.

You can create one by following the commands below, or you can just do it yourself:

```shell
kubectl apply -f docs/manifests/psql/psql-console.yaml
```

And then you will find a Pod named `psql-console` running in the Kubernetes, and you can attach to it to execute
commands inside the container with the following command:

```shell
kubectl exec -it psql-console bash
```

Finally, we can get access to the RisingWave with the `psql` command inside the Pod:

```shell
psql -h risingwave-in-memory-frontend -p 4567 -d dev -U root
```

#### NodePort

If you want to connect to the RisingWave from the nodes (e.g., EC2) in the Kubernetes, you can set the service type
to `NodePort`, and run the following commands on the node:

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

For EKS/GKE and some other Kubernetes services provided by cloud vendors, we can expose the Service to the public network with a
load balancer on the cloud. We can simply achieve this by setting the service type to `LoadBalancer`, by setting the
following field:

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
For launching RisingWaves with different storage configuration running in Kubernetes, please refer to the [Storage Guidance](/docs/general/storage.md) for more details.


## Monitoring

For monitoring the RisingWaves running in Kubernetes, please refer to the [Monitoring Guidance](./monitoring/README.md)
for more details.

## License

The RisingWave operator is developed under the Apache License 2.0. Please refer to [LICENSE](LICENSE) for more
information.

## Contributing

Thanks for your interest in contributing to the project! Please refer to
the [Contribution and Development Guidelines](CONTRIBUTING.md) for more information.
