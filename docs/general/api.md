---
title: "API reference"
description: "RisingWave operator generated API reference docs"
draft: false
images: []
menu:
docs:
parent: "operator"
weight: 208
toc: true
---
> This page is automatically generated with `gen-crd-api-reference-docs`.
<p>Packages:</p>
<ul>
<li>
<a href="#risingwave.risingwavelabs.com%2fv1alpha1">risingwave.risingwavelabs.com/v1alpha1</a>
</li>
</ul>
<h2 id="risingwave.risingwavelabs.com/v1alpha1">risingwave.risingwavelabs.com/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.ComponentGroupReplicasStatus">ComponentGroupReplicasStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">ComponentReplicasStatus</a>)
</p>
<div>
<p>ComponentGroupReplicasStatus are the running status of Pods in group.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the group.</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Target replicas of the group.</p>
</td>
</tr>
<tr>
<td>
<code>running</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Running replicas in the group.</p>
</td>
</tr>
<tr>
<td>
<code>exists</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Existence status of the group.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">ComponentReplicasStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsReplicasStatus">RisingWaveComponentsReplicasStatus</a>)
</p>
<div>
<p>ComponentReplicasStatus are the running status of Pods of the component.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>target</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total target replicas of the component.</p>
</td>
</tr>
<tr>
<td>
<code>running</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Total running replicas of the component.</p>
</td>
</tr>
<tr>
<td>
<code>groups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentGroupReplicasStatus">
[]ComponentGroupReplicasStatus
</a>
</em>
</td>
<td>
<p>List of running status of each group.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">PartialObjectMeta
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroupTemplate">RisingWaveComponentGroupTemplate</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalSpec">RisingWaveGlobalSpec</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
</p>
<div>
<p>PartialObjectMeta contains partial metadata of an object, including labels and annotations.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>labels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Labels of the object.</p>
</td>
</tr>
<tr>
<td>
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Annotations of the object.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.PersistentVolumeClaim">PersistentVolumeClaim
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">RisingWaveStoragesSpec</a>)
</p>
<div>
<p>PersistentVolumeClaim used by RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PersistentVolumeClaimPartialObjectMeta">
PersistentVolumeClaimPartialObjectMeta
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumeclaimspec-v1-core">
Kubernetes core/v1.PersistentVolumeClaimSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>spec defines the desired characteristics of a volume requested by a pod author.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims">https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims</a></p>
<br/>
<br/>
<table>
<tr>
<td>
<code>accessModes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumeaccessmode-v1-core">
[]Kubernetes core/v1.PersistentVolumeAccessMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>accessModes contains the desired access modes the volume should have.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1">https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1</a></p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>selector is a label query over volumes to consider for binding.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>resources represents the minimum resources the volume should have.
If RecoverVolumeExpansionFailure feature is enabled users are allowed to specify resource requirements
that are lower than previous value but must still be higher than capacity recorded in the
status field of the claim.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources">https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources</a></p>
</td>
</tr>
<tr>
<td>
<code>volumeName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeName is the binding reference to the PersistentVolume backing this claim.</p>
</td>
</tr>
<tr>
<td>
<code>storageClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>storageClassName is the name of the StorageClass required by the claim.
More info: <a href="https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1">https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1</a></p>
</td>
</tr>
<tr>
<td>
<code>volumeMode</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumemode-v1-core">
Kubernetes core/v1.PersistentVolumeMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeMode defines what type of volume is required by the claim.
Value of Filesystem is implied when not included in claim spec.</p>
</td>
</tr>
<tr>
<td>
<code>dataSource</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#typedlocalobjectreference-v1-core">
Kubernetes core/v1.TypedLocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>dataSource field can be used to specify either:
* An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot)
* An existing PVC (PersistentVolumeClaim)
If the provisioner or an external controller can support the specified data source,
it will create a new volume based on the contents of the specified data source.
When the AnyVolumeDataSource feature gate is enabled, dataSource contents will be copied to dataSourceRef,
and dataSourceRef contents will be copied to dataSource when dataSourceRef.namespace is not specified.
If the namespace is specified, then dataSourceRef will not be copied to dataSource.</p>
</td>
</tr>
<tr>
<td>
<code>dataSourceRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#typedobjectreference-v1-core">
Kubernetes core/v1.TypedObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>dataSourceRef specifies the object from which to populate the volume with data, if a non-empty
volume is desired. This may be any object from a non-empty API group (non
core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type of
the specified object matches some installed volume populator or dynamic
provisioner.
This field will replace the functionality of the dataSource field and as such
if both fields are non-empty, they must have the same value. For backwards
compatibility, when namespace isn&rsquo;t specified in dataSourceRef,
both fields (dataSource and dataSourceRef) will be set to the same
value automatically if one of them is empty and the other is non-empty.
When namespace is specified in dataSourceRef,
dataSource isn&rsquo;t set to the same value and must be empty.
There are three important differences between dataSource and dataSourceRef:
* While dataSource only allows two specific types of objects, dataSourceRef
allows any non-core object, as well as PersistentVolumeClaim objects.
* While dataSource ignores disallowed values (dropping them), dataSourceRef
preserves all values, and generates an error if a disallowed value is
specified.
* While dataSource only allows local objects, dataSourceRef allows objects
in any namespaces.
(Beta) Using this field requires the AnyVolumeDataSource feature gate to be enabled.
(Alpha) Using the namespace field of dataSourceRef requires the CrossNamespaceVolumeDataSource feature gate to be enabled.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.PersistentVolumeClaimPartialObjectMeta">PersistentVolumeClaimPartialObjectMeta
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.PersistentVolumeClaim">PersistentVolumeClaim</a>)
</p>
<div>
<p>PersistentVolumeClaimPartialObjectMeta is the metadata for PVC templates.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name must be unique within a namespace. Is required when creating resources, although
some resources may allow a client to request the generation of an appropriate name
automatically. Name is primarily intended for creation idempotence and configuration
definition.
Cannot be updated.
More info: <a href="http://kubernetes.io/docs/user-guide/identifiers#names">http://kubernetes.io/docs/user-guide/identifiers#names</a></p>
</td>
</tr>
<tr>
<td>
<code>labels</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Map of string keys and values that can be used to organize and categorize
(scope and select) objects. May match selectors of replication controllers
and services.
More info: <a href="http://kubernetes.io/docs/user-guide/labels">http://kubernetes.io/docs/user-guide/labels</a></p>
</td>
</tr>
<tr>
<td>
<code>annotations</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Annotations is an unstructured key value map stored with a resource that may be
set by external tools to store and retrieve arbitrary metadata. They are not
queryable and should be preserved when modifying objects.
More info: <a href="http://kubernetes.io/docs/user-guide/annotations">http://kubernetes.io/docs/user-guide/annotations</a></p>
</td>
</tr>
<tr>
<td>
<code>finalizers</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Must be empty before the object is deleted from the registry. Each entry
is an identifier for the responsible component that will remove the entry
from the list. If the deletionTimestamp of the object is non-nil, entries
in this list can only be removed.
Finalizers may be processed and removed in any order.  Order is NOT enforced
because it introduces significant risk of stuck finalizers.
finalizers is a shared field, any actor with permission can reorder it.
If the finalizer list is processed in order, then this can lead to a situation
in which the component responsible for the first finalizer in the list is
waiting for a signal (field value, external system, or other) produced by a
component responsible for a finalizer later in the list, resulting in a deadlock.
Without enforced ordering finalizers are free to order amongst themselves and
are not vulnerable to ordering changes in the list.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWave">RisingWave
</h3>
<div>
<p>RisingWave is the struct for RisingWave object.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">
RisingWaveSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>global</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalSpec">
RisingWaveGlobalSpec
</a>
</em>
</td>
<td>
<p>The spec of a shared template for components and a global scope of replicas.</p>
</td>
</tr>
<tr>
<td>
<code>storages</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">
RisingWaveStoragesSpec
</a>
</em>
</td>
<td>
<p>The spec of meta storage, object storage for compute and compactor, and PVC templates for compute.</p>
</td>
</tr>
<tr>
<td>
<code>components</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">
RisingWaveComponentsSpec
</a>
</em>
</td>
<td>
<p>The spec of ports and some controllers (such as <code>restartAt</code>) of each component,
as well as an advanced concept called <code>group</code> to override the global template and create groups
of Pods, e.g., deployment in hybrid-arch cluster.</p>
</td>
</tr>
<tr>
<td>
<code>configuration</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveConfigurationSpec">
RisingWaveConfigurationSpec
</a>
</em>
</td>
<td>
<p>The spec of configuration template for RisingWave.</p>
</td>
</tr>
<tr>
<td>
<code>enableOpenKruise</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to indicate if OpenKruise should be enabled for components.
If enabled, CloneSets will be used for meta/frontend/compactor nodes
and Advanced StateFulSets will be used for compute nodes.</p>
</td>
</tr>
<tr>
<td>
<code>enableDefaultServiceMonitor</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to indicate if a default ServiceMonitor (from Prometheus operator) should be created by the controller.
False and an empty value means the ServiceMonitor won&rsquo;t be created automatically. But even if it&rsquo;s set to true,
the controller will determine if it can create the resource by checking if the CRDs are installed.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image for RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>frontendServiceType</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#servicetype-v1-core">
Kubernetes core/v1.ServiceType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>FrontendServiceType determines the service type of the frontend service. Defaults to ClusterIP.</p>
</td>
</tr>
<tr>
<td>
<code>additionalFrontendServiceMetadata</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">
PartialObjectMeta
</a>
</em>
</td>
<td>
<p>AdditionalFrontendServiceMetadata tells the operator to add the specified metadata onto the frontend Service.
Note that the system reserved labels and annotations are not valid and will be rejected by the webhook.</p>
</td>
</tr>
<tr>
<td>
<code>metaStore</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">
RisingWaveMetaStoreBackend
</a>
</em>
</td>
<td>
<p>MetaStore determines which backend the meta store will use and the parameters for it. Defaults to memory.
But keep in mind that memory backend is not recommended in production.</p>
</td>
</tr>
<tr>
<td>
<code>stateStore</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">
RisingWaveStateStoreBackend
</a>
</em>
</td>
<td>
<p>StateStore determines which backend the state store will use and the parameters for it. Defaults to memory.
But keep in mind that memory backend is not recommended in production.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">
RisingWaveStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveAzureBlobCredentials">RisingWaveAzureBlobCredentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendAzureBlob">RisingWaveStateStoreBackendAzureBlob</a>)
</p>
<div>
<p>RisingWaveAzureBlobCredentials is the reference and keys selector to the AzureBlob access credentials stored in a local secret.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The name of the secret in the pod&rsquo;s namespace to select from.</p>
</td>
</tr>
<tr>
<td>
<code>accountNameRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccountNameKeyRef is the key of the secret to be the account name. Must be a valid secret key.
Defaults to &ldquo;AccountName&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>AccountKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccountKeyRef is the key of the secret to be the secret account key. Must be a valid secret key.
Defaults to &ldquo;AccountKey&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCommonPorts">RisingWaveComponentCommonPorts
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompactor">RisingWaveComponentCompactor</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompute">RisingWaveComponentCompute</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentConnector">RisingWaveComponentConnector</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentFrontend">RisingWaveComponentFrontend</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMetaPorts">RisingWaveComponentMetaPorts</a>)
</p>
<div>
<p>RisingWaveComponentCommonPorts are the common ports that components need to listen.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>service</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Service port of the component. For each component,
the &lsquo;service&rsquo; has different meanings. It&rsquo;s an optional field and if it&rsquo;s left out, a
default port (varies among components) will be used.</p>
</td>
</tr>
<tr>
<td>
<code>metrics</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metrics port of the component. It always serves the metrics in
Prometheus format.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompactor">RisingWaveComponentCompactor
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
<p>RisingWaveComponentCompactor is the spec describes the compactor component.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>restartAt</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time that the Pods of compactor that should be restarted. Setting a value on this
field will trigger a recreation of all Pods of this component.</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCommonPorts">
RisingWaveComponentCommonPorts
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports to be listened by compactor Pods.</p>
</td>
</tr>
<tr>
<td>
<code>groups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroup">
[]RisingWaveComponentGroup
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Groups of Pods of compactor component.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompute">RisingWaveComponentCompute
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
<p>RisingWaveComponentCompute is the spec describes the compute component.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>restartAt</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time that the Pods of compute that should be restarted. Setting a value on this
field will trigger a recreation of all Pods of this component.</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCommonPorts">
RisingWaveComponentCommonPorts
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports to be listened by compute Pods.</p>
</td>
</tr>
<tr>
<td>
<code>groups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComputeGroup">
[]RisingWaveComputeGroup
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Groups of Pods of compute component.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentConnector">RisingWaveComponentConnector
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
<p>RisingWaveComponentConnector is the spec describes the connector component.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>restartAt</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time that the Pods of connector that should be restarted. Setting a value on this
field will trigger a recreation of all Pods of this component.</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCommonPorts">
RisingWaveComponentCommonPorts
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports to be listened by compactor Pods.</p>
</td>
</tr>
<tr>
<td>
<code>groups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroup">
[]RisingWaveComponentGroup
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Groups of Pods of compactor component.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentFrontend">RisingWaveComponentFrontend
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
<p>RisingWaveComponentFrontend is the spec describes the frontend component.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>restartAt</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time that the Pods of frontend that should be restarted. Setting a value on this
field will trigger a recreation of all Pods of this component.</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCommonPorts">
RisingWaveComponentCommonPorts
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports to be listened by the frontend Pods.</p>
</td>
</tr>
<tr>
<td>
<code>groups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroup">
[]RisingWaveComponentGroup
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Groups of Pods of frontend component.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroup">RisingWaveComponentGroup
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompactor">RisingWaveComponentCompactor</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentConnector">RisingWaveComponentConnector</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentFrontend">RisingWaveComponentFrontend</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMeta">RisingWaveComponentMeta</a>)
</p>
<div>
<p>RisingWaveComponentGroup is the common deployment group of each component. Currently, we use
this group for meta/frontend/compactor.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the group.</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Replicas of Pods in this group.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroupTemplate">RisingWaveComponentGroupTemplate
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroup">RisingWaveComponentGroup</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComputeGroupTemplate">RisingWaveComputeGroupTemplate</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalSpec">RisingWaveGlobalSpec</a>)
</p>
<div>
<p>RisingWaveComponentGroupTemplate is the common deployment template for groups of each component.
Currently, we use the common template for meta/frontend/compactor.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image for RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pull policy of the RisingWave image. The default value is the same as the
default of Kubernetes.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Secrets for pulling RisingWave images.</p>
</td>
</tr>
<tr>
<td>
<code>upgradeStrategy</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategy">
RisingWaveUpgradeStrategy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Upgrade strategy for the components. By default, it is the same as the
workload&rsquo;s default strategy that the component is deployed with.
Note: the maxSurge will not take effect for the compute component.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources of the RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A map of labels describing the nodes to be scheduled on.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>priorityClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, indicates the pod&rsquo;s priority. &ldquo;system-node-critical&rdquo; and
&ldquo;system-cluster-critical&rdquo; are two special keywords which indicate the
highest priorities with the former being the highest priority. Any other
name must be defined by creating a PriorityClass object with that name.
If not specified, the pod priority will be default or zero if there is no
default.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext holds pod-level security attributes and common container settings.
Optional: Defaults to empty.  See type description for default values of each field.</p>
</td>
</tr>
<tr>
<td>
<code>dnsConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#poddnsconfig-v1-core">
Kubernetes core/v1.PodDNSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the DNS parameters of a pod.
Parameters specified here will be merged to the generated DNS
configuration based on DNSPolicy.</p>
</td>
</tr>
<tr>
<td>
<code>terminationGracePeriodSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
Value must be non-negative integer. The value zero indicates stop immediately via
the kill signal (no opportunity to shut down).
If this value is nil, the default grace period will be used instead.
The grace period is the duration in seconds after the processes running in the pod are sent
a termination signal and the time when the processes are forcibly halted with a kill signal.
Set this value longer than the expected cleanup time for your process.
Defaults to 30 seconds.</p>
</td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">
PartialObjectMeta
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>metadata of the RisingWave&rsquo;s Pods.</p>
</td>
</tr>
<tr>
<td>
<code>env</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of environment variables to set in the container.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>envFrom</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envfromsource-v1-core">
[]Kubernetes core/v1.EnvFromSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of sources to populate environment variables in the container.
The keys defined within a source must be a C_IDENTIFIER. All invalid keys
will be reported as an event when the container is starting. When a key exists in multiple
sources, the value associated with the last source will take precedence.
Values defined by an Env with a duplicate key will take precedence.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s scheduling constraints</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMeta">RisingWaveComponentMeta
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
<p>RisingWaveComponentMeta is the spec describes the meta component.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>restartAt</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The time that the Pods of frontend that should be restarted. Setting a value on this
field will trigger a recreation of all Pods of this component.</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMetaPorts">
RisingWaveComponentMetaPorts
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ports to be listened by the meta Pods.</p>
</td>
</tr>
<tr>
<td>
<code>groups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroup">
[]RisingWaveComponentGroup
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Groups of Pods of meta component.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMetaPorts">RisingWaveComponentMetaPorts
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMeta">RisingWaveComponentMeta</a>)
</p>
<div>
<p>RisingWaveComponentMetaPorts are the ports of component meta.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>service</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Service port of the component. For each component,
the &lsquo;service&rsquo; has different meanings. It&rsquo;s an optional field and if it&rsquo;s left out, a
default port (varies among components) will be used.</p>
</td>
</tr>
<tr>
<td>
<code>metrics</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Metrics port of the component. It always serves the metrics in
Prometheus format.</p>
</td>
</tr>
<tr>
<td>
<code>dashboard</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Dashboard port of the meta, a default value of 8080 will be
used if not specified.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsReplicasStatus">RisingWaveComponentsReplicasStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus</a>)
</p>
<div>
<p>RisingWaveComponentsReplicasStatus is the running status of components.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>meta</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">
ComponentReplicasStatus
</a>
</em>
</td>
<td>
<p>Running status of meta.</p>
</td>
</tr>
<tr>
<td>
<code>frontend</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">
ComponentReplicasStatus
</a>
</em>
</td>
<td>
<p>Running status of frontend.</p>
</td>
</tr>
<tr>
<td>
<code>compute</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">
ComponentReplicasStatus
</a>
</em>
</td>
<td>
<p>Running status of compute.</p>
</td>
</tr>
<tr>
<td>
<code>compactor</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">
ComponentReplicasStatus
</a>
</em>
</td>
<td>
<p>Running status of compactor.</p>
</td>
</tr>
<tr>
<td>
<code>connector</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">
ComponentReplicasStatus
</a>
</em>
</td>
<td>
<p>Running status of connector.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
</p>
<div>
<p>RisingWaveComponentsSpec is the spec describes the components of RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>meta</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMeta">
RisingWaveComponentMeta
</a>
</em>
</td>
<td>
<p>Meta component spec.The central metadata management service. It also acts as a failure detector that periodically sends heartbeats to frontend nodes and compute nodes in the cluster.</p>
</td>
</tr>
<tr>
<td>
<code>frontend</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentFrontend">
RisingWaveComponentFrontend
</a>
</em>
</td>
<td>
<p>Frontend component spec. A frontend node acts as a stateless proxy that accepts user queries through Postgres protocol. It is responsible for parsing and validating queries, optimizing query execution plans, and delivering query results.</p>
</td>
</tr>
<tr>
<td>
<code>compute</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompute">
RisingWaveComponentCompute
</a>
</em>
</td>
<td>
<p>Compute component spec. A computer node executes the optimized query plans and handles data ingestion and output.</p>
</td>
</tr>
<tr>
<td>
<code>compactor</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompactor">
RisingWaveComponentCompactor
</a>
</em>
</td>
<td>
<p>Compactor component spec. A stateless worker node that compacts data for the storage engine.</p>
</td>
</tr>
<tr>
<td>
<code>connector</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentConnector">
RisingWaveComponentConnector
</a>
</em>
</td>
<td>
<p>Connector component spec. A connector node, which enables the communication with other systems like kinesis or pulsar.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComputeGroup">RisingWaveComputeGroup
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompute">RisingWaveComponentCompute</a>)
</p>
<div>
<p>RisingWaveComputeGroup is the group for component compute.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the group.</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Replicas of Pods in this group.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComputeGroupTemplate">RisingWaveComputeGroupTemplate
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComputeGroup">RisingWaveComputeGroup</a>)
</p>
<div>
<p>RisingWaveComputeGroupTemplate is the group template for component compute, which supports specifying
the volume mounts on compute Pods. The volumes should be either local or defined in the storages.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image for RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pull policy of the RisingWave image. The default value is the same as the
default of Kubernetes.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Secrets for pulling RisingWave images.</p>
</td>
</tr>
<tr>
<td>
<code>upgradeStrategy</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategy">
RisingWaveUpgradeStrategy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Upgrade strategy for the components. By default, it is the same as the
workload&rsquo;s default strategy that the component is deployed with.
Note: the maxSurge will not take effect for the compute component.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources of the RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A map of labels describing the nodes to be scheduled on.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>priorityClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, indicates the pod&rsquo;s priority. &ldquo;system-node-critical&rdquo; and
&ldquo;system-cluster-critical&rdquo; are two special keywords which indicate the
highest priorities with the former being the highest priority. Any other
name must be defined by creating a PriorityClass object with that name.
If not specified, the pod priority will be default or zero if there is no
default.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext holds pod-level security attributes and common container settings.
Optional: Defaults to empty.  See type description for default values of each field.</p>
</td>
</tr>
<tr>
<td>
<code>dnsConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#poddnsconfig-v1-core">
Kubernetes core/v1.PodDNSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the DNS parameters of a pod.
Parameters specified here will be merged to the generated DNS
configuration based on DNSPolicy.</p>
</td>
</tr>
<tr>
<td>
<code>terminationGracePeriodSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
Value must be non-negative integer. The value zero indicates stop immediately via
the kill signal (no opportunity to shut down).
If this value is nil, the default grace period will be used instead.
The grace period is the duration in seconds after the processes running in the pod are sent
a termination signal and the time when the processes are forcibly halted with a kill signal.
Set this value longer than the expected cleanup time for your process.
Defaults to 30 seconds.</p>
</td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">
PartialObjectMeta
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>metadata of the RisingWave&rsquo;s Pods.</p>
</td>
</tr>
<tr>
<td>
<code>env</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of environment variables to set in the container.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>envFrom</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envfromsource-v1-core">
[]Kubernetes core/v1.EnvFromSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of sources to populate environment variables in the container.
The keys defined within a source must be a C_IDENTIFIER. All invalid keys
will be reported as an event when the container is starting. When a key exists in multiple
sources, the value associated with the last source will take precedence.
Values defined by an Env with a duplicate key will take precedence.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s scheduling constraints</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes to be mounted on the Pods.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveCondition">RisingWaveCondition
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus</a>)
</p>
<div>
<p>RisingWaveCondition indicates a condition of RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveConditionType">
RisingWaveConditionType
</a>
</em>
</td>
<td>
<p>Type of the condition</p>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#conditionstatus-v1-meta">
Kubernetes meta/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of the condition</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Last time the condition transitioned from one status to another.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The reason for the condition&rsquo;s last transition.</p>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Human-readable message indicating details about last transition.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveConditionType">RisingWaveConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveCondition">RisingWaveCondition</a>)
</p>
<div>
<p>RisingWaveConditionType is the condition type of RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Failed&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Initializing&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Running&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unknown&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Upgrading&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveConfigurationSpec">RisingWaveConfigurationSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
</p>
<div>
<p>RisingWaveConfigurationSpec is the configuration spec.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>configmap</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#configmapkeyselector-v1-core">
Kubernetes core/v1.ConfigMapKeySelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The reference to a key in a config map that contains the base config for RisingWave.
It&rsquo;s an optional field and can be left out. If not specified, a default config is going to be used.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveEtcdCredentials">RisingWaveEtcdCredentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendEtcd">RisingWaveMetaStoreBackendEtcd</a>)
</p>
<div>
<p>RisingWaveEtcdCredentials is the reference and keys selector to the etcd access credentials stored in a local secret.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The name of the secret in the pod&rsquo;s namespace to select from.</p>
</td>
</tr>
<tr>
<td>
<code>usernameKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>UsernameKeyRef is the key of the secret to be the username. Must be a valid secret key.
Defaults to &ldquo;username&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>passwordKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>PasswordKeyRef is the key of the secret to be the password. Must be a valid secret key.
Defaults to &ldquo;password&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveGCSCredentials">RisingWaveGCSCredentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendGCS">RisingWaveStateStoreBackendGCS</a>)
</p>
<div>
<p>RisingWaveGCSCredentials is the reference and keys selector to the GCS access credentials stored in a local secret.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The name of the secret in the pod&rsquo;s namespace to select from.</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountCredentialsKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceAccountCredentialsKeyRef is the key of the secret to be the service account credentials. Must be a valid secret key.
Defaults to &ldquo;ServiceAccountCredentials&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalReplicas">RisingWaveGlobalReplicas
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalSpec">RisingWaveGlobalSpec</a>)
</p>
<div>
<p>RisingWaveGlobalReplicas are the replicas of each component, declared in global scope.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>meta</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas of meta component. Replicas specified here is in a default group (with empty name &ldquo;).</p>
</td>
</tr>
<tr>
<td>
<code>frontend</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas of frontend component. Replicas specified here is in a default group (with empty name &ldquo;).</p>
</td>
</tr>
<tr>
<td>
<code>compute</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas of compute component. Replicas specified here is in a default group (with empty name &ldquo;).</p>
</td>
</tr>
<tr>
<td>
<code>compactor</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas of compactor component. Replicas specified here is in a default group (with empty name &ldquo;).</p>
</td>
</tr>
<tr>
<td>
<code>connector</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas of connector component. Replicas specified here is in a default group (with empty name &ldquo;).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalSpec">RisingWaveGlobalSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
</p>
<div>
<p>RisingWaveGlobalSpec is the global spec.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image for RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pull policy of the RisingWave image. The default value is the same as the
default of Kubernetes.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br/>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Secrets for pulling RisingWave images.</p>
</td>
</tr>
<tr>
<td>
<code>upgradeStrategy</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategy">
RisingWaveUpgradeStrategy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Upgrade strategy for the components. By default, it is the same as the
workload&rsquo;s default strategy that the component is deployed with.
Note: the maxSurge will not take effect for the compute component.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources of the RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>A map of labels describing the nodes to be scheduled on.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s tolerations.</p>
</td>
</tr>
<tr>
<td>
<code>priorityClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, indicates the pod&rsquo;s priority. &ldquo;system-node-critical&rdquo; and
&ldquo;system-cluster-critical&rdquo; are two special keywords which indicate the
highest priorities with the former being the highest priority. Any other
name must be defined by creating a PriorityClass object with that name.
If not specified, the pod priority will be default or zero if there is no
default.</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext holds pod-level security attributes and common container settings.
Optional: Defaults to empty.  See type description for default values of each field.</p>
</td>
</tr>
<tr>
<td>
<code>dnsConfig</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#poddnsconfig-v1-core">
Kubernetes core/v1.PodDNSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the DNS parameters of a pod.
Parameters specified here will be merged to the generated DNS
configuration based on DNSPolicy.</p>
</td>
</tr>
<tr>
<td>
<code>terminationGracePeriodSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
Value must be non-negative integer. The value zero indicates stop immediately via
the kill signal (no opportunity to shut down).
If this value is nil, the default grace period will be used instead.
The grace period is the duration in seconds after the processes running in the pod are sent
a termination signal and the time when the processes are forcibly halted with a kill signal.
Set this value longer than the expected cleanup time for your process.
Defaults to 30 seconds.</p>
</td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">
PartialObjectMeta
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>metadata of the RisingWave&rsquo;s Pods.</p>
</td>
</tr>
<tr>
<td>
<code>env</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of environment variables to set in the container.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>envFrom</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#envfromsource-v1-core">
[]Kubernetes core/v1.EnvFromSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of sources to populate environment variables in the container.
The keys defined within a source must be a C_IDENTIFIER. All invalid keys
will be reported as an event when the container is starting. When a key exists in multiple
sources, the value associated with the last source will take precedence.
Values defined by an Env with a duplicate key will take precedence.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod&rsquo;s scheduling constraints</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalReplicas">
RisingWaveGlobalReplicas
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Replicas of each component in default groups.</p>
</td>
</tr>
<tr>
<td>
<code>serviceType</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#servicetype-v1-core">
Kubernetes core/v1.ServiceType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Service type of the frontend service.</p>
</td>
</tr>
<tr>
<td>
<code>serviceMetadata</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">
PartialObjectMeta
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Service metadata of the frontend service.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">RisingWaveMetaStoreBackend
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">RisingWaveStoragesSpec</a>)
</p>
<div>
<p>RisingWaveMetaStoreBackend is the collection of parameters for the meta store that RisingWave uses. Note that one
and only one of the first-level fields could be set.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>memory</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Memory indicates to store the metadata in memory. It is only for test usage and strongly
discouraged to be set in production. If one is using the memory storage for meta,
replicas will not work because they are not going to share the same metadata and any kinds
exit of the process will cause a permanent loss of the data.</p>
</td>
</tr>
<tr>
<td>
<code>etcd</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendEtcd">
RisingWaveMetaStoreBackendEtcd
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Endpoint of the etcd service for storing the metadata.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendEtcd">RisingWaveMetaStoreBackendEtcd
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">RisingWaveMetaStoreBackend</a>)
</p>
<div>
<p>RisingWaveMetaStoreBackendEtcd is the collection of parameters for the etcd backend meta store.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveEtcdCredentials">
RisingWaveEtcdCredentials
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RisingWaveEtcdCredentials is the credentials provider from a Secret. It could be optional to mean that
the etcd service could be accessed without authentication.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<p>Endpoint of etcd. It must be provided.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Secret contains the credentials of access the etcd, it must contain the following keys:
* username
* password
But it is an optional field. Empty value indicates etcd is available without authentication.
Deprecated: Please use &ldquo;credentials&rdquo; field instead. The &ldquo;Secret&rdquo; field will be removed in a future release.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendType">RisingWaveMetaStoreBackendType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreStatus">RisingWaveMetaStoreStatus</a>)
</p>
<div>
<p>RisingWaveMetaStoreBackendType is the type for the meta store backends.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Etcd&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Memory&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unknown&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreStatus">RisingWaveMetaStoreStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus</a>)
</p>
<div>
<p>RisingWaveMetaStoreStatus is the status of the meta store.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>backend</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendType">
RisingWaveMetaStoreBackendType
</a>
</em>
</td>
<td>
<p>Backend type of the meta store.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMinIOCredentials">RisingWaveMinIOCredentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendMinIO">RisingWaveStateStoreBackendMinIO</a>)
</p>
<div>
<p>RisingWaveMinIOCredentials is the reference and keys selector to the MinIO access credentials stored in a local secret.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The name of the secret in the pod&rsquo;s namespace to select from.</p>
</td>
</tr>
<tr>
<td>
<code>usernameKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>UsernameKeyRef is the key of the secret to be the username. Must be a valid secret key.
Defaults to &ldquo;username&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>passwordKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>PasswordKeyRef is the key of the secret to be the password. Must be a valid secret key.
Defaults to &ldquo;password&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveRollingUpdate">RisingWaveRollingUpdate
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategy">RisingWaveUpgradeStrategy</a>)
</p>
<div>
<p>RisingWaveRollingUpdate is the spec to define rolling update strategies.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>maxUnavailable</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr#IntOrString">
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The maximum number of pods that can be unavailable during the update.
Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
Absolute number is calculated from percentage by rounding down.
Defaults to 25%.</p>
</td>
</tr>
<tr>
<td>
<code>partition</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr#IntOrString">
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Partition is the desired number of pods in old revisions.
Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
Absolute number is calculated from percentage by rounding up by default.
It means when partition is set during pods updating, (replicas - partition value) number of pods will be updated.
Default value is 0.</p>
</td>
</tr>
<tr>
<td>
<code>maxSurge</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/util/intstr#IntOrString">
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The maximum number of pods that can be scheduled above the desired replicas during update or specified delete.
Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
Absolute number is calculated from percentage by rounding up.
Defaults to 0.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveS3Credentials">RisingWaveS3Credentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendAliyunOSS">RisingWaveStateStoreBackendAliyunOSS</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendS3">RisingWaveStateStoreBackendS3</a>)
</p>
<div>
<p>RisingWaveS3Credentials is the reference and keys selector to the AWS access credentials stored in a local secret.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>secretName</code><br/>
<em>
string
</em>
</td>
<td>
<p>The name of the secret in the pod&rsquo;s namespace to select from.</p>
</td>
</tr>
<tr>
<td>
<code>accessKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccessKeyRef is the key of the secret to be the access key. Must be a valid secret key.
Defaults to &ldquo;AccessKeyID&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>secretAccessKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>SecretAccessKeyRef is the key of the secret to be the secret access key. Must be a valid secret key.
Defaults to &ldquo;SecretAccessKey&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleView">RisingWaveScaleView
</h3>
<div>
<p>RisingWaveScaleView is the struct for RisingWaveScaleView.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpec">
RisingWaveScaleViewSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>targetRef</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewTargetRef">
RisingWaveScaleViewTargetRef
</a>
</em>
</td>
<td>
<p>Reference of the target RisingWave.</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Desired replicas.</p>
</td>
</tr>
<tr>
<td>
<code>labelSelector</code><br/>
<em>
string
</em>
</td>
<td>
<p>Serialized label selector. Would be set by the webhook.</p>
</td>
</tr>
<tr>
<td>
<code>scalePolicy</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpecScalePolicy">
[]RisingWaveScaleViewSpecScalePolicy
</a>
</em>
</td>
<td>
<p>An array of groups and the policies for scale, optional and empty means the default group with the default policy.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewStatus">
RisingWaveScaleViewStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewLock">RisingWaveScaleViewLock
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus</a>)
</p>
<div>
<p>RisingWaveScaleViewLock is a lock record for RisingWaveScaleViews. For example, if there&rsquo;s a RisingWaveScaleView
targets the current RisingWave, the controller will try to create a new RisingWaveScaleViewLock with the name, uid,
target component, generation, and the replicas of targeting groups of the RisingWaveScaleView. After the record is set,
the validation webhook will reject any updates on the replicas of any targeting group that doesn&rsquo;t equal the
replicas recorded, which makes it a lock similar thing.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the owned RisingWaveScaleView object.</p>
</td>
</tr>
<tr>
<td>
<code>uid</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/types#UID">
k8s.io/apimachinery/pkg/types.UID
</a>
</em>
</td>
<td>
<p>UID of the owned RisingWaveScaleView object.</p>
</td>
</tr>
<tr>
<td>
<code>component</code><br/>
<em>
string
</em>
</td>
<td>
<p>Component of the lock.</p>
</td>
</tr>
<tr>
<td>
<code>generation</code><br/>
<em>
int64
</em>
</td>
<td>
<p>Generation of the lock.</p>
</td>
</tr>
<tr>
<td>
<code>groupLocks</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewLockGroupLock">
[]RisingWaveScaleViewLockGroupLock
</a>
</em>
</td>
<td>
<p>Group locks.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewLockGroupLock">RisingWaveScaleViewLockGroupLock
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewLock">RisingWaveScaleViewLock</a>)
</p>
<div>
<p>RisingWaveScaleViewLockGroupLock is the lock record of RisingWaveScaleView.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Group name.</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Locked replica value.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpec">RisingWaveScaleViewSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleView">RisingWaveScaleView</a>)
</p>
<div>
<p>RisingWaveScaleViewSpec is the spec of RisingWaveScaleView.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>targetRef</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewTargetRef">
RisingWaveScaleViewTargetRef
</a>
</em>
</td>
<td>
<p>Reference of the target RisingWave.</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Desired replicas.</p>
</td>
</tr>
<tr>
<td>
<code>labelSelector</code><br/>
<em>
string
</em>
</td>
<td>
<p>Serialized label selector. Would be set by the webhook.</p>
</td>
</tr>
<tr>
<td>
<code>scalePolicy</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpecScalePolicy">
[]RisingWaveScaleViewSpecScalePolicy
</a>
</em>
</td>
<td>
<p>An array of groups and the policies for scale, optional and empty means the default group with the default policy.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpecScalePolicy">RisingWaveScaleViewSpecScalePolicy
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpec">RisingWaveScaleViewSpec</a>)
</p>
<div>
<p>RisingWaveScaleViewSpecScalePolicy is the scale policy of a group.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>group</code><br/>
<em>
string
</em>
</td>
<td>
<p>Group name.</p>
</td>
</tr>
<tr>
<td>
<code>priority</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>0-10, optional. The groups will be sorted by the priority and the current replicas.
The higher it is, the more replicas of the target group will be considered kept, i.e. scale out first, scale in last.</p>
</td>
</tr>
<tr>
<td>
<code>maxReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>MaxReplicas is the limit of the replicas.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpecScalePolicyConstraints">RisingWaveScaleViewSpecScalePolicyConstraints
</h3>
<div>
<p>RisingWaveScaleViewSpecScalePolicyConstraints is the constraints of replicas in scale policy.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>max</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Maximum value of the replicas.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewStatus">RisingWaveScaleViewStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleView">RisingWaveScaleView</a>)
</p>
<div>
<p>RisingWaveScaleViewStatus is the status of RisingWaveScaleView.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>replicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Running replicas.</p>
</td>
</tr>
<tr>
<td>
<code>locked</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Lock status.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewTargetRef">RisingWaveScaleViewTargetRef
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewSpec">RisingWaveScaleViewSpec</a>)
</p>
<div>
<p>RisingWaveScaleViewTargetRef is the reference of the target RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name of the RisingWave object.</p>
</td>
</tr>
<tr>
<td>
<code>component</code><br/>
<em>
string
</em>
</td>
<td>
<p>Component name. Must be one of meta, frontend, compute, and compactor.</p>
</td>
</tr>
<tr>
<td>
<code>uid</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/types#UID">
k8s.io/apimachinery/pkg/types.UID
</a>
</em>
</td>
<td>
<p>UID of the target RisingWave object. Should be set by the mutating webhook.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWave">RisingWave</a>)
</p>
<div>
<p>RisingWaveSpec is the overall spec.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>global</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalSpec">
RisingWaveGlobalSpec
</a>
</em>
</td>
<td>
<p>The spec of a shared template for components and a global scope of replicas.</p>
</td>
</tr>
<tr>
<td>
<code>storages</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">
RisingWaveStoragesSpec
</a>
</em>
</td>
<td>
<p>The spec of meta storage, object storage for compute and compactor, and PVC templates for compute.</p>
</td>
</tr>
<tr>
<td>
<code>components</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">
RisingWaveComponentsSpec
</a>
</em>
</td>
<td>
<p>The spec of ports and some controllers (such as <code>restartAt</code>) of each component,
as well as an advanced concept called <code>group</code> to override the global template and create groups
of Pods, e.g., deployment in hybrid-arch cluster.</p>
</td>
</tr>
<tr>
<td>
<code>configuration</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveConfigurationSpec">
RisingWaveConfigurationSpec
</a>
</em>
</td>
<td>
<p>The spec of configuration template for RisingWave.</p>
</td>
</tr>
<tr>
<td>
<code>enableOpenKruise</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to indicate if OpenKruise should be enabled for components.
If enabled, CloneSets will be used for meta/frontend/compactor nodes
and Advanced StateFulSets will be used for compute nodes.</p>
</td>
</tr>
<tr>
<td>
<code>enableDefaultServiceMonitor</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to indicate if a default ServiceMonitor (from Prometheus operator) should be created by the controller.
False and an empty value means the ServiceMonitor won&rsquo;t be created automatically. But even if it&rsquo;s set to true,
the controller will determine if it can create the resource by checking if the CRDs are installed.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Image for RisingWave component.</p>
</td>
</tr>
<tr>
<td>
<code>frontendServiceType</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#servicetype-v1-core">
Kubernetes core/v1.ServiceType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>FrontendServiceType determines the service type of the frontend service. Defaults to ClusterIP.</p>
</td>
</tr>
<tr>
<td>
<code>additionalFrontendServiceMetadata</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">
PartialObjectMeta
</a>
</em>
</td>
<td>
<p>AdditionalFrontendServiceMetadata tells the operator to add the specified metadata onto the frontend Service.
Note that the system reserved labels and annotations are not valid and will be rejected by the webhook.</p>
</td>
</tr>
<tr>
<td>
<code>metaStore</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">
RisingWaveMetaStoreBackend
</a>
</em>
</td>
<td>
<p>MetaStore determines which backend the meta store will use and the parameters for it. Defaults to memory.
But keep in mind that memory backend is not recommended in production.</p>
</td>
</tr>
<tr>
<td>
<code>stateStore</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">
RisingWaveStateStoreBackend
</a>
</em>
</td>
<td>
<p>StateStore determines which backend the state store will use and the parameters for it. Defaults to memory.
But keep in mind that memory backend is not recommended in production.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">RisingWaveStoragesSpec</a>)
</p>
<div>
<p>RisingWaveStateStoreBackend is the collection of parameters for the state store that RisingWave uses. Note that one
and only one of the first-level fields could be set.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>dataDirectory</code><br/>
<em>
string
</em>
</td>
<td>
<p>DataDirectory is the directory to store the data in the object storage. It is an optional field.</p>
</td>
</tr>
<tr>
<td>
<code>memory</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Memory indicates to store the data in memory. It&rsquo;s only for test usage and strongly discouraged to
be used in production.</p>
</td>
</tr>
<tr>
<td>
<code>minio</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendMinIO">
RisingWaveStateStoreBackendMinIO
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MinIO storage spec.</p>
</td>
</tr>
<tr>
<td>
<code>s3</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendS3">
RisingWaveStateStoreBackendS3
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>S3 storage spec.</p>
</td>
</tr>
<tr>
<td>
<code>GCS</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendGCS">
RisingWaveStateStoreBackendGCS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>GCS storage spec.</p>
</td>
</tr>
<tr>
<td>
<code>aliyunOSS</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendAliyunOSS">
RisingWaveStateStoreBackendAliyunOSS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AliyunOSS storage spec.</p>
</td>
</tr>
<tr>
<td>
<code>azureBlob</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendAzureBlob">
RisingWaveStateStoreBackendAzureBlob
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Azure Blob storage spec.</p>
</td>
</tr>
<tr>
<td>
<code>hdfs</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendHDFS">
RisingWaveStateStoreBackendHDFS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HDFS storage spec.</p>
</td>
</tr>
<tr>
<td>
<code>webhdfs</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendHDFS">
RisingWaveStateStoreBackendHDFS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>WebHDFS storage spec.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendAliyunOSS">RisingWaveStateStoreBackendAliyunOSS
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendAliyunOSS is the details of Aliyun OSS storage (S3 compatible) for compute and compactor components.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveS3Credentials">
RisingWaveS3Credentials
</a>
</em>
</td>
<td>
<p>RisingWaveS3Credentials is the credentials provider from a Secret.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the Aliyun OSS service. It must contain the following keys:
* AccessKeyID
* SecretAccessKey
* Region (optional if region is specified in the field.)
Deprecated: Please use &ldquo;credentials&rdquo; field instead. The &ldquo;Secret&rdquo; field will be removed in a future release.</p>
</td>
</tr>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region of Aliyun OSS service. It is an optional field that overrides the <code>Region</code> key from the secret.
Specifying the region here makes a guarantee that it won&rsquo;t be changed anymore.</p>
</td>
</tr>
<tr>
<td>
<code>bucket</code><br/>
<em>
string
</em>
</td>
<td>
<p>Bucket of the Aliyun OSS service.</p>
</td>
</tr>
<tr>
<td>
<code>internalEndpoint</code><br/>
<em>
bool
</em>
</td>
<td>
<p>InternalEndpoint indicates if we use the internal endpoint to access Aliyun OSS, which is
only available in the internal network.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendAzureBlob">RisingWaveStateStoreBackendAzureBlob
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendAzureBlob is the details of Azure blob storage (S3 compatible) for compute and compactor components.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveAzureBlobCredentials">
RisingWaveAzureBlobCredentials
</a>
</em>
</td>
<td>
<p>RisingWaveAzureBlobCredentials is the credentials provider from a Secret.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the Azure Blob service. It must contain the following keys:
* AccessKeyID
* SecretAccessKey
Deprecated: Please use &ldquo;credentials&rdquo; field instead. The &ldquo;Secret&rdquo; field will be removed in a future release.</p>
</td>
</tr>
<tr>
<td>
<code>container</code><br/>
<em>
string
</em>
</td>
<td>
<p>Container Name of the Azure Blob service.</p>
</td>
</tr>
<tr>
<td>
<code>root</code><br/>
<em>
string
</em>
</td>
<td>
<p>Working directory root of the Azure Blob service.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<p>Endpoint of the Azure Blob service.
e.g. <a href="https://yufantest.blob.core.windows.net">https://yufantest.blob.core.windows.net</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendGCS">RisingWaveStateStoreBackendGCS
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendGCS is the collection of parameters for the GCS backend state store.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGCSCredentials">
RisingWaveGCSCredentials
</a>
</em>
</td>
<td>
<p>RisingWaveGCSCredentials is the credentials provider from a Secret.</p>
</td>
</tr>
<tr>
<td>
<code>useWorkloadIdentity</code><br/>
<em>
bool
</em>
</td>
<td>
<p>UseWorkloadIdentity indicates to use workload identity to access the GCS service. If this is enabled, secret is not required, and ADC is used.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the GCS service. It must contain the following keys:
* ServiceAccountCredentials
Deprecated: Please use &ldquo;credentials&rdquo; field instead. The &ldquo;Secret&rdquo; field will be removed in a future release.</p>
</td>
</tr>
<tr>
<td>
<code>bucket</code><br/>
<em>
string
</em>
</td>
<td>
<p>Bucket of the GCS bucket service.</p>
</td>
</tr>
<tr>
<td>
<code>root</code><br/>
<em>
string
</em>
</td>
<td>
<p>Working directory root of the GCS bucket</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendHDFS">RisingWaveStateStoreBackendHDFS
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendHDFS is the details of HDFS storage (S3 compatible) for compute and compactor components.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>nameNode</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name node of the HDFS</p>
</td>
</tr>
<tr>
<td>
<code>root</code><br/>
<em>
string
</em>
</td>
<td>
<p>Working directory root of the HDFS</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendMinIO">RisingWaveStateStoreBackendMinIO
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendMinIO is the collection of parameters for the MinIO backend state store.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMinIOCredentials">
RisingWaveMinIOCredentials
</a>
</em>
</td>
<td>
<p>RisingWaveMinIOCredentials is the credentials provider from a Secret.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the MinIO service. It must contain the following keys:
* username
* password
Deprecated: Please use &ldquo;credentials&rdquo; field instead. The &ldquo;Secret&rdquo; field will be removed in a future release.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<p>Endpoint of the MinIO service.</p>
</td>
</tr>
<tr>
<td>
<code>bucket</code><br/>
<em>
string
</em>
</td>
<td>
<p>Bucket of the MinIO service.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendS3">RisingWaveStateStoreBackendS3
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendS3 is the collection of parameters for the S3 backend state store.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>credentials</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveS3Credentials">
RisingWaveS3Credentials
</a>
</em>
</td>
<td>
<p>RisingWaveS3Credentials is the credentials provider from a Secret.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the AWS S3 service. It must contain the following keys:
* AccessKeyID
* SecretAccessKey
* Region (optional if region is specified in the field.)
Deprecated: Please use &ldquo;credentials&rdquo; field instead. The &ldquo;Secret&rdquo; field will be removed in a future release.</p>
</td>
</tr>
<tr>
<td>
<code>bucket</code><br/>
<em>
string
</em>
</td>
<td>
<p>Bucket of the AWS S3 service.</p>
</td>
</tr>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region of AWS S3 service. It is an optional field that overrides the <code>Region</code> key from the secret.
Specifying the region here makes a guarantee that it won&rsquo;t be changed anymore.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Endpoint of the AWS (or other vendor&rsquo;s S3-compatible) service. Leave it empty when using AWS S3 service.
You can reference the <code>REGION</code> and <code>BUCKET</code> variables in the endpoint with <code>${BUCKET}</code> and <code>${REGION}</code>, e.g.,
s3.${REGION}.amazonaws.com
${BUCKET}.s3.${REGION}.amazonaws.com</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendType">RisingWaveStateStoreBackendType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreStatus">RisingWaveStateStoreStatus</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendType is the type for the state store backends.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;AliyunOSS&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;AzureBlob&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;GCS&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;HDFS&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Memory&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;MinIO&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;S3&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unknown&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;WebHDFS&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreStatus">RisingWaveStateStoreStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus</a>)
</p>
<div>
<p>RisingWaveStateStoreStatus is the status of the state store.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>backend</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendType">
RisingWaveStateStoreBackendType
</a>
</em>
</td>
<td>
<p>Backend type of the state store.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWave">RisingWave</a>)
</p>
<div>
<p>RisingWaveStatus is the status of RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>observedGeneration</code><br/>
<em>
int64
</em>
</td>
<td>
<p>Observed generation by controller. It will be updated
when controller observes the changes on the spec and going to sync the subresources.</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version of the Global Image</p>
</td>
</tr>
<tr>
<td>
<code>componentReplicas</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsReplicasStatus">
RisingWaveComponentsReplicasStatus
</a>
</em>
</td>
<td>
<p>Replica status of components.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveCondition">
[]RisingWaveCondition
</a>
</em>
</td>
<td>
<p>Conditions of the RisingWave.</p>
</td>
</tr>
<tr>
<td>
<code>scaleViews</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleViewLock">
[]RisingWaveScaleViewLock
</a>
</em>
</td>
<td>
<p>Scale view locks.</p>
</td>
</tr>
<tr>
<td>
<code>metaStore</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreStatus">
RisingWaveMetaStoreStatus
</a>
</em>
</td>
<td>
<p>Status of the meta store.</p>
</td>
</tr>
<tr>
<td>
<code>stateStore</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreStatus">
RisingWaveStateStoreStatus
</a>
</em>
</td>
<td>
<p>Status of the state store.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">RisingWaveStoragesSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
</p>
<div>
<p>RisingWaveStoragesSpec is the storages spec.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>meta</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">
RisingWaveMetaStoreBackend
</a>
</em>
</td>
<td>
<p>Storage spec for meta.</p>
</td>
</tr>
<tr>
<td>
<code>object</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">
RisingWaveStateStoreBackend
</a>
</em>
</td>
<td>
<p>Storage spec for object storage.</p>
</td>
</tr>
<tr>
<td>
<code>pvcTemplates</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PersistentVolumeClaim">
[]PersistentVolumeClaim
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>The persistent volume claim templates for the compute component. PVCs declared here
can be referenced in the groups of compute component.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategy">RisingWaveUpgradeStrategy
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentGroupTemplate">RisingWaveComponentGroupTemplate</a>)
</p>
<div>
<p>RisingWaveUpgradeStrategy is the spec of upgrade strategy used by RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategyType">
RisingWaveUpgradeStrategyType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type of upgrade. Can be &ldquo;Recreate&rdquo; or &ldquo;RollingUpdate&rdquo;. Default is RollingUpdate.</p>
</td>
</tr>
<tr>
<td>
<code>rollingUpdate</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveRollingUpdate">
RisingWaveRollingUpdate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Rolling update config params. Present only if DeploymentStrategyType = RollingUpdate.</p>
</td>
</tr>
<tr>
<td>
<code>inPlaceUpdateStrategy</code><br/>
<em>
github.com/openkruise/kruise-api/apps/pub.InPlaceUpdateStrategy
</em>
</td>
<td>
<em>(Optional)</em>
<p>InPlaceUpdateStrategy contains strategies for in-place update.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategyType">RisingWaveUpgradeStrategyType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveUpgradeStrategy">RisingWaveUpgradeStrategy</a>)
</p>
<div>
<p>RisingWaveUpgradeStrategyType is the type of upgrade strategies used in RisingWave.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;InPlaceIfPossible&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;InPlaceOnly&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Recreate&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;RollingUpdate&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<hr/>
