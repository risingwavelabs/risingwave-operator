# Introduction
1. This is a [ramp up task](https://github.com/risingwavelabs/risingwave-operator/pull/350) for people who are not familiar with Kubebuilder. 
2. During the development, these resources maybe helpful: [k8s ducoment](https://kubernetes.io/docs/home/), [kuberbuilder book](https://book.kubebuilder.io/), [our blog](https://www.risingwave-labs.com/blog/one-step-towards-cloud-the-risingwave-operator/).
3. You will get familiar with some basic concepts in k8s 
   1. How to create a yaml file to apply  
   2. [CustomResourceDefinitions](https://book.kubebuilder.io/reference/generating-crd.html), kubebuilder

# What is Operator

1. [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) is costomized resource definition. Regardless of the abstraction concepts, for k8s, all it need to manage are pods. 
2. [Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) are clients of the Kubernetes API that act as controllers for a Custom Resource.
3. We want to define our own CRD like `risingwave` which is an abstract concepts consist of service, machines and others resource. And k8s will help us to management it. In codebase, we define a golang struct `RisingWave` at [here](/apis/risingwave/v1alpha1/risingwave_types.go)
4. K8s use declarative API. We told k8s what we want, and then it will automatically achieve it. In practice, we define the YAML file, and then use `kubectl apply -f <YAML file>` to apply the change.
5. K8s will use controller to monitor these change and then contoll the real machine. We need to implement the logic of the controller.
6. After finis the `kubectl apply -f <YAML file>` according to the [instruction](/README.md). The controller will apply the change. Then you can use `kubectl get crd` to check the CRD you defined, and use `kubectl get rw` to check the status of your CRD instance.


# Some Useful Commands

`Kubectl get pods`

Show the status of pods
```bash
NAME                                              READY   STATUS    RESTARTS        AGE
risingwave-in-memory-compactor-57557dc6fc-q4496   1/1     Running   0               26m
risingwave-in-memory-compute-0                    1/1     Running   0               26m
risingwave-in-memory-frontend-59f59675b-fl8b4     1/1     Running   0               26m
risingwave-in-memory-meta-0                       1/1     Running   0               26m
```

`kubectl describe pod <pod name>`

Show the status of the specific pod. It is useful when you node status is not correct. You can use it to check the detailed error.

`Kubectl get crd`

Show currently all the CRDs types.
```bash
NAME                                                   CREATED AT
certificaterequests.cert-manager.io                    2023-02-24T00:33:56Z
certificates.cert-manager.io                           2023-02-24T00:33:56Z
challenges.acme.cert-manager.io                        2023-02-24T00:33:56Z
clusterissuers.cert-manager.io                         2023-02-24T00:33:56Z
issuers.cert-manager.io                                2023-02-24T00:33:56Z
orders.acme.cert-manager.io                            2023-02-24T00:33:56Z
risingwavepodtemplates.risingwave.risingwavelabs.com   2023-02-24T00:34:03Z
risingwaves.risingwave.risingwavelabs.com              2023-02-24T00:34:03Z
risingwavescaleviews.risingwave.risingwavelabs.com     2023-02-24T00:34:04Z
```

`Kubectl get risingwave`  or   `Kubectl get rw`

Show the status of risingwave CRD 
```bash
NAME                   RUNNING   STORAGE(META)   STORAGE(OBJECT)   AGE
risingwave-in-memory   True      Memory          Memory            24m
```

`kubectl describe risingwave`  or   `kubectl describe rw`

Show the basic information of risingwave

