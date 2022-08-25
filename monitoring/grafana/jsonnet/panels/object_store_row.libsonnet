local grafana = import 'grafonnet/grafana.libsonnet';
local row = grafana.row;
local timeseriesPanel = grafana.timeseriesPanel;
local prometheus = grafana.prometheus;

local span = import 'mixon/span.libsonnet';
local strings = import 'lib/strings.libsonnet';

local panels = {
  pod_level_object_store_read_throughput::
    local expr = |||
      sum by (risingwave_component, pod) (
        rate(object_store_read_bytes{namespace="$namespace",risingwave_name="$instance"}[$__rate_interval])
      )
    |||;

    timeseriesPanel.new(
      title='Read Throughput (Pod-Level)',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }} ({{ risingwave_component }})',
      )
    ) + span.span('half'),

  pod_level_object_store_write_throughput::
    local expr = |||
      sum by (risingwave_component, pod) (
        rate(object_store_write_bytes{namespace="$namespace",risingwave_name="$instance"}[$__rate_interval])
      )
    |||;

    timeseriesPanel.new(
      title='Write Throughput (Pod-Level)',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }} ({{ risingwave_component }})',
      )
    ) + span.span('half'),

  component_level_object_store_read_throughput::
    local expr = |||
      sum by (risingwave_component) (
        rate(object_store_read_bytes{namespace="$namespace",risingwave_name="$instance"}[$__rate_interval])
      )
    |||;

    timeseriesPanel.new(
      title='Read Throughput (Component-Level)',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

  component_level_object_store_write_throughput::
    local expr = |||
      sum by (risingwave_component) (
        rate(object_store_write_bytes{namespace="$namespace",risingwave_name="$instance"}[$__rate_interval])
      )
    |||;

    timeseriesPanel.new(
      title='Write Throughput (Component-Level)',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),
};

row.new(
  title='Object Store',
  height='300px',
  collapse=true,
).addPanels(
  [
    panels.component_level_object_store_read_throughput,
    panels.component_level_object_store_write_throughput,

    panels.pod_level_object_store_read_throughput,
    panels.pod_level_object_store_write_throughput,
  ]
)
