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

We use the scripts provided by Risingwave to generate new dashboard. Please first read [README in Risingwave repo](https://raw.githubusercontent.com/risingwavelabs/risingwave/main/grafana/README.md) for toolchain details.

```shell
# use RISINGWAVE_DASHBOARD_COMMIT_ID to specify commit to use, default: "main"
RISINGWAVE_DASHBOARD_COMMIT_ID={{commit_id}} ./generate_risingwave_dashboard.sh
```
