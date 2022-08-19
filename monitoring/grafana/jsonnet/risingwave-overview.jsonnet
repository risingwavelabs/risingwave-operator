local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local template = grafana.template;

local rows = import 'panels/risingwave_overview_rows.libsonnet';

dashboard.new(
  title='RisingWave Overview',
  time_from='now-1h',
  time_to='now',
  editable=true,
  tags=['RisingWave', 'Streaming Database', 'Singularity Data']
).addTemplates(
  [
    template.datasource(
      'PROMETHEUS_DS',
      'prometheus',
      'Prometheus',
      hide='label',
    ),
    template.new(
      'namespace',
      '$PROMETHEUS_DS',
      'label_values(up{risingwave_name=~".+"}, namespace)',
      label='Namespace',
      refresh='load',
    ),
    template.new(
      'instance',
      '$PROMETHEUS_DS',
      'label_values(up{risingwave_name=~".+"}, risingwave_name)',
      label='RisingWave',
      refresh='load',
    ),
  ]
).addRows(
  [
    rows.resource_utilities,
  ]
)
