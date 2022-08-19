local grafana = import 'grafonnet/grafana.libsonnet';
local row = grafana.row;
local timeseriesPanel = grafana.timeseriesPanel;
local prometheus = grafana.prometheus;

local span = import 'mixon/span.libsonnet';

row.new(
  title='Resource Utilities',
  height='300px',
  collapse=false,
).addPanels(
  [
    timeseriesPanel.new(
      title='CPU Utilities (Component-Level)',
      unit='percentunit',
    ).addTarget(
      prometheus.target(
        expr='sum((rate(process_cpu_seconds_total{namespace="$namespace"}[$__rate_interval]) * on(namespace, pod) group_left(node) topk by(namespace, pod) (1, max by(namespace, pod, node) (kube_pod_info{node!=""}))) * on(namespace, pod) group_left() topk by(namespace, pod) (1, max by(namespace, risingwave_component, pod) (up{risingwave_name="$instance"}))) by (pod, risingwave_component, node)',
        legendFormat='{{ pod }} ({{ risingwave_component }}) @ {{ node }}',
      )
    ) + span.span('half'),


  ]
)
