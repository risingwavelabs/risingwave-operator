local grafana7 = import 'grafonnet-7.0/grafana.libsonnet';
local grafana = import 'grafonnet/grafana.libsonnet';
local timeseriesPanel = grafana.timeseriesPanel;
local tablePanel = grafana7.panel.table;
local prometheus = grafana.prometheus;

local span = import 'mixon/span.libsonnet';
local strings = import 'lib/strings.libsonnet';

{
  component_overview::
    local pod_info_expr = |||
      topk by (pod) (1, max by (pod, node) (kube_pod_info{node!="", namespace="$namespace"}))
      * on (pod) group_left (risingwave_component)
      	topk by (risingwave_component, pod) (
      		1,
      		max by (risingwave_component, pod) (up{namespace="$namespace", risingwave_name="risingwave-tpch-bench"})
      	)
    |||;

    local pod_cpu_expr = |||
      sum by (pod) (
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

    local pod_mem_expr = |||
      avg by (pod) (
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

    tablePanel.new(
      title='Component Overview',
    ).setGridPos(
      w=24
    ).addOverride(
      matcher={ id: 'byName', options: 'risingwave_component' },
      properties=[
        {
          id: 'custom.filterable',
          value: true,
        },
      ]
    ).addOverride(
      matcher={ id: 'byName', options: 'Value #C' },
      properties=[
        {
          id: 'unit',
          value: 'percentunit',
        },
        {
          id: 'custom.displayMode',
          value: 'gradient-gauge',
        },
      ]
    ).addOverride(
      matcher={ id: 'byName', options: 'Value #M' },
      properties=[
        {
          id: 'unit',
          value: 'decbytes',
        },
      ]
    ) + {
      targets+: [
        prometheus.target(
          expr=strings.trim(pod_info_expr),
          format='table',
          instant=true,
        ) + { refId: 'I' },
        prometheus.target(
          expr=strings.trim(pod_cpu_expr),
          format='table',
          instant=true,
        ) + { refId: 'C' },
        prometheus.target(
          expr=strings.trim(pod_mem_expr),
          format='table',
          instant=true,
        ) + { refId: 'M' },
      ],
    } + {
      options: {
        footer: {
          enablePagination: true,
        },
      },
    } + {
      transformations: [
        {
          id: 'seriesToColumns',
          options: {
            byField: 'pod',
          },
        },
        {
          id: 'sortBy',
          options: {
            fields: {},
            sort: [
              {
                field: 'risingwave_component',
                desc: true,
              },
            ],
          },
        },
        {
          id: 'filterFieldsByName',
          options: {
            exclude: {
              names: [
                'Time',
                'Value #I',
              ],
            },
          },
        },
        {
          id: 'organize',
          options: {
            indexByName: {
              pod: 0,
              risingwave_component: 1,
              node: 2,
              'Value #C': 3,
              'Value #M': 4,
            },
            renameByName: {
              node: 'Node',
              risingwave_component: 'Component',
              pod: 'Pod',
              'Value #C': 'CPU',
              'Value #M': 'Mem',
            },
          },
        },

      ],
    },
}
