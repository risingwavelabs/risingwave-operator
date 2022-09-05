|                    |                                                                          |
| -------            |--------------------------------------------------------------------------|
| Feature            | kubectl-rw: a kind of kubectl plugin                                     |
| Status             | Completed                                                                |
| Date               | 2022-06-30                                                               |
| Authors            | Luke & Xinyu                                                             |
| RFC PR #           | [#100](https://github.com/risingwavelabs/risingwave-operator/pull/100) |
| Implementation PR #| [#175](https://github.com/risingwavelabs/risingwave-operator/pull/175) |
|                    |                                                                          |

# **Summary**

kubectl-rw: a kubectl-plugin to deploy and manage RisingWave operator and instances.

# **Motivation**

For convenience, we will design a command line tool that allows users to quickly manage the RisingWave operator and instances.

For example, user can do the operations as follows:

```shell

kubectl rw install # deploy the operator in the kubernetes cluster

kubectl rw create --meta-storage memory --name xxx # create a RisingWave instance in the kubernetes cluster and use the memory storage for mete-node

kubectl rw upgrade # upgrade the RisingWave instance to the latest version
```
### Supported Commands (Continually updated)

| Command   | Description                                                     |
|-----------|-----------------------------------------------------------------|
| install   | deploy the RisingWave operator.                                 |
| uninstall | delete the RisingWave operator.                                 |
| create    | create a RisingWave instance.                                   |
| delete    | delete a specific RisingWave instance.                          |
| describe  | get the human-readable status of the RisingWave instance.       |
| list      | list all RisingWave instances.                                  |
| update    | update the RisingWave instance.                                 |
| upgrade   | upgrade the RisingWave instance to the latest version.          |
| scale     | scale the RisingWave instance.                                  |
| stop      | stop but not delete the RisingWave instance.                    |
| resume    | resume the stopped RisingWave instance.                         |
| restart   | delete and re-create the RisingWave instance.                   |
| config    | update the configuration of the RisingWave instance on the fly. |



# **Ref**

* [Kubernetes -- kubectl-plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)

# **Explanation**
### Communication with API
The plugin will use `client-go` to communicate with kube-apiserver follows these rules:
1. If the --kubeconfig flag is set, will use that file to communicate with the API server of the cluster.
2. Otherwise, ${HOME}/.kube/config is used

# **Future Possibilities**
In the future, the plugin maybe be integrated into [risedev](https://github.com/risingwavelabs/risingwave/blob/main/risedev)
