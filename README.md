## Introduction

The RisingWave Kubernetes Operator is a RisingWave deployment management tool based on Kubernetes. The `risingwave-operator` provides the following custom resources:
- risingwave.singularity-data.com

## Quick Start

#### Install cert-manager

We need to install the `cert-manager` in the cluster before installing the `risingwave-operator`.

The default static configuration cert-manager can be installed as follows:

```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml
```

More information on this install cert-manager method [can be found here](https://cert-manager.io/docs/installation/#default-static-install).


#### Install risingwave-operator

`risingwave-operator` can be installed as follows:

```shell
kubectl apply -f https://raw.githubusercontent.com/singularity-data/risingwave-operator/main/config/risingwave-operator.yaml
```

## Examples

You can deploy RisingWave which uses MinIO on Linux/amd64 arch nodes as follows:

```shell
kubectl create namespace test
kubectl apply -f https://raw.githubusercontent.com/singularity-data/risingwave-operator/main/examples/minio-risingwave-amd.yaml
```

## First Query

#### Install psql

To connect to the RisingWave server, you will need to [install PostgreSQL shell](./CONTRIBUTING.md#PostgreSQL) (`psql`) in advance.


#### Query

We use `NodePort` service for the frontend.

Please get the nodePort of the frontend service as `psql port` and get the `INTERNAL-IP` address of any node as follows:

```shell
PHOST=`kubectl get node -o=jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'`
```

```shell
PPORT=`kubectl get service -n test test-risingwave-amd64-frontend -o=jsonpath='{.spec.ports[0].nodePort}'`
```

Connect to the frontend by `psql` as follows:

```shell
psql -h $PHOST -p $PPORT -d dev
```

## Storage

### S3

We support using AWS S3 as storage. You can use S3 as follows:

1. Create a `secret` named `cloud-provider-configure` in your namespace.
```shell
kubectl create secret generic -n test cloud-provider-configure --from-literal AccessKeyID=XXXXXXX --from-literal SecretAccessKey=YYYYYYY --from-literal Region=ZZZZZ
```

2. Create a `bucket` in the AWS Console.
3. Use the bucket by setting the following fields: 
```yamlex
objectStorage:
  s3:
    provider: aws
    bucket: xxxx #your-bucket-name
```


## Configuration

You can get the `risingwave-operator` configuration as follows:

```shell
kubectl get cm risingwave-operator-controller-manager-config -n risingwave-operator-system -o yaml
```

If you edit the `ConfigMap`, please restart the `risingwave-operator` to reload the configuration.

## Monitoring

We recommend using the [prometheus-operator](https://github.com/prometheus-operator/prometheus-operator#quickstart) to install Prometheus.

You can use the command to install `prometheus-operator` as follows:

```shell
kubectl create -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml
```

You can use the command to create a `Prometheus` in your cluster as follows:

```shell
kubectl create -f https://raw.githubusercontent.com/singularity-data/risingwave-operator/main/examples/monitoring/prometheus.yaml
```

If you already deployed the `prometheus-operator` in your cluster, `risingwave-operator` will create `ServiceMonitor` when creating the RisingWave cluster.

## License

The `risingwave-operator` is under the Apache License 2.0. Please refer to [LICENSE](LICENSE) for more information.

## Contributing

Thanks for your interest in contributing to the project! Please refer to the [Contribution and Development Guidelines](CONTRIBUTING.md) for more information.
