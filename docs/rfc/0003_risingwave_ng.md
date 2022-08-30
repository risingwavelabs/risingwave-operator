|                    |                                                                      |
| -------            |----------------------------------------------------------------------|
| Feature            | A new design of RisingWave of flexibility and ease of use            |
| Status             | Completed                                                            |
| Date               | 2022-07-08                                                           |
| Authors            | arkbriar                                                             |
| RFC PR #           | (PR # of this RFC. This helps to track the reviews of this proposal) |
| Implementation PR #| (PR # of code changes made to implement the proposal.)               |
|                    |                                                                      |

# **Summary**

This RFC proposes a new design of our RisingWave CRD targeting flexibility and ease of use:

+ Provide a neat way to declare clusters, mainly for test and development, the essential fields include
  + the image
  + the replicas of each component
  + the resources (unified for all components)
  + a simple node selector
+ Support overriding the pods' spec with templates
+ Support defining groups of pods, so that we can make such thing happen:
  + Declare a group of compactors nodes runs on normal EC2 instances (long-running group) and another group runs on spot instances (scaling group)
  + Specify the upgrade strategy for each component group
+ Group storage, configuration and security specs
+ Support deploying the instances onto mixed-arch nodes

# **Motivation**

The current CRD is simple to use but lacks details. For example, if we want to specify different specs for different RisingWave instances, like enabling the highest privilege for tests or setting DNS configs for production, there's no way for us to do it now. However, if we just embed the Pod's spec into the CRD, it is too complex for humans. So I would like something easy to use in normal cases but provides the flexibility of mutating the advanced details. What this proposal supports is mentioned in the [Summary](#summary) part.

# **Explanation**

The RFC mainly adds a new CRD and rewrites the RisingWave CRD. It introduces the `RisingWavePodTemplate` for declaring pod templates, just like the template in Deployment and StatefulSet.

## RisingWavePodTemplate

```yaml
apiVersion: risingwave.singularity-data.com/v1alpha1
kind: RisingWavePodTemplate
metadata:
  name: template-a
spec:
  metadata: # Pod annotations and labels
  spec:     # Pod Spec, only the first container's spec would be overridden, and the other containers' spec will be copied
```

`RisingWavePodTemplate` is immutable so can not be changed after it's created. When creating a RisingWave that refers to the template, the Webhook should validate if there's such an object and reject if not. And also, the webhook should add finalizers to the object so that it would not be deleted accidentally and remove the finalizers when RisingWave objects are going to be deleted.

## RisingWave

Here's an example [risingwave_ng.yaml](./examples/risingwave_ng.yaml). It mainly includes these parts:

+ `global` for simplified specs, such as image, image pull policy, pod template, resources, node selector and replicas
+ `components` for advanced specs, one can
  + Declare complex groups of pods of components, specifying different templates, images, resources and other things related
  + Specifying the listen ports of components
  + The configs in one `group` should be merged into the globals, e.g.,
    + maps are merged and overridden if there're new values (resources)
    + lists are merged
    + values are overridden
  + The groups here are extended groups to the global ones, i.e. if there're 2 compactors in global and a group of 2 compactors in components, there will be 4 in total
+ `configuration` for using pre-defined configs
+ `storages` for specifying the storage types and credentials, as well as the PVC template for compute nodes
+ `security` for specifying the TLS configs

Also, the `podTemplate` will be read and applied onto the Pods of RisingWave if it is set.

Here're examples to demonstrate the new CRD:

+ 1/1/1/1, 1C1G

```yaml
apiVersion: risingwave.singularity-data.com/v1alpha1
kind: RisingWave
metadata:
  name: test-risingwave
spec:
  global:
    image: ghcr.io/singularity-data/risingwave:latest
    replicas:
      meta: 1
      frontend: 1
      compute: 1
      compactor: 1
    resources:
      limits:
        cpu: 1
        memory: 1Gi
  storages:
    meta:
      memory: true
    object:
      memory: true
```

+ 1/1/1/1, 2 spot compactors on a node group labeled with "spot", 1C1G

```yaml
apiVersion: risingwave.singularity-data.com/v1alpha1
kind: RisingWave
metadata:
  name: test-risingwave
spec:
  # ...
  components:
    compactor:
      groups:
      - name: spot
        replicas: 2
        nodeSelector:
          node-group: spot
```

+ Setting the Pods to be privileged

```yaml
apiVersion: risingwave.singularity-data.com/v1alpha1
kind: RisingWavePodTemplate
metadata:
  name: privileged-pods
spec:
  spec:
    containers:
    - securityContext:
        privileged: true
---
apiVersion: risingwave.singularity-data.com/v1alpha1
kind: RisingWave
metadata:
  name: test-risingwave
spec:
  global:
    podTemplate: privileged-pods
```

+ Mixed arch instance (compactors, explicitly)

```yaml
apiVersion: risingwave.singularity-data.com/v1alpha1
kind: RisingWave
metadata:
  name: test-risingwave
spec:
  global:
    image: ghcr.io/singularity-data/risingwave:latest
    replicas:
      meta: 1
      frontend: 1
      compute: 1
    resources:
      limits:
        cpu: 1
        memory: 1Gi
    nodeSelector:
      kubernetes.io/arch: amd64
  components:
    compactor:
      groups:
      - name: group-amd64
        replicas: 1
        image: ghcr.io/singularity-data/risingwave:latest
        nodeSelector:
          kubernetes.io/arch: amd64
      - name: group-arm64
        image: public.ecr.aws/x5u3w5h6/risingwave-arm:latest
        replicas: 1
        nodeSelector:
          kubernetes.io/arch: arm64
```

+ Restart the frontend Pods (update)

```yaml
apiVersion: risingwave.singularity-data.com/v1alpha1
kind: RisingWave
metadata:
  name: test-risingwave
spec:
  components:
    frontend:
      restartAt: 123456789  # restart timestamp
```

# **Drawbacks**

It's more complex than the previous version and is expected to be harder to understand and write a correct YAML file if the users want an instance with advanced topology.

# **Rationale and Alternatives**

+ Why is this design the best in the space of possible designs?

It balances the complexity between essential and detailed specs. A straightforward way to change the Pod's spec is to embed it into our CRD but it would make the CRD too complex to use. Here in this proposal, we add a CR to package the Pod template so that we can refer to it while declaring our RisingWave, just like referring to ConfigMaps. It also provides an easy way to declare instances for test and dev environments. You can leave your hands out of the advanced parts if they are not necessary.

+ What other designs have been considered and what is the rationale for not choosing them?

As mentioned before.

+ What is the impact of not doing this?

We would have to change our CR frequently to support different requirements, like mutating the Pod's spec and setting upgrade strategies.

# **Unresolved questions**

None

# **Future Possibilities**

Extending the `components` part to support further advanced requirements.
