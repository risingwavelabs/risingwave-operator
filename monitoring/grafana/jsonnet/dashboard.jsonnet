local grafana = import 'grafonnet/grafana.libsonnet';

local dashboard = grafana.dashboard;
local template = grafana.template;

local lib = import 'mixon/lib.libsonnet';

dashboard.new(
  'RisingWave Dashboard',
  schemaVersion=16,
  tags=['RisingWave', 'Streaming Database', 'Singularity Data']
).addTemplates(
  lib.templates
)
