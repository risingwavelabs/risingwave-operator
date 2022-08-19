local grafana = import 'grafonnet/grafana.libsonnet';
local template = grafana.template;

{
  templates+: [
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
      'risingwave',
      '$PROMETHEUS_DS',
      'label_values(up{risingwave_name=~".+"}, risingwave_name)',
      label='RisingWave',
      refresh='load',
    ),
  ],
}
