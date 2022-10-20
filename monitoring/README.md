# Monitoring Guidance

## Install the `kube-prometheus-stack`

You can install the monitoring stack manually or via the install script.

```shell
./monitoring/install.sh
```

It will create the `monitoring` namespace and deploy everything inside it.

### Prometheus RemoteWrite (AWS)

Prometheus has provided a functionality called [`remote-write`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write),
and AWS provides a [managed Prometheus service](https://aws.amazon.com/prometheus/), so we can write the local metrics to the Prometheus on the cloud.

Before getting started, you need to ensure that you have an account which has the permission to write the managed Prometheus, i.e., with the `AmazonPrometheusRemoteWriteAccess` permission.

Follow the instructions below to set up the remote write:

1. Copy the [prometheus-remote-write-aws.yaml](./kube-prometheus-stack/prometheus-remote-write-aws.yaml) file and replace the values of the these variables:
- `${KUBERNETES_NAME}`: the name of the Kubernetes, e.g., `local-dev`. You can also add `externalLabels` yourself.
- `${AWS_REGION}`: the region of the AWS Prometheus service, e.g., `ap-southeast-1`
- `${WORKSPACE_ID}`: the workspace ID, e.g., `ws-12345678-abcd-1234-abcd-123456789012`

2. Run the install script

```shell
# You can use dry run first with 
# ./monitoring/install.sh -d -r -k <aws_access_key> -s <aws_secret_key>
# See more customization options with 
# ./monitoring/install.sh -h

./monitoring/install.sh -r -k <aws_access_key> -s <aws_secret_key>
```

Now, you can check the Prometheus logs to see if the remote write works, with the following commands:

```shell
kubectl -n monitoring logs prometheus-prometheus-kube-prometheus-prometheus-0
```

The expected output is like this:

```plain
ts=2022-07-20T09:46:38.437Z caller=dedupe.go:112 component=remote level=info remote_name=edcf97 url=https://aps-workspaces.ap-southeast-1.amazonaws.com/workspaces/ws-12345678-abcd-1234-abcd-123456789012/api/v1/remote_write msg="Remote storage resharding" from=2 to=1
```

## Start monitoring

The RisingWave operator has integrated with the [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator). If you have installed the Prometheus Operator in the Kubernetes, it will create a `ServiceMonitor` for the `RisingWave` object and keep it synced automatically. You can check the `ServiceMonitor` with the following command:

```shell
kubectl get servicemonitors -l risingwave/name
```

The expected output is like this:

```plain
NAME                              AGE
risingwave-risingwave-etcd-minio   119m
```

Let's try to forward the web port of Grafana to localhost, with the following command:

```shell
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:http-web
```

Now we can access the Grafana inside the Kubernetes via [http://localhost:3000](http://localhost:3000). By default, the username is `admin` and the password is `prom-operator`.
Let's open the `RisingWave/RisingWave Dashboard` and select the instance you'd like to observe, and here are the panels.

![RisingWave Dashboard](../docs/assets/risingwave-dashboard.png)

## Logging

In addition to the metrics collection and monitoring, we can also integrate the logging stack into the Kubernetes. One of the famous
open source logging stacks is the [Grafana loki](https://grafana.com/docs/loki/latest/). Follow the instructions below to install them in the Kubernetes:

NOTE: this tutorial requires that you have the `helm` and the `kube-prometheus-stack` installed. You can follow the tutorials above to install them.

1. Add the `grafana` repo and update

```shell
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

2. Install the `loki-distributed` chart, including the components of loki

```shell
helm --namespace monitoring --create-namespace upgrade --install loki grafana/loki-distributed
```

3. Install the `promtail` chart, which is an agent that collects the logs and pushes them into the loki

```shell
helm --namespace monitoring --create-namespace upgrade --install promtail grafana/promtail \
  -f https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/monitoring/promtail/loki-promtail-clients.yaml
```

4. Upgrade or install the `kube-prometheus-stack` chart

```shell
helm --namespace monitoring --create-namespace upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  -f https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/monitoring/kube-prometheus-stack/kube-prometheus-stack.yaml \
  -f https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/monitoring/kube-prometheus-stack/grafana-loki-data-source.yaml
```

Now, we are ready to view the logs in Grafana. Just forward the traffics to the localhost, and open the [http://localhost:3000](http://localhost:3000) like mentioned in the
chapter above. Navigate to the `Explore` panel on the left side, and select the loki as data source. Here's an example:

![Grafana Loki](../docs/assets/grafana-loki-example.png)