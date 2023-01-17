#!/usr/bin/python3

import json

input_file = "risingwave-dashboard.json"
output_file = "risingwave-dashboard_new.json"

f = open(input_file,encoding = "utf-8")
content = f.read()
json_data = json.loads(content)

# update ["annotations"]["list"]
annotations_list = [{
    "builtIn": 1,
    "datasource": {
        "type": "grafana",
        "uid": "-- Grafana --"
    },
    "enable": True,
    "hide": True,
    "iconColor": "rgba(0, 211, 255, 1)",
    "name": "Annotations & Alerts",
    "target": {
        "limit": 100,
        "matchAny": False,
        "tags": [],
        "type": "dashboard"
    },
    "type": "dashboard"
}]
templating_list = [
    {
        "current": {
            "selected": False,
            "text": "default",
            "value": "default"
        },
        "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
        },
        "definition": "label_values(up{risingwave_name=~\".+\"}, namespace)",
        "description": "Kubernetes namespace.",
        "hide": 0,
        "includeAll": False,
        "label": "Namespace",
        "multi": False,
        "name": "namespace",
        "options": [],
        "query": {
            "query": "label_values(up{risingwave_name=~\".+\"}, namespace)",
            "refId": "StandardVariableQuery"
        },
        "refresh": 1,
        "regex": "",
        "skipUrlSync": False,
        "sort": 1,
        "type": "query"
    },
    {
        "datasource": {
            "type": "prometheus",
            "uid": "prometheus"
        },
        "definition": "label_values(up{namespace=\"$namespace\", risingwave_name=~\".+\"}, risingwave_name)",
        "description": "RisingWave pod.",
        "hide": 0,
        "includeAll": False,
        "label": "RisingWave",
        "multi": False,
        "name": "instance",
        "options": [],
        "query": {
            "query": "label_values(up{namespace=\"$namespace\", risingwave_name=~\".+\"}, risingwave_name)",
            "refId": "StandardVariableQuery"
        },
        "refresh": 1,
        "regex": "",
        "skipUrlSync": False,
        "sort": 5,
        "type": "query"
    }
]
json_data["annotations"]["list"] = annotations_list
json_data["templating"]["list"] = templating_list
json_data["title"] = "RisingWave Dashboard"
json_data["time"]["from"] = "now-5m"
del json_data["__inputs"]

panels_key = "panels"
datasource_key = "datasource"
targets_key = "targets"
expr_key = "expr"
legend_format_key = "legendFormat"

# define "datasource"
datasource_value = {
    "type": "prometheus",
    "uid": "prometheus"
}

def contains_str(string, target):
    return target in string and target + "_" not in string and "_" + target not in string

def update_expr(target): 
    if contains_str(target[expr_key], "$instance"):
        # mask $instance
        target[expr_key] = target[expr_key].replace("$instance", "XXXXXXXXXXXXXXXXX")
    if contains_str(target[expr_key], "job"):
        target[expr_key] = target[expr_key].replace("job", "risingwave_component")
    if contains_str(target[expr_key], "instance"):
        target[expr_key] = target[expr_key].replace("instance", "pod")
    # unmask $instance
    target[expr_key] = target[expr_key].replace("XXXXXXXXXXXXXXXXX", "$instance")
    
def update_legend_format(target):
    if contains_str(target[legend_format_key], "{{job}}"):
        target[legend_format_key] = target[legend_format_key].replace("{{job}}", "{{risingwave_component}}")
    if contains_str(target[legend_format_key], "{{instance}}"):
        target[legend_format_key] = target[legend_format_key].replace("{{instance}}", "{{pod}}")

def update_targets(targets):
    for target in targets:
        target[datasource_key] = datasource_value
        if expr_key in target:
            update_expr(target)
        if legend_format_key in target:
            update_legend_format(target)
        
def update_panels(panels):
    for panel in panels:
        panel[datasource_key] = datasource_value
        if targets_key in panel:
            update_targets(panel[targets_key])
        if panels_key in panel:
            update_panels(panel[panels_key])
            
            
update_panels(json_data[panels_key])

output = json.dumps(json_data, separators=(',', ':'))
f = open(output_file, 'w', encoding = 'utf-8')
f.write(output)