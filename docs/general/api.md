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
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplate">RisingWaveNodePodTemplate</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
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
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroup">RisingWaveNodeGroup</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStandaloneComponent">RisingWaveStandaloneComponent</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumeresourcerequirements-v1-core">
Kubernetes core/v1.VolumeResourceRequirements
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
<tr>
<td>
<code>volumeAttributesClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeAttributesClassName may be used to set the VolumeAttributesClass used by this claim.
If specified, the CSI driver will create or update the volume with the attributes defined
in the corresponding VolumeAttributesClass. This has a different purpose than storageClassName,
it can be changed after the claim is created. An empty string value means that no VolumeAttributesClass
will be applied to the claim but it&rsquo;s not allowed to reset this field to empty string once it is set.
If unspecified and the PersistentVolumeClaim is unbound, the default VolumeAttributesClass
will be set by the persistentvolume controller if it exists.
If the resource referred to by volumeAttributesClass does not exist, this PersistentVolumeClaim will be
set to a Pending state, as reflected by the modifyVolumeStatus field, until such as a resource
exists.
More info: <a href="https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/">https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/</a>
(Beta) Using this field requires the VolumeAttributesClass feature gate to be enabled (off by default).</p>
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
<code>enableFullKubernetesAddr</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to indicate if full kubernetes address should be enabled for components.
If enabled, address will be [<pod>.]<service>.<namespace>.svc. Otherwise, it will be [<pod>.]<service>.
Enabling this flag on existing RisingWave will cause incompatibility.</p>
</td>
</tr>
<tr>
<td>
<code>enableStandaloneMode</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to control whether to deploy in standalone mode or distributed mode. If standalone mode is used,
spec.components will be ignored. Standalone mode can be turned on/off dynamically.</p>
</td>
</tr>
<tr>
<td>
<code>enableEmbeddedServingMode</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to control whether to enable embedded serving mode. If enabled, the frontend nodes will be created
with embedded serving node enabled, and the compute nodes will serve streaming workload only.</p>
</td>
</tr>
<tr>
<td>
<code>enableAdvertisingWithIP</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Flag to control whether to enable advertising with IP. If enabled, the meta and compute nodes will be advertised
with their IP addresses. This is useful when one wants to avoid the DNS resolution overhead and latency.</p>
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
<tr>
<td>
<code>tls</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfiguration">
RisingWaveTLSConfiguration
</a>
</em>
</td>
<td>
<p>TLS configures the TLS/SSL certificates for SQL access.</p>
</td>
</tr>
<tr>
<td>
<code>standaloneMode</code><br/>
<em>
int32
</em>
</td>
<td>
<p>StandaloneMode determines which style of command-line args should be used for the standalone mode.
0 - auto detect by image version, 1 - the old standalone mode, 2 - standalone mode V2 (single-node).
This is only for backward compatibility and will be deprecated in the future.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveAliyunOSSCredentials">RisingWaveAliyunOSSCredentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendAliyunOSS">RisingWaveStateStoreBackendAliyunOSS</a>)
</p>
<div>
<p>RisingWaveAliyunOSSCredentials is the reference and keys selector to the AliyunOSS access credentials stored in a local secret.</p>
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
<code>accessKeyIDRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccessKeyIDRef is the key of the secret to be the access key. Must be a valid secret key.
Defaults to &ldquo;AccessKeyIDRef&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>accessKeySecretRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccessKeySecretRef is the key of the secret to be the secret access key. Must be a valid secret key.
Defaults to &ldquo;AccessKeySecretRef&rdquo;.</p>
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
<code>accountKeyRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccountKeyRef is the key of the secret to be the secret account key. Must be a valid secret key.
Defaults to &ldquo;AccountKey&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>useServiceAccount</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>UseServiceAccount indicates whether to use the service account token mounted in the pod.
If this is enabled, secret and keys are ignored. Defaults to false.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponent">RisingWaveComponent
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
<p>RisingWaveComponent determines how a RisingWave component is deployed.</p>
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
<code>logLevel</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogLevel controls the log level of the running nodes. It can be in any format that the underlying component supports,
e.g., in the RUST_LOG format for Rust programs. Defaults to INFO.</p>
</td>
</tr>
<tr>
<td>
<code>disallowPrintStackTraces</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DisallowPrintStackTraces determines if the stack traces are allowed to print when panic happens. This options applies
to both Rust and Java programs. Defaults to false.</p>
</td>
</tr>
<tr>
<td>
<code>nodeGroups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroup">
[]RisingWaveNodeGroup
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeGroups of the component deployment.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentStatus">RisingWaveComponentStatus
</h3>
<div>
<p>RisingWaveComponentStatus is the status of a component.</p>
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
<code>total</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.WorkloadReplicaStatus">
WorkloadReplicaStatus
</a>
</em>
</td>
<td>
<p>Total is the replica status of the component.</p>
</td>
</tr>
<tr>
<td>
<code>nodeGroups</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupStatus">
[]RisingWaveNodeGroupStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeGroups are the status list of all declared node groups of some component.</p>
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
<code>standalone</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.ComponentReplicasStatus">
ComponentReplicasStatus
</a>
</em>
</td>
<td>
<p>Running status of standalone component.</p>
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
<p>RisingWaveComponentsSpec is the spec for RisingWave components.</p>
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
<code>standalone</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStandaloneComponent">
RisingWaveStandaloneComponent
</a>
</em>
</td>
<td>
<p>Standalone contains configuration of the standalone component.</p>
</td>
</tr>
<tr>
<td>
<code>meta</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponent">
RisingWaveComponent
</a>
</em>
</td>
<td>
<p>Meta contains configuration of the meta component.</p>
</td>
</tr>
<tr>
<td>
<code>frontend</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponent">
RisingWaveComponent
</a>
</em>
</td>
<td>
<p>Frontend contains configuration of the frontend component.</p>
</td>
</tr>
<tr>
<td>
<code>compute</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponent">
RisingWaveComponent
</a>
</em>
</td>
<td>
<p>Compute contains configuration of the compute component.</p>
</td>
</tr>
<tr>
<td>
<code>compactor</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponent">
RisingWaveComponent
</a>
</em>
</td>
<td>
<p>Compactor contains configuration of the compactor component.</p>
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
<code>configMap</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfigurationConfigMapSource">
RisingWaveNodeConfigurationConfigMapSource
</a>
</em>
</td>
<td>
<p>ConfigMap where the <code>risingwave.toml</code> locates.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfigurationSecretSource">
RisingWaveNodeConfigurationSecretSource
</a>
</em>
</td>
<td>
<p>Secret where the <code>risingwave.toml</code> locates.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveDBCredentials">RisingWaveDBCredentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendMySQL">RisingWaveMetaStoreBackendMySQL</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendPostgreSQL">RisingWaveMetaStoreBackendPostgreSQL</a>)
</p>
<div>
<p>RisingWaveDBCredentials is the reference and keys selector to the DB access credentials stored in a local secret.</p>
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
<code>useWorkloadIdentity</code><br/>
<em>
bool
</em>
</td>
<td>
<p>UseWorkloadIdentity indicates to use workload identity to access the GCS service.
If this is enabled, secret is not required, and ADC is used.</p>
</td>
</tr>
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
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveHuaweiCloudOBSCredentials">RisingWaveHuaweiCloudOBSCredentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendHuaweiCloudOBS">RisingWaveStateStoreBackendHuaweiCloudOBS</a>)
</p>
<div>
<p>RisingWaveHuaweiCloudOBSCredentials is the reference and keys selector to the HuaweiCloudOBS access credentials stored in a local secret.</p>
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
<code>accessKeyIDRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccessKeyIDRef is the key of the secret to be the access key. Must be a valid secret key.
Defaults to &ldquo;AccessKeyIDRef&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>accessKeySecretRef</code><br/>
<em>
string
</em>
</td>
<td>
<p>AccessKeySecretRef is the key of the secret to be the secret access key. Must be a valid secret key.
Defaults to &ldquo;AccessKeySecretRef&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveInternalStatus">RisingWaveInternalStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus</a>)
</p>
<div>
<p>RisingWaveInternalStatus stores some internal status of RisingWave, such as internal states.</p>
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
<code>stateStoreRootPath</code><br/>
<em>
string
</em>
</td>
<td>
<p>StateStoreRootPath stores the root path of the state store data directory. It&rsquo;s for compatibility purpose and
should not be updated in most cases.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">RisingWaveMetaStoreBackend
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
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
<p>Stores metadata in etcd.</p>
</td>
</tr>
<tr>
<td>
<code>sqlite</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendSQLite">
RisingWaveMetaStoreBackendSQLite
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SQLite stores metadata in a SQLite DB file.</p>
</td>
</tr>
<tr>
<td>
<code>mysql</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendMySQL">
RisingWaveMetaStoreBackendMySQL
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MySQL stores metadata in a MySQL DB.</p>
</td>
</tr>
<tr>
<td>
<code>postgresql</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendPostgreSQL">
RisingWaveMetaStoreBackendPostgreSQL
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PostgreSQL stores metadata in a PostgreSQL DB.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendMySQL">RisingWaveMetaStoreBackendMySQL
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">RisingWaveMetaStoreBackend</a>)
</p>
<div>
<p>RisingWaveMetaStoreBackendMySQL describes the options of MySQL DB backend.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveDBCredentials">
RisingWaveDBCredentials
</a>
</em>
</td>
<td>
<p>RisingWaveDBCredentials is the reference credentials. User must provide a secret contains
<code>username</code> and <code>password</code> (or one can customize the key references) keys and the correct values.</p>
</td>
</tr>
<tr>
<td>
<code>host</code><br/>
<em>
string
</em>
</td>
<td>
<p>Host of the MySQL DB.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
uint32
</em>
</td>
<td>
<p>Port of the MySQL DB. Defaults to 3306.</p>
</td>
</tr>
<tr>
<td>
<code>database</code><br/>
<em>
string
</em>
</td>
<td>
<p>Database of the MySQL DB.</p>
</td>
</tr>
<tr>
<td>
<code>options</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Options when connecting to the MySQL DB. Optional.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendPostgreSQL">RisingWaveMetaStoreBackendPostgreSQL
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">RisingWaveMetaStoreBackend</a>)
</p>
<div>
<p>RisingWaveMetaStoreBackendPostgreSQL describes the options of PostgreSQL DB backend.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveDBCredentials">
RisingWaveDBCredentials
</a>
</em>
</td>
<td>
<p>RisingWaveDBCredentials is the reference credentials. User must provide a secret contains
<code>username</code> and <code>password</code> (or one can customize the key references) keys and the correct values.</p>
</td>
</tr>
<tr>
<td>
<code>host</code><br/>
<em>
string
</em>
</td>
<td>
<p>Host of the PostgreSQL DB.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
uint32
</em>
</td>
<td>
<p>Port of the PostgreSQL DB. Defaults to 5432.</p>
</td>
</tr>
<tr>
<td>
<code>database</code><br/>
<em>
string
</em>
</td>
<td>
<p>Database of the PostgreSQL DB.</p>
</td>
</tr>
<tr>
<td>
<code>options</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Options when connecting to the PostgreSQL DB. Optional.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackendSQLite">RisingWaveMetaStoreBackendSQLite
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStoreBackend">RisingWaveMetaStoreBackend</a>)
</p>
<div>
<p>RisingWaveMetaStoreBackendSQLite describes the options of SQLite DB backend.</p>
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
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<p>Path of the DB file.</p>
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
</tr><tr><td><p>&#34;MySQL&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;PostgreSQL&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;SQLite&#34;</p></td>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfiguration">RisingWaveNodeConfiguration
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveConfigurationSpec">RisingWaveConfigurationSpec</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroup">RisingWaveNodeGroup</a>)
</p>
<div>
<p>RisingWaveNodeConfiguration determines where the configurations are from, either ConfigMap, Secret, or raw string.</p>
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
<code>configMap</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfigurationConfigMapSource">
RisingWaveNodeConfigurationConfigMapSource
</a>
</em>
</td>
<td>
<p>ConfigMap where the <code>risingwave.toml</code> locates.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfigurationSecretSource">
RisingWaveNodeConfigurationSecretSource
</a>
</em>
</td>
<td>
<p>Secret where the <code>risingwave.toml</code> locates.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfigurationConfigMapSource">RisingWaveNodeConfigurationConfigMapSource
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfiguration">RisingWaveNodeConfiguration</a>)
</p>
<div>
<p>RisingWaveNodeConfigurationConfigMapSource refers to a ConfigMap where the RisingWave configuration is stored.</p>
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
<p>Name determines the ConfigMap to provide the configs RisingWave requests. It will be mounted on the Pods
directly. It the ConfigMap isn&rsquo;t provided, the controller will use empty value as the configs.</p>
</td>
</tr>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Key to the configuration file. Defaults to <code>risingwave.toml</code>.</p>
</td>
</tr>
<tr>
<td>
<code>optional</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional determines if the key must exist in the ConfigMap. Defaults to false.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfigurationSecretSource">RisingWaveNodeConfigurationSecretSource
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfiguration">RisingWaveNodeConfiguration</a>)
</p>
<div>
<p>RisingWaveNodeConfigurationSecretSource refers to a Secret where the RisingWave configuration is stored.</p>
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
<p>Name determines the Secret to provide the configs RisingWave requests. It will be mounted on the Pods
directly. It the Secret isn&rsquo;t provided, the controller will use empty value as the configs.</p>
</td>
</tr>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Key to the configuration file. Defaults to <code>risingwave.toml</code>.</p>
</td>
</tr>
<tr>
<td>
<code>optional</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional determines if the key must exist in the Secret. Defaults to false.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeContainer">RisingWaveNodeContainer
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplateSpec">RisingWaveNodePodTemplateSpec</a>)
</p>
<div>
<p>RisingWaveNodeContainer determines the container specs of a RisingWave node.</p>
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
<p>Container image name.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images">https://kubernetes.io/docs/concepts/containers/images</a>
This field is optional to allow higher level config management to default or override
container images in workload controllers like Deployments and StatefulSets.</p>
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
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Compute Resources required by this container.
Cannot be updated.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
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
<p>Pod volumes to mount into the container&rsquo;s filesystem.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>volumeDevices</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumedevice-v1-core">
[]Kubernetes core/v1.VolumeDevice
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeDevices is the list of block devices to be used by the container.</p>
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
<p>Image pull policy.
One of Always, Never, IfNotPresent.
Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
Cannot be updated.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#updating-images">https://kubernetes.io/docs/concepts/containers/images#updating-images</a></p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext defines the security options the container should be run with.
If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.
More info: <a href="https://kubernetes.io/docs/tasks/configure-pod-container/security-context/">https://kubernetes.io/docs/tasks/configure-pod-container/security-context/</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroup">RisingWaveNodeGroup
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponent">RisingWaveComponent</a>)
</p>
<div>
<p>RisingWaveNodeGroup is the definition of a group of RisingWave nodes of the same component.</p>
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
<p>Name of the node group.</p>
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
<p>RestartAt is the time that the Pods under the group should be restarted. Setting a value on this field will
trigger a full recreation of the Pods. Defaults to nil.</p>
</td>
</tr>
<tr>
<td>
<code>configuration</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeConfiguration">
RisingWaveNodeConfiguration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Configuration determines the configuration to be used for the RisingWave nodes.</p>
</td>
</tr>
<tr>
<td>
<code>upgradeStrategy</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupUpgradeStrategy">
RisingWaveNodeGroupUpgradeStrategy
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
<code>minReadySeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)</p>
</td>
</tr>
<tr>
<td>
<code>volumeClaimTemplates</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PersistentVolumeClaim">
[]PersistentVolumeClaim
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeClaimTemplates is a list of claims that pods are allowed to reference.
The StatefulSet controller is responsible for mapping network identities to
claims in a way that maintains the identity of a pod. Every claim in
this list must have at least one matching (by name) volumeMount in one
container in the template. A claim in this list takes precedence over
any volumes in the template, with the same name.</p>
</td>
</tr>
<tr>
<td>
<code>persistentVolumeClaimRetentionPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps">
Kubernetes apps/v1.StatefulSetPersistentVolumeClaimRetentionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>persistentVolumeClaimRetentionPolicy describes the lifecycle of persistent
volume claims created from volumeClaimTemplates. By default, all persistent
volume claims are created as needed and retained until manually deleted. This
policy allows the lifecycle to be altered, for example by deleting persistent
volume claims when their stateful set is deleted, or when their pod is scaled
down. This requires the StatefulSetAutoDeletePVC feature gate to be enabled,
which is alpha.</p>
</td>
</tr>
<tr>
<td>
<code>progressDeadlineSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<p>The maximum time in seconds for a deployment to make progress before it
is considered to be failed. The deployment controller will continue to
process failed deployments and a condition with a ProgressDeadlineExceeded
reason will be surfaced in the deployment status. Note that progress will
not be estimated during the time a deployment is paused. Defaults to 600s.</p>
</td>
</tr>
<tr>
<td>
<code>template</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplate">
RisingWaveNodePodTemplate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Template tells how the Pod should be started. It is an optional field. If it&rsquo;s empty, then the pod template in
the first-level fields under spec will be used.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupRollingUpdate">RisingWaveNodeGroupRollingUpdate
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupUpgradeStrategy">RisingWaveNodeGroupUpgradeStrategy</a>)
</p>
<div>
<p>RisingWaveNodeGroupRollingUpdate is the spec to define rolling update strategies.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupStatus">RisingWaveNodeGroupStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentStatus">RisingWaveComponentStatus</a>)
</p>
<div>
<p>RisingWaveNodeGroupStatus is the status of a node group.</p>
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
<p>Name of the node group.</p>
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
<p>Replicas is the declared replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>readyReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>ReadyReplicas is the ready replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>availableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>AvailableReplicas is the available replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>updatedReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>UpdatedReplicas is the update replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>unavailableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>UnavailableReplicas is the unavailable replicas of the workload.</p>
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
<p>Existence status of the node group.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupUpgradeStrategy">RisingWaveNodeGroupUpgradeStrategy
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroup">RisingWaveNodeGroup</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStandaloneComponent">RisingWaveStandaloneComponent</a>)
</p>
<div>
<p>RisingWaveNodeGroupUpgradeStrategy is the spec of upgrade strategy used by RisingWave.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupUpgradeStrategyType">
RisingWaveNodeGroupUpgradeStrategyType
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupRollingUpdate">
RisingWaveNodeGroupRollingUpdate
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupUpgradeStrategyType">RisingWaveNodeGroupUpgradeStrategyType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupUpgradeStrategy">RisingWaveNodeGroupUpgradeStrategy</a>)
</p>
<div>
<p>RisingWaveNodeGroupUpgradeStrategyType is the type of upgrade strategies used in RisingWave.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplate">RisingWaveNodePodTemplate
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroup">RisingWaveNodeGroup</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStandaloneComponent">RisingWaveStandaloneComponent</a>)
</p>
<div>
<p>RisingWaveNodePodTemplate determines the Pod specs of a RisingWave node.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.PartialObjectMeta">
PartialObjectMeta
</a>
</em>
</td>
<td>
<p>PartialObjectMeta tells the operator to add the specified metadata onto the Pod.</p>
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplateSpec">
RisingWaveNodePodTemplateSpec
</a>
</em>
</td>
<td>
<p>RisingWaveNodePodTemplateSpec determines the Pod spec to start the RisingWave pod.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Container image name.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images">https://kubernetes.io/docs/concepts/containers/images</a>
This field is optional to allow higher level config management to default or override
container images in workload controllers like Deployments and StatefulSets.</p>
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
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Compute Resources required by this container.
Cannot be updated.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
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
<p>Pod volumes to mount into the container&rsquo;s filesystem.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>volumeDevices</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumedevice-v1-core">
[]Kubernetes core/v1.VolumeDevice
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeDevices is the list of block devices to be used by the container.</p>
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
<p>Image pull policy.
One of Always, Never, IfNotPresent.
Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
Cannot be updated.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#updating-images">https://kubernetes.io/docs/concepts/containers/images#updating-images</a></p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext defines the security options the container should be run with.
If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.
More info: <a href="https://kubernetes.io/docs/tasks/configure-pod-container/security-context/">https://kubernetes.io/docs/tasks/configure-pod-container/security-context/</a></p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of volumes that can be mounted by containers belonging to the pod.
More info: <a href="https://kubernetes.io/docs/concepts/storage/volumes">https://kubernetes.io/docs/concepts/storage/volumes</a></p>
</td>
</tr>
<tr>
<td>
<code>activeDeadlineSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the pod may be active on the node relative to
StartTime before the system will actively try to mark it failed and kill associated containers.
Value must be a positive integer.</p>
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
<code>dnsPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#dnspolicy-v1-core">
Kubernetes core/v1.DNSPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Set DNS policy for the pod.
Defaults to &ldquo;ClusterFirst&rdquo;.
Valid values are &lsquo;ClusterFirstWithHostNet&rsquo;, &lsquo;ClusterFirst&rsquo;, &lsquo;Default&rsquo; or &lsquo;None&rsquo;.
DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
To have DNS options set along with hostNetwork, you have to specify DNS policy
explicitly to &lsquo;ClusterFirstWithHostNet&rsquo;.</p>
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
<p>NodeSelector is a selector which must be true for the pod to fit on a node.
Selector which must match a node&rsquo;s labels for the pod to be scheduled on that node.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/assign-pod-node/">https://kubernetes.io/docs/concepts/configuration/assign-pod-node/</a></p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceAccountName is the name of the ServiceAccount to use to run this pod.
More info: <a href="https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/">https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/</a></p>
</td>
</tr>
<tr>
<td>
<code>automountServiceAccountToken</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.</p>
</td>
</tr>
<tr>
<td>
<code>hostPID</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use the host&rsquo;s pid namespace.
Optional: Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>hostIPC</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use the host&rsquo;s ipc namespace.
Optional: Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>shareProcessNamespace</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.
Optional: Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>podSecurityContext</code><br/>
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
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.</p>
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
<code>hostAliases</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#hostalias-v1-core">
[]Kubernetes core/v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HostAliases is an optional list of hosts and IPs that will be injected into the pod&rsquo;s hosts
file if specified. This is only valid for non-hostNetwork pods.</p>
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
<code>priority</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The priority value. Various system components use this field to find the
priority of the pod. When Priority Admission Controller is enabled, it
prevents users from setting this field. The admission controller populates
this field from PriorityClassName.
The higher the value, the higher the priority.</p>
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
<code>runtimeClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used
to run this pod.  If no RuntimeClass resource matches the named class, the pod will not be run.
If unset or empty, the &ldquo;legacy&rdquo; RuntimeClass will be used, which is an implicit class with an
empty definition that uses the default runtime handler.
More info: <a href="https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class">https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class</a></p>
</td>
</tr>
<tr>
<td>
<code>preemptionPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#preemptionpolicy-v1-core">
Kubernetes core/v1.PreemptionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PreemptionPolicy is the Policy for preempting pods with lower priority.
One of Never, PreemptLowerPriority.
Defaults to PreemptLowerPriority if unset.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TopologySpreadConstraints describes how a group of pods ought to spread across topology
domains. Scheduler will schedule pods in a way which abides by the constraints.
All topologySpreadConstraints are ANDed.</p>
</td>
</tr>
<tr>
<td>
<code>setHostnameAsFQDN</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>If true the pod&rsquo;s hostname will be configured as the pod&rsquo;s FQDN, rather than the leaf name (the default).
In Linux containers, this means setting the FQDN in the hostname field of the kernel (the nodename field of struct utsname).
In Windows containers, this means setting the registry value of hostname for the registry key HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters to FQDN.
If a pod does not have FQDN, this has no effect.
Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>os</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podos-v1-core">
Kubernetes core/v1.PodOS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the OS of the containers in the pod.
Some pod and container fields are restricted if this is set.</p>
<p>If the OS field is set to linux, the following fields must be unset:
-securityContext.windowsOptions</p>
<p>If the OS field is set to windows, following fields must be unset:
- spec.hostPID
- spec.hostIPC
- spec.hostUsers
- spec.securityContext.seLinuxOptions
- spec.securityContext.seccompProfile
- spec.securityContext.fsGroup
- spec.securityContext.fsGroupChangePolicy
- spec.securityContext.sysctls
- spec.shareProcessNamespace
- spec.securityContext.runAsUser
- spec.securityContext.runAsGroup
- spec.securityContext.supplementalGroups
- spec.containers[<em>].securityContext.seLinuxOptions
- spec.containers[</em>].securityContext.seccompProfile
- spec.containers[<em>].securityContext.capabilities
- spec.containers[</em>].securityContext.readOnlyRootFilesystem
- spec.containers[<em>].securityContext.privileged
- spec.containers[</em>].securityContext.allowPrivilegeEscalation
- spec.containers[<em>].securityContext.procMount
- spec.containers[</em>].securityContext.runAsUser
- spec.containers[*].securityContext.runAsGroup</p>
</td>
</tr>
<tr>
<td>
<code>hostUsers</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use the host&rsquo;s user namespace.
Optional: Default to true.
If set to true or not present, the pod will be run in the host user namespace, useful
for when the pod needs a feature only available to the host user namespace, such as
loading a kernel module with CAP_SYS_MODULE.
When set to false, a new userns is created for the pod. Setting false is useful for
mitigating container breakout vulnerabilities even allowing users to run their
containers as root without actually having root privileges on the host.
This field is alpha-level and is only honored by servers that enable the UserNamespacesSupport feature.</p>
</td>
</tr>
<tr>
<td>
<code>additionalContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Additional containers to run in the same Pod. The containers will be appended to the Pod&rsquo;s containers array in order.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplateSpec">RisingWaveNodePodTemplateSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplate">RisingWaveNodePodTemplate</a>)
</p>
<div>
<p>RisingWaveNodePodTemplateSpec is a template for a RisingWave&rsquo;s Pod.</p>
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
<p>Container image name.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images">https://kubernetes.io/docs/concepts/containers/images</a>
This field is optional to allow higher level config management to default or override
container images in workload controllers like Deployments and StatefulSets.</p>
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
<code>resources</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Compute Resources required by this container.
Cannot be updated.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/">https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/</a></p>
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
<p>Pod volumes to mount into the container&rsquo;s filesystem.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>volumeDevices</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volumedevice-v1-core">
[]Kubernetes core/v1.VolumeDevice
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeDevices is the list of block devices to be used by the container.</p>
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
<p>Image pull policy.
One of Always, Never, IfNotPresent.
Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
Cannot be updated.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#updating-images">https://kubernetes.io/docs/concepts/containers/images#updating-images</a></p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SecurityContext defines the security options the container should be run with.
If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.
More info: <a href="https://kubernetes.io/docs/tasks/configure-pod-container/security-context/">https://kubernetes.io/docs/tasks/configure-pod-container/security-context/</a></p>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of volumes that can be mounted by containers belonging to the pod.
More info: <a href="https://kubernetes.io/docs/concepts/storage/volumes">https://kubernetes.io/docs/concepts/storage/volumes</a></p>
</td>
</tr>
<tr>
<td>
<code>activeDeadlineSeconds</code><br/>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>Optional duration in seconds the pod may be active on the node relative to
StartTime before the system will actively try to mark it failed and kill associated containers.
Value must be a positive integer.</p>
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
<code>dnsPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#dnspolicy-v1-core">
Kubernetes core/v1.DNSPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Set DNS policy for the pod.
Defaults to &ldquo;ClusterFirst&rdquo;.
Valid values are &lsquo;ClusterFirstWithHostNet&rsquo;, &lsquo;ClusterFirst&rsquo;, &lsquo;Default&rsquo; or &lsquo;None&rsquo;.
DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
To have DNS options set along with hostNetwork, you have to specify DNS policy
explicitly to &lsquo;ClusterFirstWithHostNet&rsquo;.</p>
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
<p>NodeSelector is a selector which must be true for the pod to fit on a node.
Selector which must match a node&rsquo;s labels for the pod to be scheduled on that node.
More info: <a href="https://kubernetes.io/docs/concepts/configuration/assign-pod-node/">https://kubernetes.io/docs/concepts/configuration/assign-pod-node/</a></p>
</td>
</tr>
<tr>
<td>
<code>serviceAccountName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceAccountName is the name of the ServiceAccount to use to run this pod.
More info: <a href="https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/">https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/</a></p>
</td>
</tr>
<tr>
<td>
<code>automountServiceAccountToken</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutomountServiceAccountToken indicates whether a service account token should be automatically mounted.</p>
</td>
</tr>
<tr>
<td>
<code>hostPID</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use the host&rsquo;s pid namespace.
Optional: Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>hostIPC</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use the host&rsquo;s ipc namespace.
Optional: Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>shareProcessNamespace</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from other containers
in the same pod, and the first process in each container will not be assigned PID 1.
HostPID and ShareProcessNamespace cannot both be set.
Optional: Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>podSecurityContext</code><br/>
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
<code>imagePullSecrets</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
If specified, these secrets will be passed to individual puller implementations for them to use.
More info: <a href="https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod">https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod</a></p>
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
<code>schedulerName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the pod will be dispatched by specified scheduler.
If not specified, the pod will be dispatched by default scheduler.</p>
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
<code>hostAliases</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#hostalias-v1-core">
[]Kubernetes core/v1.HostAlias
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HostAliases is an optional list of hosts and IPs that will be injected into the pod&rsquo;s hosts
file if specified. This is only valid for non-hostNetwork pods.</p>
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
<code>priority</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>The priority value. Various system components use this field to find the
priority of the pod. When Priority Admission Controller is enabled, it
prevents users from setting this field. The admission controller populates
this field from PriorityClassName.
The higher the value, the higher the priority.</p>
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
<code>runtimeClassName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>RuntimeClassName refers to a RuntimeClass object in the node.k8s.io group, which should be used
to run this pod.  If no RuntimeClass resource matches the named class, the pod will not be run.
If unset or empty, the &ldquo;legacy&rdquo; RuntimeClass will be used, which is an implicit class with an
empty definition that uses the default runtime handler.
More info: <a href="https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class">https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class</a></p>
</td>
</tr>
<tr>
<td>
<code>preemptionPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#preemptionpolicy-v1-core">
Kubernetes core/v1.PreemptionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PreemptionPolicy is the Policy for preempting pods with lower priority.
One of Never, PreemptLowerPriority.
Defaults to PreemptLowerPriority if unset.</p>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TopologySpreadConstraints describes how a group of pods ought to spread across topology
domains. Scheduler will schedule pods in a way which abides by the constraints.
All topologySpreadConstraints are ANDed.</p>
</td>
</tr>
<tr>
<td>
<code>setHostnameAsFQDN</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>If true the pod&rsquo;s hostname will be configured as the pod&rsquo;s FQDN, rather than the leaf name (the default).
In Linux containers, this means setting the FQDN in the hostname field of the kernel (the nodename field of struct utsname).
In Windows containers, this means setting the registry value of hostname for the registry key HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters to FQDN.
If a pod does not have FQDN, this has no effect.
Default to false.</p>
</td>
</tr>
<tr>
<td>
<code>os</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podos-v1-core">
Kubernetes core/v1.PodOS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the OS of the containers in the pod.
Some pod and container fields are restricted if this is set.</p>
<p>If the OS field is set to linux, the following fields must be unset:
-securityContext.windowsOptions</p>
<p>If the OS field is set to windows, following fields must be unset:
- spec.hostPID
- spec.hostIPC
- spec.hostUsers
- spec.securityContext.seLinuxOptions
- spec.securityContext.seccompProfile
- spec.securityContext.fsGroup
- spec.securityContext.fsGroupChangePolicy
- spec.securityContext.sysctls
- spec.shareProcessNamespace
- spec.securityContext.runAsUser
- spec.securityContext.runAsGroup
- spec.securityContext.supplementalGroups
- spec.containers[<em>].securityContext.seLinuxOptions
- spec.containers[</em>].securityContext.seccompProfile
- spec.containers[<em>].securityContext.capabilities
- spec.containers[</em>].securityContext.readOnlyRootFilesystem
- spec.containers[<em>].securityContext.privileged
- spec.containers[</em>].securityContext.allowPrivilegeEscalation
- spec.containers[<em>].securityContext.procMount
- spec.containers[</em>].securityContext.runAsUser
- spec.containers[*].securityContext.runAsGroup</p>
</td>
</tr>
<tr>
<td>
<code>hostUsers</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Use the host&rsquo;s user namespace.
Optional: Default to true.
If set to true or not present, the pod will be run in the host user namespace, useful
for when the pod needs a feature only available to the host user namespace, such as
loading a kernel module with CAP_SYS_MODULE.
When set to false, a new userns is created for the pod. Setting false is useful for
mitigating container breakout vulnerabilities even allowing users to run their
containers as root without actually having root privileges on the host.
This field is alpha-level and is only honored by servers that enable the UserNamespacesSupport feature.</p>
</td>
</tr>
<tr>
<td>
<code>additionalContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>Additional containers to run in the same Pod. The containers will be appended to the Pod&rsquo;s containers array in order.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveS3Credentials">RisingWaveS3Credentials
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendS3">RisingWaveStateStoreBackendS3</a>)
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
<code>useServiceAccount</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>UseServiceAccount indicates whether to use the service account token mounted in the pod. It only works when using
the AWS S3. If this is enabled, secret and keys are ignored. Defaults to false.</p>
</td>
</tr>
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
<code>enableFullKubernetesAddr</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to indicate if full kubernetes address should be enabled for components.
If enabled, address will be [<pod>.]<service>.<namespace>.svc. Otherwise, it will be [<pod>.]<service>.
Enabling this flag on existing RisingWave will cause incompatibility.</p>
</td>
</tr>
<tr>
<td>
<code>enableStandaloneMode</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to control whether to deploy in standalone mode or distributed mode. If standalone mode is used,
spec.components will be ignored. Standalone mode can be turned on/off dynamically.</p>
</td>
</tr>
<tr>
<td>
<code>enableEmbeddedServingMode</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Flag to control whether to enable embedded serving mode. If enabled, the frontend nodes will be created
with embedded serving node enabled, and the compute nodes will serve streaming workload only.</p>
</td>
</tr>
<tr>
<td>
<code>enableAdvertisingWithIP</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Flag to control whether to enable advertising with IP. If enabled, the meta and compute nodes will be advertised
with their IP addresses. This is useful when one wants to avoid the DNS resolution overhead and latency.</p>
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
<tr>
<td>
<code>tls</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfiguration">
RisingWaveTLSConfiguration
</a>
</em>
</td>
<td>
<p>TLS configures the TLS/SSL certificates for SQL access.</p>
</td>
</tr>
<tr>
<td>
<code>standaloneMode</code><br/>
<em>
int32
</em>
</td>
<td>
<p>StandaloneMode determines which style of command-line args should be used for the standalone mode.
0 - auto detect by image version, 1 - the old standalone mode, 2 - standalone mode V2 (single-node).
This is only for backward compatibility and will be deprecated in the future.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStandaloneComponent">RisingWaveStandaloneComponent
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
<p>RisingWaveStandaloneComponent contains the spec for standalone component.</p>
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
<code>logLevel</code><br/>
<em>
string
</em>
</td>
<td>
<p>LogLevel controls the log level of the running nodes. It can be in any format that the underlying component supports,
e.g., in the RUST_LOG format for Rust programs. Defaults to INFO.</p>
</td>
</tr>
<tr>
<td>
<code>disallowPrintStackTraces</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DisallowPrintStackTraces determines if the stack traces are allowed to print when panic happens. This options applies
to both Rust and Java programs. Defaults to false.</p>
</td>
</tr>
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
<p>RestartAt is the time that the Pods under the group should be restarted. Setting a value on this field will
trigger a full recreation of the Pods. Defaults to nil.</p>
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
<p>Replicas is the number of the standalone Pods. Maximum is 1. Defaults to 1.</p>
</td>
</tr>
<tr>
<td>
<code>upgradeStrategy</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupUpgradeStrategy">
RisingWaveNodeGroupUpgradeStrategy
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
<code>minReadySeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Minimum number of seconds for which a newly created pod should be ready
without any of its container crashing for it to be considered available.
Defaults to 0 (pod will be considered available as soon as it is ready)</p>
</td>
</tr>
<tr>
<td>
<code>volumeClaimTemplates</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.PersistentVolumeClaim">
[]PersistentVolumeClaim
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumeClaimTemplates is a list of claims that pods are allowed to reference.
The StatefulSet controller is responsible for mapping network identities to
claims in a way that maintains the identity of a pod. Every claim in
this list must have at least one matching (by name) volumeMount in one
container in the template. A claim in this list takes precedence over
any volumes in the template, with the same name.</p>
</td>
</tr>
<tr>
<td>
<code>persistentVolumeClaimRetentionPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps">
Kubernetes apps/v1.StatefulSetPersistentVolumeClaimRetentionPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>persistentVolumeClaimRetentionPolicy describes the lifecycle of persistent
volume claims created from volumeClaimTemplates. By default, all persistent
volume claims are created as needed and retained until manually deleted. This
policy allows the lifecycle to be altered, for example by deleting persistent
volume claims when their stateful set is deleted, or when their pod is scaled
down. This requires the StatefulSetAutoDeletePVC feature gate to be enabled,
which is alpha.</p>
</td>
</tr>
<tr>
<td>
<code>progressDeadlineSeconds</code><br/>
<em>
int32
</em>
</td>
<td>
<p>The maximum time in seconds for a deployment to make progress before it
is considered to be failed. The deployment controller will continue to
process failed deployments and a condition with a ProgressDeadlineExceeded
reason will be surfaced in the deployment status. Note that progress will
not be estimated during the time a deployment is paused. Defaults to 600s.</p>
</td>
</tr>
<tr>
<td>
<code>template</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodePodTemplate">
RisingWaveNodePodTemplate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Template tells how the Pod should be started. It is an optional field. If it&rsquo;s empty, then the pod template in
the first-level fields under spec will be used.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
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
<p>DataDirectory is the directory to store the data in the object storage.
Defaults to hummock.</p>
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
<code>localDisk</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendLocalDisk">
RisingWaveStateStoreBackendLocalDisk
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Local indicates to store the data in local disk. It&rsquo;s only for test usage and strongly discouraged to
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
<code>gcs</code><br/>
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
<tr>
<td>
<code>huaweiCloudOBS</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendHuaweiCloudOBS">
RisingWaveStateStoreBackendHuaweiCloudOBS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HuaweiCloudOBS storage spec.</p>
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
<p>RisingWaveStateStoreBackendAliyunOSS is the details of AliyunOSS for compute and compactor components.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveAliyunOSSCredentials">
RisingWaveAliyunOSSCredentials
</a>
</em>
</td>
<td>
<p>RisingWaveAliyunOSSCredentials is the credentials provider from a Secret.</p>
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
<p>Bucket name of your Aliyun OSS.</p>
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
<em>(Optional)</em>
<p>Root directory of the Aliyun OSS bucket.</p>
<p>Deprecated: the field is redundant since there&rsquo;s already the data directory.
Mark it as optional now and will deprecate it in the future.</p>
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
<p>Region of Aliyun OSS service</p>
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
<em>(Optional)</em>
<p>Root directory of the Azure Blob container.</p>
<p>Deprecated: the field is redundant since there&rsquo;s already the data directory.
Mark it as optional now and will deprecate it in the future.</p>
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
<em>(Optional)</em>
<p>Root directory of the GCS bucket.</p>
<p>Deprecated: the field is redundant since there&rsquo;s already the data directory.
Mark it as optional now and will deprecate it in the future.</p>
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
<em>(Optional)</em>
<p>Root directory of the HDFS.</p>
<p>Deprecated: the field is redundant since there&rsquo;s already the data directory.
Mark it as optional now and will deprecate it in the future.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendHuaweiCloudOBS">RisingWaveStateStoreBackendHuaweiCloudOBS
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendHuaweiCloudOBS is the details of HuaweiCloudOBS for compute and compactor components.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveHuaweiCloudOBSCredentials">
RisingWaveHuaweiCloudOBSCredentials
</a>
</em>
</td>
<td>
<p>RisingWaveHuaweiCloudOBSCredentials is the credentials provider from a Secret.</p>
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
<p>Bucket name.</p>
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
<p>Region of Huawei Cloud OBS.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackendLocalDisk">RisingWaveStateStoreBackendLocalDisk
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStateStoreBackend">RisingWaveStateStoreBackend</a>)
</p>
<div>
<p>RisingWaveStateStoreBackendLocalDisk is the details of local storage for compute and compactor components.</p>
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
<code>root</code><br/>
<em>
string
</em>
</td>
<td>
<p>Root is the root directory to store the data in the object storage. It shadows the data directory.</p>
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
<p>Region of AWS S3 service. Defaults to &ldquo;us-east-1&rdquo;.</p>
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
${BUCKET}.s3.${REGION}.amazonaws.com
Both HTTP and HTTPS are allowed. The default scheme is HTTPS if not specified.</p>
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
</tr><tr><td><p>&#34;HuaweiCloudOBS&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;LocalDisk&#34;</p></td>
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
<code>internal</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveInternalStatus">
RisingWaveInternalStatus
</a>
</em>
</td>
<td>
<p>Internal status.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfiguration">RisingWaveTLSConfiguration
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
</p>
<div>
<p>RisingWaveTLSConfiguration is the TLS/SSL configuration for RisingWave&rsquo;s SQL access.</p>
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
<p>SecretName that contains the certificates. The keys must be <code>tls.key</code> and <code>tls.crt</code>.
If the secret name isn&rsquo;t provided, then TLS/SSL won&rsquo;t be enabled.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.WorkloadReplicaStatus">WorkloadReplicaStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentStatus">RisingWaveComponentStatus</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveNodeGroupStatus">RisingWaveNodeGroupStatus</a>)
</p>
<div>
<p>WorkloadReplicaStatus is a common structure for replica status of some workload.</p>
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
<p>Replicas is the declared replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>readyReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>ReadyReplicas is the ready replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>availableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>AvailableReplicas is the available replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>updatedReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>UpdatedReplicas is the update replicas of the workload.</p>
</td>
</tr>
<tr>
<td>
<code>unavailableReplicas</code><br/>
<em>
int32
</em>
</td>
<td>
<p>UnavailableReplicas is the unavailable replicas of the workload.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
