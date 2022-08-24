local grafana = import 'grafonnet/grafana.libsonnet';
local row = grafana.row;
local timeseriesPanel = grafana.timeseriesPanel;
local prometheus = grafana.prometheus;

local span = import 'mixon/span.libsonnet';
local strings = import 'lib/strings.libsonnet';

local panels = {
    compaction_success_failure_counts_overview::
    local expr = |||
      sum(storage_level_compact_frequency) by (result)
    |||;

    timeseriesPanel.new(
      title='Compaction Success & Failure Counts (Overview)',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

    compaction_success_failure_counts_detailed::
    local expr = |||
      sum(storage_level_compact_frequency) by (instance, group, result)
    |||;

    timeseriesPanel.new(
      title='Compaction Success & Failure Counts (Detailed)',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

    max_compaction_task_duration::
    local expr = |||
      histogram_quantile(1, sum(rate(state_store_compact_task_duration_bucket[1m])) by (le, job, instance))
    |||;

    timeseriesPanel.new(
      title='Longest Compaction Duration',
      unit='s',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

    median_compaction_task_duration::
    local expr = |||
      histogram_quantile(0.5, sum(rate(state_store_compact_task_duration_bucket[1m])) by (le, job, instance))
    |||;

    timeseriesPanel.new(
      title='Median Compaction Duration',
      unit='s',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

    compaction_write_rate::
    local expr = |||
      sum(rate(storage_level_compact_write[1m])) by(job,instance)
    |||;

    timeseriesPanel.new(
      title='Compaction Rate of Writing',
      unit='Bps',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
      )
    ) + span.span('half'),

    sorted_string_table_file_size::
    local expr = |||
      sum(storage_level_total_file_size) by (instance, level_index)
    |||;

    timeseriesPanel.new(
      title='Sorted String Table File Sizes',
      unit='decbytes',
    ).addTarget(
      prometheus.target(
        expr=strings.trim(expr),
        legendFormat='{{ risingwave_component }}',
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

    panels.max_compaction_task_duration,
    panels.median_compaction_task_duration,

    panels.compaction_write_rate,
    panels.sorted_string_table_file_size,
  ]
)
