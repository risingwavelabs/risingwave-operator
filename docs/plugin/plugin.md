# RisingWave Kubectl Plugin

## Prerequisites

- Kubernetes >= v1.19.0.
- kubectl installed on your local machine, configured to talk to the Kubernetes cluster.

## Install Plugin

```shell

```

## Management Commands

### Operator Deployment

Command: `kubectl rw install [options]`

Install RisingWave Operator in the cluster.

Options:

- `--version` the image version for risingwave-operator, if not set, will use the latest image.


### Operator Deletion
Command: `kubectl rw uninstall`

Uninstall the RisingWave Operator from the cluster

## Basic Commands

#### RisingWave Instance Creation

Command: `kubectl rw create INSTANCE_NAME [options]`

Create a risingwave named example-rw by configuration file.

example: `kubectl rw create example-rw -n test -c example.toml`

Options:

- `--namespace=risingwave`
- `--config=example.toml`

#### RisingWave Instance Deletion

Command: `kubectl rw delete INSTANCE_NAME [options]`

Add new volumes (and nodes) to existing MinIO Tenant.

example: `kubectl rw delete example -n risingwave`

Options:

- `--namespace=risingwave`

#### RisingWave Instance List

Command: `kubectl minio tenant info TENANT_NAME [options]`

List all existing MinIO pools in the given MinIO Tenant.

example: `kubectl minio tenant info tenant1`

Options:

- `--namespace=minio`
