local grafana = import 'grafonnet/grafana.libsonnet';
local row = grafana.row;
local timeseriesPanel = grafana.timeseriesPanel;
local prometheus = grafana.prometheus;

local span = import 'mixon/span.libsonnet';
local strings = import 'lib/strings.libsonnet';

local panels = {
  compaction_success_failure_counts_overview::
    local expr = |||
      sum(storage_level_compact_frequency{namespace="$namespace", risingwave_name="$instance"}) by (result)
    |||;

    timeseriesPanel.new(
      title='Compaction Success & Failure Counts (Overview)',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ result }}',
      )
    ) + span.span('half'),

  compaction_success_failure_counts_detailed::
    local expr = |||
      sum(storage_level_compact_frequency{namespace="$namespace", risingwave_name="$instance"}) by (pod, group, result)
    |||;

    timeseriesPanel.new(
      title='Compaction Success & Failure Counts (Detailed)',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }} - group{{ group }} - {{ result }}',
      )
    ) + span.span('half'),

  max_compaction_task_duration_overview::
    local expr = |||
      histogram_quantile(1, sum by (le) (rate(state_store_compact_task_duration_bucket{namespace="$namespace", risingwave_name="$instance"}[$__rate_interval])))
    |||;

    timeseriesPanel.new(
      title='Longest Compaction Duration (Overview)',
      unit='s',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='overview',
      )
    ) + span.span('half'),

  max_compaction_task_duration_detailed::
    local expr = |||
      histogram_quantile(1, sum by (pod, le) (rate(state_store_compact_task_duration_bucket{namespace="$namespace", risingwave_name="$instance"}[$__rate_interval])))
    |||;

    timeseriesPanel.new(
      title='Longest Compaction Duration (Detailed)',
      unit='s',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }}',
      )
    ) + span.span('half'),

  median_compaction_task_duration_overview::
    local expr = |||
      histogram_quantile(0.5, sum by (le) (rate(state_store_compact_task_duration_bucket{namespace="$namespace", risingwave_name="$instance"}[$__rate_interval])))
    |||;

    timeseriesPanel.new(
      title='Median Compaction Duration (Overview)',
      unit='s',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='overview',
      )
    ) + span.span('half'),

  median_compaction_task_duration_detailed::
    local expr = |||
      histogram_quantile(0.5, sum by (pod, le) (rate(state_store_compact_task_duration_bucket{namespace="$namespace", risingwave_name="$instance"}[$__rate_interval])))
    |||;

    timeseriesPanel.new(
      title='Median Compaction Duration (Detailed)',
      unit='s',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }}',
      )
    ) + span.span('half'),

  compaction_read_write_rate::
    local expr1 = |||
      sum(rate(storage_level_compact_write{namespace="$namespace", risingwave_name="$instance"}[$__rate_interval])) by(pod)
    |||;
    local expr2 = |||
      sum by(pod) (
        rate(storage_level_compact_read_next{namespace="$namespace", risingwave_name="$instance"}[$__rate_interval])
      ) 
      + 
      sum by(pod) (
        rate(storage_level_compact_read_curr{namespace="$namespace", risingwave_name="$instance"}[$__rate_interval])
      )
    |||;

    timeseriesPanel.new(
      title='Compaction Rate of Reading & Writing',
      unit='Bps',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr1),
        legendFormat='Rate of Writes ({{ pod }})',
      )
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr2),
        legendFormat='Rate of Reads ({{ pod }})',
      )
    ) + span.span('half'),

  sorted_string_table_file_size::
    local expr = |||
      sum(storage_level_total_file_size{namespace="$namespace", risingwave_name="$instance"}) by (pod, level_index)
    |||;

    timeseriesPanel.new(
      title='Sorted String Table File Sizes',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ pod }} - {{ level_index }}',
      )
    ) + span.span('half'),
};

row.new(
  title='Compactor',
  height='300px',
  collapse=true,
).addPanels(
  [
    panels.compaction_success_failure_counts_overview,
    panels.compaction_success_failure_counts_detailed,

    panels.max_compaction_task_duration_overview,
    panels.max_compaction_task_duration_detailed,

    panels.median_compaction_task_duration_overview,
    panels.median_compaction_task_duration_detailed,

    panels.compaction_read_write_rate,
    panels.sorted_string_table_file_size,
  ]
)
