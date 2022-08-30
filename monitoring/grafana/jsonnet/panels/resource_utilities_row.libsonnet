local grafana = import 'grafonnet/grafana.libsonnet';
local row = grafana.row;
local timeseriesPanel = grafana.timeseriesPanel;
local prometheus = grafana.prometheus;

local span = import 'mixon/span.libsonnet';
local strings = import 'lib/strings.libsonnet';

local panels = {
  component_level_avg_cpu_utilities::
    local expr = |||
      avg by (risingwave_component) (
          (
              rate(process_cpu_seconds_total{namespace="$namespace"}[$__rate_interval])
            * on (namespace, pod) group_left (node)
              topk by (namespace, pod) (1, max by (namespace, pod, node) (kube_pod_info{node!=""}))
          )
        * on (namespace, pod) group_left ()
          topk by (namespace, pod) (
            1,
            max by (namespace, risingwave_component, pod) (up{risingwave_name="$instance"})
          )
      )
    |||;

    timeseriesPanel.new(
      title='Avg CPU (Component-Level)',
      unit='percentunit',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

  component_level_avg_mem_utilities::
    local expr = |||
      avg by (risingwave_component) (
          (
              process_resident_memory_bytes{namespace="$namespace"}
            * on (namespace, pod) group_left (node)
              topk by (namespace, pod) (1, max by (namespace, pod, node) (kube_pod_info{node!=""}))
          )
        * on (namespace, pod) group_left ()
          topk by (namespace, pod) (
            1,
            max by (namespace, risingwave_component, pod) (up{risingwave_name=~"$instance"})
          )
      )
    |||;

    timeseriesPanel.new(
      title='Avg Memory (Component-Level)',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

  component_level_max_cpu_utilities::
    local expr = |||
      max by (risingwave_component) (
          (
              rate(process_cpu_seconds_total{namespace="$namespace"}[$__rate_interval])
            * on (namespace, pod) group_left (node)
              topk by (namespace, pod) (1, max by (namespace, pod, node) (kube_pod_info{node!=""}))
          )
        * on (namespace, pod) group_left ()
          topk by (namespace, pod) (
            1,
            max by (namespace, risingwave_component, pod) (up{risingwave_name="$instance"})
          )
      )
    |||;

    timeseriesPanel.new(
      title='Max CPU (Component-Level)',
      unit='percentunit',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

  component_level_max_mem_utilities::
    local expr = |||
      max by (risingwave_component) (
          (
              process_resident_memory_bytes{namespace="$namespace"}
            * on (namespace, pod) group_left (node)
              topk by (namespace, pod) (1, max by (namespace, pod, node) (kube_pod_info{node!=""}))
          )
        * on (namespace, pod) group_left ()
          topk by (namespace, pod) (
            1,
            max by (namespace, risingwave_component, pod) (up{risingwave_name=~"$instance"})
          )
      )
    |||;

    timeseriesPanel.new(
      title='Max Memory (Component-Level)',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

  pod_level_cpu_utilities::
    local expr = |||
      sum by (pod, risingwave_component, node) (
          (
              rate(process_cpu_seconds_total{namespace="$namespace"}[$__rate_interval])
            * on (namespace, pod) group_left (node)
              topk by (namespace, pod) (1, max by (namespace, pod, node) (kube_pod_info{node!=""}))
          )
        * on (namespace, pod) group_left ()
          topk by (namespace, pod) (
            1,
            max by (namespace, risingwave_component, pod) (up{risingwave_name="$instance"})
          )
      )
    |||;
    timeseriesPanel.new(
      title='CPU (Pod-Level)',
      unit='percentunit',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }} ({{ risingwave_component }}) @ {{ node }}',
      )
    ) + span.span('half'),

  pod_level_mem_utilities::
    local expr = |||
      avg by (pod, risingwave_component, node) (
          (
              process_resident_memory_bytes{namespace="$namespace"}
            * on (namespace, pod) group_left (node)
              topk by (namespace, pod) (1, max by (namespace, pod, node) (kube_pod_info{node!=""}))
          )
        * on (namespace, pod) group_left ()
          topk by (namespace, pod) (
            1,
            max by (namespace, risingwave_component, pod) (up{risingwave_name=~"$instance"})
          )
      )
    |||;

    timeseriesPanel.new(
      title='Memory (Pod-Level)',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }} ({{ risingwave_component }}) @ {{ node }}',
      )
    ) + span.span('half'),
};

row.new(
  title='Resource Utilities',
  height='300px',
  collapse=true,
).addPanels(
  [
    panels.component_level_avg_cpu_utilities,
    panels.component_level_avg_mem_utilities,

    panels.component_level_max_cpu_utilities,
    panels.component_level_max_mem_utilities,

    panels.pod_level_cpu_utilities,
    panels.pod_level_mem_utilities,
  ]
)
