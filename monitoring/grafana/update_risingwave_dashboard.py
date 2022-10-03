#!/usr/bin/python3

import json
import requests
import sys

err_message = "Usage: python3 update_risingwave_dashboard.py {commit_id or branch name, optional, default: main} {output_file, optional, default: ./risingwave-dashboard.json}"
if len(sys.argv) > 3: 
    print(err_message)
    exit(1)

commit_id = "main"
if len(sys.argv) >= 2:
    commit_id = sys.argv[1]
    
output_file = "./risingwave-dashboard.json"
if len(sys.argv) == 3:
    output_file = sys.argv[2]

url = "https://raw.githubusercontent.com/risingwavelabs/risingwave/{commit}/grafana/risingwave-dashboard.json"
response = requests.get(url.format(commit = commit_id))
content = response.content.decode("utf-8")
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
    if contains_str(target[expr_key], "job"):
        target[expr_key] = target[expr_key].replace("job", "risingwave_component")
    if contains_str(target[expr_key], "instance"):
        target[expr_key] = target[expr_key].replace("instance", "pod")
    if contains_str(target[expr_key], "[$__rate_interval]"):
        if contains_str(target[expr_key], "}[$__rate_interval]"):
            target[expr_key] = target[expr_key].replace("}[$__rate_interval]", ", namespace=\"$namespace\", risingwave_name=\"$instance\"}[$__rate_interval]")
            index = target[expr_key].find("[$__rate_interval]")
            found = True
            while found :
                if index - 1 > 0 and target[expr_key][index - 1] != '}':
                    target[expr_key] = target[expr_key].replace("[$__rate_interval]", "{namespace=\"$namespace\", risingwave_name=\"$instance\"}[$__rate_interval]")
                    index = target[expr_key].find("[$__rate_interval]", index + 1)
                # found next
                index = target[expr_key].find("[$__rate_interval]", index + 1)
                if index == -1:
                    found = False
        else:
            target[expr_key] = target[expr_key].replace("[$__rate_interval]", "{namespace=\"$namespace\", risingwave_name=\"$instance\"}[$__rate_interval]")
    elif contains_str(target[expr_key], ") by ("):
        target[expr_key] = target[expr_key].replace(") by (", "{namespace=\"$namespace\", risingwave_name=\"$instance\"}) by (")
    else:
        target[expr_key] = target[expr_key] + "{namespace=\"$namespace\", risingwave_name=\"$instance\"}"
    
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