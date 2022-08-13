# RisingWave Kubectl Plugin

## Prerequisites

- Kubernetes >= v1.19.0.
- kubectl installed on your local machine, configured to talk to the Kubernetes cluster.

## Install Plugin

```shell
make build-plugin
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

### RisingWave Instance Creation

Command: `kubectl rw create INSTANCE_NAME [options]`

Create a RisingWave instance

Examples:

- Create a risingwave named example-rw in the test namespace by default configuration.

`kubectl rw create example-rw -n test`

- Create a risingwave named example-rw by configuration file.

`kubectl rw create -c rw.config`

Options:

- `-a, --arch=''` The default arch(will be override if config file also set the arch).

- `-c, --config=''` The config file used when creating the instance.

- `-n, --namespace=risingwave` Namespace of risingwave instance. If not set, select the default namespace.

### RisingWave Instance Deletion

Command: `kubectl rw delete INSTANCE_NAME [options]`

Delete the RisingWave instance

Examples:

- Delete the risingwave named example-rw in the test namespace.

`kubectl rw delete example-rw -n test`

- Force delete the risingwave named example-rw in the test namespace.
  
`kubectl rw delete example-rw -n test -f`

Options:
    
- `-f, --force=false` Force delete the tenant

- `--namespace=risingwave` Namespace of risingwave instance. If not set, select the default namespace.

### RisingWave Instance List

Command: `kubectl rw list INSTANCE_NAME [options]`

Examples:

- List all clusters and sync to local config

`kubectl rw list`

- Filter by namespace

`kubectl rw list --namespace=foo`

- Get risingwave instances by selector

`kubectl rw list -l foo=bar`

Options:

- `-A, --all-namespaces=false` Whether list instances in all namespaces.

- `-l, --selector=''` Selector (label query) to filter on, not including uninitialized ones, supports '=', '==', and '!='. (e.g. -l key1=value1,key2=value2)

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.

### RisingWave Instance Describe

Command: `kubectl rw describe INSTANCE_NAME [options]`

Describe a risingwave instance.

Examples:

- Describe risingwave named example-rw.

`kubectl rw describe example-rw`

- Describe risingwave instance named example-rw in namespace foo.
   
`kubectl rw describe example-rw -n foo`

Options:

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.

- `-c, --choice='spec'` The section of the risingwave instance you would like to describe. Spec, status and all are the only valid values.

## Deployment commands

### Scale

Command: `kubectl rw scale INSTANCE_NAME [options]`

Scale a risingwave Instance

Examples:

- Scale compute-node of the risingwave named example-rw to 2.
  
`kubectl rw scale example-rw -t 2`

- Scale frontend of the risingwave named example-rw to 2 in the foo namespace.

`kubectl rw scale example-rw -n foo -c frontend -t 2`

- Scale frontend of the risingwave which named example-rw to 2 and in the foo namespace and in the test group.

`kubectl rw scale example-rw -n foo -c frontend -t 2 -g test`

Options:

- `-c, --component='compute-node'` The component which you want to scale.

- `-g, --group='default'` The group name of the component. If not set, scale the default group

- `-t, --target=-1` The target number.Describe risingwave named example-rw.

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.

### Stop

Command: `kubectl rw stop INSTANCE_NAME [options]`

Stop a risingwave instance.

Examples:

- Stop risingwave named example-rw in default namespace.
  
`kubectl rw stop example-rw`

- Stop risingwave named example-rw in foo namespace.

`kubectl rw stop example-rw -n foo`

Options:

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.

### Restart

Command: `kubectl rw restart INSTANCE_NAME [options]`

Restart a risingwave instance.

Examples:

- Restart risingwave named example-rw in default namespace.

`kubectl rw restart example-rw`

- Restart risingwave named example-rw in foo namespace.
  
`kubectl rw restart example-rw -n foo`

Options:

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.

### Resume

command: `kubectl rw resume instance_name [options]`

resume a risingwave instance.

examples:

- Resume risingwave named example-rw in default namespace.

`kubectl rw resume example-rw`

- Resume risingwave named example-rw in foo namespace.
  
`kubectl rw resume example-rw -n foo`

options:

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.

## Configuration commands

### Update 

command: `kubectl rw update instance_name [options]`

Update the CPU and memory configuration for risingwave instances.

Limits and requests for CPU resources are measured in cpu units while that of memory resources are measured in bytes.

Accepted values for resources:

- CPU: Plain integer or using millicpu. For example, 1.0 or 100m, these are equivalent.

- Memory: Plain integer or as a fixed-point number using one of these quantity suffixes: E, P, T, G, M, k. You can also use the power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki. For example, 1G, 1Gi, 1024M or 128974848.

Examples:

- Update compute request and limit config of global component in risingwave named example-rw.

`kubectl rw update example-rw --cpurequest 200m --cpulimit 1000m`

- Update memory request of global component in risingwave named example-rw in namespace foo.

`kubectl rw update example-rw -n foo --memoryrequest 256Mi`

- Update memory request of meta component in risingwave named example-rw in namespace foo and group test.
  
`kubectl rw update example-rw -n foo -c meta -g test --memoryrequest 256Mi`

Options:

- `-c, --component='global'` The component to be updated. If not set, update global resources.

- `--cpulimit=''` The target cpu limit.

- `--cpurequest=''` The target cpu request.

- `-g, --group='default'` The group to be updated. If not set, update the default group.

- `--memorylimit=''` The target memory limit.

- `--memoryrequest=''` The target memory request.resume a risingwave instance.

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.

### Upgrade 

Command: `kubectl rw upgrade INSTANCE_NAME [options]`

Upgrade a risingwave instance to a specified version.

Examples:

- Upgrade risingwave named example-rw to the latest version.

`kubectl rw upgrade example-rw`

- Upgrade risingwave named example-rw in namespace foo to the nightly version.
  
`kubectl rw upgrade example-rw -n foo -v nightly`

Options:

- `-v, --version='latest'` The version to upgrade to. If not specified, the latest version will be used.

- `-n, --namespace='default'` Namespace of risingwave instance. If not set, select the default namespace.