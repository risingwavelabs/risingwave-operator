# Introduction
1. This is a [ramp up task](https://github.com/risingwavelabs/risingwave-operator/pull/350) for people who are not familiar with Kubebuilder. 
2. During the development, these resources maybe helpful: [k8s ducoment](https://kubernetes.io/docs/home/), [kuberbuilder book](https://book.kubebuilder.io/), [our blog](https://www.risingwave-labs.com/blog/one-step-towards-cloud-the-risingwave-operator/).
3. You will get familiar with some basic concepts in k8s 
   1. How to create a yaml file to apply  
   2. [CustomResourceDefinitions](https://book.kubebuilder.io/reference/generating-crd.html), kubebuilder
4. Some people hope to build their own custom resources (CRD-Customize Resource Definition) like build-in resources, and then write a corresponding controller for the custom resources to launch their own declarative API. K8s provides a CRD extension method to meet the needs of users, and because this extension method is very flexible, a considerable enhancement has been made to CRD in the latest version 1.15. For users, implementing CRD extension mainly does two things:
   1. Write CRD and deploy it to the K8s cluster
      - The function of this step is to let K8s know that there is this resource and its structural attributes. 
      - When the user submits the definition of the custom resource (usually YAML file definition), K8s can successfully verify the resource and create a corresponding Go struct for persistence , while triggering the tuning logic of the controller.

   2. Write the Controller and deploy it to the K8s cluster. The role of this step is to implement the tuning logic.

5. Kubebuilder is a tool to help us simplify these two things, and now we start to introduce the protagonist.


# What is the Operator

1. [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) is costomized resource definition. Regardless of the abstraction concepts, for k8s, all it needs to manage are pods. 
2. [Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) are clients of the Kubernetes API that act as controllers for a Custom Resource.
3. We want to define our own CRD like `risingwave` which is abstract concept consisting of services, machines and other resources. And k8s will help us to manage it. In the codebase, we define a golang struct `RisingWave` at [here](/apis/risingwave/v1alpha1/risingwave_types.go)
4. K8s use declarative API. We told k8s what we want, and then it will automatically achieve it. In practice, we define the YAML file, and then use `kubectl apply -f <YAML file>` to apply the change.
5. K8s will use the controller to monitor these changes and then control the real machine. We need to implement the logic of the controller.
6. After finis the `kubectl apply -f <YAML file>` according to the [instruction](/README.md). The controller will apply the change. Then you can use `kubectl get crd` to check the CRD you defined, and use `kubectl get rw` to check the status of your CRD instance.


# Some Useful Commands

`kubectl get pods`

Show the status of pods
```bash
NAME                                              READY   STATUS    RESTARTS        AGE
risingwave-in-memory-compactor-57557dc6fc-q4496   1/1     Running   0               26m
risingwave-in-memory-compute-0                    1/1     Running   0               26m
risingwave-in-memory-frontend-59f59675b-fl8b4     1/1     Running   0               26m
risingwave-in-memory-meta-0                       1/1     Running   0               26m
```

`kubectl describe pod <pod name>`

Show the status of the specific pod. It is useful when your node status is not correct. You can use it to check the detailed error.

`kubectl get crd`

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

`kubectl get risingwave`  or   `kubectl get rw`

Show the status of RisingWave CR 
```bash
NAME                   RUNNING   STORAGE(META)   STORAGE(OBJECT)   AGE
risingwave-in-memory   True      Memory          Memory            24m
```

`kubectl describe risingwave`  or   `kubectl describe rw`

Show the basic information of risingwave


# The field in Yaml file
All the [obejct](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/) has the following fileds
 1. ApiVersion + Kind can represent one class of Object
 2. metaData
    1. name: defined by ourselves
    2. namespace: "" means default
    3. `namepace` + `name` can point to object in cluster which is the instance of Object 
 3. 1. spec: defined by user
 4. status: the operator will update it

# Codebase Framework

- `apis/risingwave/v1alpha1`
  - `*_types.go` : define the CRD
  - `groupversion_info.go`: code related to CRD generation and schema generation
  - `zz_generated.deepcopy.go`: auto generated, the implementation of the runtime.object interface
- `config`
  - `crd`: Yaml File for CRD deployment
  - `manager`: Yaml File for Controller deployment
  - `sample`: crd samples
  - `rbac`: the rbac permission that controller needed when running
- `pkg`: the main logic of controller
- `cmd/manager/manager.go`: the entry point
  - Init a manager
  - Init a controller
  - Start the manager
- `pkg/controller/risingwave_controller.go`: reconcile logic


# How Kubectl works
 When we use `Kubectl get risingwave`, we will get 
```bash
NAME                   RUNNING   STORAGE(META)   STORAGE(OBJECT)   AGE
risingwave-in-memory   True      Memory          Memory            24m
```
1. The k8s will record some basic information about the object, and the `kubectl get` will select some of the fields and show them.
2. We need to define the CRD `risingwave` and then generate the yaml file. Use yaml file to register the CRD into k8s.
3. Then we can create the `risingwave` instance.
    ```
    kubectl apply -f https://raw.githubusercontent.com/risingwavelabs/risingwave-operator/main/docs/manifests/risingwave/risingwave-in-memory.yaml
    ```
4. The operator is the plugin of the k8s. The k8s can only deal with part of the logic. For the logic cannot deal with, we should impelement them in controller. 

# How to implement
## 1. How to define the additional print column
1. The `RUNNING`, `STORAGE(META)`, `STORAGE(OBJECT)`, `AGE` are [Additional Printer Columns](https://book.kubebuilder.io/reference/generating-crd.html#additional-printer-columns) defined by ourselves. 
2. We get these value from yaml file then print.
3. In the codebase, we define it [here](/apis/risingwave/v1alpha1/risingwave_types.go). We define them in comments in some specific format and the kubebuilder will parse it. We need to define `name`, `type`, and `JSONPath`. For detailed information about `JSONPath`, check [here](https://kubernetes.io/docs/reference/kubectl/jsonpath/).
   
```golang
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rw,categories=all;streaming
// +kubebuilder:printcolumn:name="RUNNING",type=string,JSONPath=`.status.conditions[?(@.type=="Running")].status`
// +kubebuilder:printcolumn:name="STORAGE(META)",type=string,JSONPath=`.status.storages.meta.type`
// +kubebuilder:printcolumn:name="STORAGE(OBJECT)",type=string,JSONPath=`.status.storages.object.type`
// +kubebuilder:printcolumn:name="VERSION",type=string,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`
```


## 2. Define the struct if needed

We can use `JSONPath = .spec.XXX`, `JSONPath = .metadata.XXX` or `JSONPath = .status.XXX`. If you want to add some new field, in codebase, we need to change the struct field [here](../../apis/risingwave/v1alpha1/risingwave_types.go)

```golang
type RisingWave struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   RisingWaveSpec   `json:"spec,omitempty"`
    Status RisingWaveStatus `json:"status,omitempty"`
}
```


## 3. How to register CRD into k8s

We use `make manifests` and `make install-local` to create a new YAML file. The command will generate new [Yaml file](/config/crd/bases/risingwave.risingwavelabs.com_risingwaves.yaml) according to [golang file](../../apis/risingwave/v1alpha1/risingwave_types.go), and regist the CRD `risingwave` into k8s. Then, the user can create the `risingwave` instance in k8s.


## 4. How to update the column
1. Firstly, because `JSONPath` can only support some simple logic, but sometime we want some complex logic. In order to achieve the complex logic, we need to define the field as `.status.XX`. 
2. If you use the `metaData` or `Spec`, the value should be immutable, so we will not change them. But if you define the `status`, we need some extract work to update it. 
3. Create updating logic in controller. We need to implement the logic in controller. In codebase, we implement it [here](/pkg/manager/risingwave_controller_manager_impl.go).

```golang
mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
    // Report meta storage status.
    metaStorage := &risingwave.Spec.Storages.Meta
    status.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorageStatus{
        Type: buildMetaStorageType(metaStorage),
    }

    // Report object storage status.
    objectStorage := &risingwave.Spec.Storages.Object
    status.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorageStatus{
        Type: buildObjectStorageType(objectStorage),
    }

    // Report Version status.
    status.Version = utils.GetVersionFromImage(mgr.risingwaveManager.RisingWave().Spec.Global.Image)

    // Report component replicas.
    status.ComponentReplicas = componentReplicas
})
```

4. Currently, if we run `Kubectl get risingwave`, we will get that:
    ```bash
    NAME                   RUNNING   STORAGE(META)   STORAGE(OBJECT)   VERSION   AGE
    risingwave-in-memory   True      Memory          Memory                      20m
    ```
    Because we have not start the controller, there is not value in `VERSION`. Run `make run-local` to start controller, the controller will update the field:
    ```bash
    NAME                   RUNNING   STORAGE(META)   STORAGE(OBJECT)   VERSION   AGE
    risingwave-in-memory   True      Memory          Memory            v0.1.16   20m
    ```
5. When we launch the k8s, the k8s will update the name. The controller will update the `status` field, the k8s will collect the column we want and show it.



## 5. Add unit test


# Reconcile 
1. It is an infinite loop (actually implemented by event-driven + timing synchronization) that constantly compares the desired state with the actual state, and if there is any discrepancy, the Reconcile (tuning) logic is performed to adjust the actual state to the desired state. 
2. The expected state is our object definition (usually a YAML file), the actual state is the current running state in the cluster (usually from the status summary of related resources inside and outside the K8s cluster). The final result of tuning is generally some kind of write operation on the controlled object, such as adding/deleting/modifying Pods.
3. The Reconcile uses a structure concurrency method. Different actions and reconciles are performed according to different events. Although the k8s change is synchronous, it needs to be entered multiple times before it can be completed.
4. The meaning of Reconcile is to synchronize the differences between the current status and the expected result.
5. All things are done in the [risingwave controller](/pkg/controller/risingwave_controller.go), the most important thing is `workflow`. Better understand the code with the [blog](https://www.risingwave-labs.com/blog/one-step-towards-cloud-the-risingwave-operator/)
6. `Join` means that all branches will go in, but `sequential` means that everything is executed sequentially. If the first step is wrong, then the subsequent ones will not be done.
