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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.MetaStorageType">MetaStorageType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorageStatus">RisingWaveMetaStorageStatus</a>)
</p>
<div>
<p>MetaStorageType is the type name of meta storage.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.ObjectStorageType">ObjectStorageType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageStatus">RisingWaveObjectStorageStatus</a>)
</p>
<div>
<p>ObjectStorageType is the type name of object storage.</p>
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
</tr><tr><td><p>&#34;Memory&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;MinIO&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;S3&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unknown&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWave">RisingWave
</h3>
<div>
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
<code>security</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSecuritySpec">
RisingWaveSecuritySpec
</a>
</em>
</td>
<td>
<p>The spec of security configurations, such as TLS config.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCommonPorts">RisingWaveComponentCommonPorts
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompactor">RisingWaveComponentCompactor</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompute">RisingWaveComponentCompute</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentFrontend">RisingWaveComponentFrontend</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMetaPorts">RisingWaveComponentMetaPorts</a>)
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentFrontend">RisingWaveComponentFrontend
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
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
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentCompactor">RisingWaveComponentCompactor</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentFrontend">RisingWaveComponentFrontend</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMeta">RisingWaveComponentMeta</a>)
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
<code>podTemplate</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Base template for Pods of RisingWave. By default, there&rsquo;s no such template
and the controller will set all unrelated fields to the default value.</p>
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
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentMeta">RisingWaveComponentMeta
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveComponentsSpec">RisingWaveComponentsSpec</a>)
</p>
<div>
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
<p>Meta component spec.</p>
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
<p>Frontend component spec.</p>
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
<p>Compute component spec.</p>
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
<p>Compactor component.</p>
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
<code>podTemplate</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Base template for Pods of RisingWave. By default, there&rsquo;s no such template
and the controller will set all unrelated fields to the default value.</p>
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
<code>podTemplate</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Base template for Pods of RisingWave. By default, there&rsquo;s no such template
and the controller will set all unrelated fields to the default value.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplatePartialObjectMeta">
RisingWavePodTemplatePartialObjectMeta
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorage">RisingWaveMetaStorage
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">RisingWaveStoragesSpec</a>)
</p>
<div>
<p>RisingWaveMetaStorage is the storage for the meta component.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorageEtcd">
RisingWaveMetaStorageEtcd
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorageEtcd">RisingWaveMetaStorageEtcd
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorage">RisingWaveMetaStorage</a>)
</p>
<div>
<p>RisingWaveMetaStorageEtcd is the etcd storage for the meta component.</p>
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
But it is an optional field. Empty value indicates etcd is available without authentication.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorageStatus">RisingWaveMetaStorageStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesStatus">RisingWaveStoragesStatus</a>)
</p>
<div>
<p>RisingWaveMetaStorageStatus is the status of meta storage.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.MetaStorageType">
MetaStorageType
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorage">RisingWaveObjectStorage
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesSpec">RisingWaveStoragesSpec</a>)
</p>
<div>
<p>RisingWaveObjectStorage is the object storage for compute and compactor components.</p>
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
<p>Memory indicates to store the data in memory. It&rsquo;s only for test usage and strongly discouraged to
be used in production.</p>
</td>
</tr>
<tr>
<td>
<code>minio</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageMinIO">
RisingWaveObjectStorageMinIO
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageS3">
RisingWaveObjectStorageS3
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
<code>aliyunOSS</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageAliyunOSS">
RisingWaveObjectStorageAliyunOSS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AliyunOSS storage spec.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageAliyunOSS">RisingWaveObjectStorageAliyunOSS
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorage">RisingWaveObjectStorage</a>)
</p>
<div>
<p>RisingWaveObjectStorageAliyunOSS is the details of Aliyun OSS storage (S3 compatible) for compute and compactor components.</p>
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
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the Aliyun OSS service. It must contain the following keys:
* AccessKeyID
* SecretAccessKey
* Region (optional if region is specified in the field.)</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageMinIO">RisingWaveObjectStorageMinIO
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorage">RisingWaveObjectStorage</a>)
</p>
<div>
<p>RisingWaveObjectStorageMinIO is the details of MinIO storage for compute and compactor components.</p>
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
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the MinIO service. It must contain the following keys:
* username
* password</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageS3">RisingWaveObjectStorageS3
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorage">RisingWaveObjectStorage</a>)
</p>
<div>
<p>RisingWaveObjectStorageS3 is the details of AWS S3 storage for compute and compactor components.</p>
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
<code>secret</code><br/>
<em>
string
</em>
</td>
<td>
<p>Secret contains the credentials to access the AWS S3 service. It must contain the following keys:
* AccessKeyID
* SecretAccessKey
* Region (optional if region is specified in the field.)</p>
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
<tr>
<td>
<code>virtualHostedStyle</code><br/>
<em>
bool
</em>
</td>
<td>
<p>VirtualHostedStyle indicates to use a virtual hosted endpoint when endpoint is specified. The operator automatically
adds the bucket prefix for you if this is enabled. Be careful about doubly using the style by specifying an endpoint
of virtual hosted style as well as enabling this.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageStatus">RisingWaveObjectStorageStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesStatus">RisingWaveStoragesStatus</a>)
</p>
<div>
<p>RisingWaveObjectStorageStatus is the status of object storage.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.ObjectStorageType">
ObjectStorageType
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplate">RisingWavePodTemplate
</h3>
<div>
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
<code>template</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplateSpec">
RisingWavePodTemplateSpec
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplatePartialObjectMeta">RisingWavePodTemplatePartialObjectMeta
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveGlobalSpec">RisingWaveGlobalSpec</a>, <a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplateSpec">RisingWavePodTemplateSpec</a>)
</p>
<div>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplateSpec">RisingWavePodTemplateSpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplate">RisingWavePodTemplate</a>)
</p>
<div>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWavePodTemplatePartialObjectMeta">
RisingWavePodTemplatePartialObjectMeta
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podspec-v1-core">
Kubernetes core/v1.PodSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
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
<code>initContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>List of initialization containers belonging to the pod.
Init containers are executed in order prior to containers being started. If any
init container fails, the pod is considered to have failed and is handled according
to its restartPolicy. The name for an init container or normal container must be
unique among all containers.
Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes.
The resourceRequirements of an init container are taken into account during scheduling
by finding the highest request/limit for each resource type, and then using the max of
of that value or the sum of the normal containers. Limits are applied to init containers
in a similar fashion.
Init containers cannot currently be added or removed.
Cannot be updated.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/init-containers/">https://kubernetes.io/docs/concepts/workloads/pods/init-containers/</a></p>
</td>
</tr>
<tr>
<td>
<code>containers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#container-v1-core">
[]Kubernetes core/v1.Container
</a>
</em>
</td>
<td>
<p>List of containers belonging to the pod.
Containers cannot currently be added or removed.
There must be at least one container in a Pod.
Cannot be updated.</p>
</td>
</tr>
<tr>
<td>
<code>ephemeralContainers</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#ephemeralcontainer-v1-core">
[]Kubernetes core/v1.EphemeralContainer
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>List of ephemeral containers run in this pod. Ephemeral containers may be run in an existing
pod to perform user-initiated actions such as debugging. This list cannot be specified when
creating a pod, and it cannot be modified by updating the pod spec. In order to add an
ephemeral container to an existing pod, use the pod&rsquo;s ephemeralcontainers subresource.
This field is beta-level and available on clusters that haven&rsquo;t disabled the EphemeralContainers feature gate.</p>
</td>
</tr>
<tr>
<td>
<code>restartPolicy</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#restartpolicy-v1-core">
Kubernetes core/v1.RestartPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Restart policy for all containers within the pod.
One of Always, OnFailure, Never.
Default to Always.
More info: <a href="https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy">https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy</a></p>
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
<code>serviceAccount</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>DeprecatedServiceAccount is a depreciated alias for ServiceAccountName.
Deprecated: Use serviceAccountName instead.</p>
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
<code>nodeName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
the scheduler simply schedules this pod onto that node, assuming that it fits resource
requirements.</p>
</td>
</tr>
<tr>
<td>
<code>hostNetwork</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Host networking requested for this pod. Use the host&rsquo;s network namespace.
If this option is set, the ports that will be used must be specified.
Default to false.</p>
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
<code>hostname</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Specifies the hostname of the Pod
If not specified, the pod&rsquo;s hostname will be set to a system-defined value.</p>
</td>
</tr>
<tr>
<td>
<code>subdomain</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, the fully qualified Pod hostname will be &ldquo;<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>&rdquo;.
If not specified, the pod will not have a domainname at all.</p>
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
<code>readinessGates</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#podreadinessgate-v1-core">
[]Kubernetes core/v1.PodReadinessGate
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If specified, all readiness gates will be evaluated for pod readiness.
A pod is ready when all its containers are ready AND
all conditions specified in the readiness gates have status equal to &ldquo;True&rdquo;
More info: <a href="https://git.k8s.io/enhancements/keps/sig-network/580-pod-readiness-gates">https://git.k8s.io/enhancements/keps/sig-network/580-pod-readiness-gates</a></p>
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
<code>enableServiceLinks</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>EnableServiceLinks indicates whether information about services should be injected into pod&rsquo;s
environment variables, matching the syntax of Docker links.
Optional: Defaults to true.</p>
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
<code>overhead</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overhead represents the resource overhead associated with running a pod for a given RuntimeClass.
This field will be autopopulated at admission time by the RuntimeClass admission controller. If
the RuntimeClass admission controller is enabled, overhead must not be set in Pod create requests.
The RuntimeClass admission controller will reject Pod create requests which have the overhead already
set. If RuntimeClass is configured and selected in the PodSpec, Overhead will be set to the value
defined in the corresponding RuntimeClass, otherwise it will remain unset and treated as zero.
More info: <a href="https://git.k8s.io/enhancements/keps/sig-node/688-pod-overhead/README.md">https://git.k8s.io/enhancements/keps/sig-node/688-pod-overhead/README.md</a></p>
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
- spec.containers[*].securityContext.runAsGroup
This is a beta field and requires the IdentifyPodOS feature</p>
</td>
</tr>
</table>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveScaleView">RisingWaveScaleView
</h3>
<div>
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
<p>RisingWaveScaleViewLock is a lock for RisingWaveScaleViews.</p>
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveSecuritySpec">RisingWaveSecuritySpec
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSpec">RisingWaveSpec</a>)
</p>
<div>
<p>RisingWaveSecuritySpec is the security spec.</p>
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
<code>tls</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfig">
RisingWaveTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS config of RisingWave.</p>
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
<code>security</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSecuritySpec">
RisingWaveSecuritySpec
</a>
</em>
</td>
<td>
<p>The spec of security configurations, such as TLS config.</p>
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
<code>storages</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesStatus">
RisingWaveStoragesStatus
</a>
</em>
</td>
<td>
<p>Status of the external storages.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorage">
RisingWaveMetaStorage
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>object</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorage">
RisingWaveObjectStorage
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#persistentvolumeclaim-v1-core">
[]Kubernetes core/v1.PersistentVolumeClaim
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
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveStoragesStatus">RisingWaveStoragesStatus
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveStatus">RisingWaveStatus</a>)
</p>
<div>
<p>RisingWaveStoragesStatus is the status of external storages.</p>
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
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveMetaStorageStatus">
RisingWaveMetaStorageStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>object</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveObjectStorageStatus">
RisingWaveObjectStorageStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfig">RisingWaveTLSConfig
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveSecuritySpec">RisingWaveSecuritySpec</a>)
</p>
<div>
<p>RisingWaveTLSConfig is the TLS config of RisingWave.</p>
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
<code>enabled</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Enabled indicates if TLS is enabled on RisingWave.</p>
</td>
</tr>
<tr>
<td>
<code>secret</code><br/>
<em>
<a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfigSecret">
RisingWaveTLSConfigSecret
</a>
</em>
</td>
<td>
<p>Secret contains the TLS config. If TLS is enabled, the secret
must be provided.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfigSecret">RisingWaveTLSConfigSecret
</h3>
<p>
(<em>Appears on:</em><a href="#risingwave.risingwavelabs.com/v1alpha1.RisingWaveTLSConfig">RisingWaveTLSConfig</a>)
</p>
<div>
<p>RisingWaveTLSConfigSecret is the secret reference that contains the key and cert for TLS.</p>
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
<p>Name of the secret.</p>
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
<p>The key of the TLS key. A default value of &ldquo;tls.key&rdquo; will be used if not specified.</p>
</td>
</tr>
<tr>
<td>
<code>cert</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The key of the TLS cert. A default value of &ldquo;tls.crt&rdquo; will be used if not specified.</p>
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
