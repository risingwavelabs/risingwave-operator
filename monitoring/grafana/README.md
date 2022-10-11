## Initialize Project

```shell
cd jsonnet
make vendor
```

## Update Operator Panels

Make your changes in `jsonnet/panels`
Register changes by running update command

```shell
make update
```

## Update Dashboard Panels

```shell
# `python3 update_risingwave_dashboard.py` equals to `python3 update_risingwave_dashboard.py main risingwave-dashboard.json`
python3 update_risingwave_dashboard.py {{commit_id or branch name}} {{output_file}}
```
