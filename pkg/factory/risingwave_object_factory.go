/*
 * Copyright 2023 RisingWave Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package factory

import (
	"fmt"
	"math"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	"golang.org/x/mod/semver"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/factory/envs"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

const (
	risingWaveConfigVolume = "risingwave-config"
	risingWaveConfigMapKey = "risingwave.toml"

	risingwaveExecutablePath  = "/risingwave/bin/risingwave"
	risingwaveConfigMountPath = "/risingwave/config"
	risingwaveConfigFileName  = "risingwave.toml"
)

var (
	aliyunOSSEndpoint         = fmt.Sprintf("https://oss-$(%s).aliyuncs.com", envs.AliyunOSSRegion)
	internalAliyunOSSEndpoint = fmt.Sprintf("https://oss-$(%s)-internal.aliyuncs.com", envs.AliyunOSSRegion)
	huaweiCloudOBSEndpoint    = fmt.Sprintf("https://obs.$(%s).myhuaweicloud.com", envs.HuaweiCloudOBSRegion)
)

// The minimum version that supports specifying username and password separately for SQL meta stores.
// Refer to https://github.com/risingwavelabs/risingwave/pull/17530 for more details.
const minVersionForSQLCredentialsViaEnv = "v1.10.1"

func isSQLCredentialsViaEnvSupported(version string) bool {
	// Either the version is greater than or equal to the minimum version,
	// or it is a nightly version after the date when the feature was introduced.
	return semver.Compare(version, minVersionForSQLCredentialsViaEnv) >= 0 ||
		utils.IsRisingWaveNightlyVersionAfter(version, utils.Date(2023, 7, 14))
}

// RisingWaveObjectFactory is the object factory to help create owned objects like Deployments, StatefulSets, Services, etc.
type RisingWaveObjectFactory struct {
	scheme     *runtime.Scheme
	risingwave *risingwavev1alpha1.RisingWave

	inheritedLabels map[string]string
	operatorVersion string
}

func (f *RisingWaveObjectFactory) namespace() string {
	return f.risingwave.Namespace
}

func (f *RisingWaveObjectFactory) isStateStoreMemory() bool {
	return ptr.Deref(f.risingwave.Spec.StateStore.Memory, false)
}

func (f *RisingWaveObjectFactory) isStateStoreS3() bool {
	return f.risingwave.Spec.StateStore.S3 != nil && len(f.risingwave.Spec.StateStore.S3.Endpoint) == 0
}

func (f *RisingWaveObjectFactory) isStateStoreS3Compatible() bool {
	return f.risingwave.Spec.StateStore.S3 != nil && len(f.risingwave.Spec.StateStore.S3.Endpoint) > 0
}

func (f *RisingWaveObjectFactory) isStateStoreGCS() bool {
	return f.risingwave.Spec.StateStore.GCS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreAliyunOSS() bool {
	return f.risingwave.Spec.StateStore.AliyunOSS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreAzureBlob() bool {
	return f.risingwave.Spec.StateStore.AzureBlob != nil
}

func (f *RisingWaveObjectFactory) isStateStoreHDFS() bool {
	return f.risingwave.Spec.StateStore.HDFS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreWebHDFS() bool {
	return f.risingwave.Spec.StateStore.WebHDFS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreMinIO() bool {
	return f.risingwave.Spec.StateStore.MinIO != nil
}

func (f *RisingWaveObjectFactory) isStateStoreLocalDisk() bool {
	return f.risingwave.Spec.StateStore.LocalDisk != nil
}

func (f *RisingWaveObjectFactory) isStateStoreHuaweiCloudOBS() bool {
	return f.risingwave.Spec.StateStore.HuaweiCloudOBS != nil
}

func (f *RisingWaveObjectFactory) isMetaStoreMemory() bool {
	return ptr.Deref(f.risingwave.Spec.MetaStore.Memory, false)
}

func (f *RisingWaveObjectFactory) isMetaStoreEtcd() bool {
	return f.risingwave.Spec.MetaStore.Etcd != nil
}

func (f *RisingWaveObjectFactory) isMetaStoreSQLite() bool {
	return f.risingwave.Spec.MetaStore.SQLite != nil
}

func (f *RisingWaveObjectFactory) isMetaStoreMySQL() bool {
	return f.risingwave.Spec.MetaStore.MySQL != nil
}

func (f *RisingWaveObjectFactory) isMetaStorePostgreSQL() bool {
	return f.risingwave.Spec.MetaStore.PostgreSQL != nil
}

func (f *RisingWaveObjectFactory) isFullKubernetesAddr() bool {
	return ptr.Deref(f.risingwave.Spec.EnableFullKubernetesAddr, false)
}

func (f *RisingWaveObjectFactory) hummockConnectionStr() string {
	stateStore := f.risingwave.Spec.StateStore
	switch {
	case f.isStateStoreMemory():
		return "hummock+memory"
	case f.isStateStoreS3():
		bucket := stateStore.S3.Bucket
		return fmt.Sprintf("hummock+s3://%s", bucket)
	case f.isStateStoreS3Compatible():
		bucket := stateStore.S3.Bucket
		return fmt.Sprintf("hummock+s3://%s", bucket)
	case stateStore.MinIO != nil:
		minio := stateStore.MinIO
		return fmt.Sprintf("hummock+minio://$(%s):$(%s)@%s/%s", envs.MinIOUsername, envs.MinIOPassword, minio.Endpoint, minio.Bucket)
	case f.isStateStoreGCS():
		return fmt.Sprintf("hummock+gcs://%s", stateStore.GCS.Bucket)
	case f.isStateStoreAliyunOSS():
		aliyunOSS := stateStore.AliyunOSS
		return fmt.Sprintf("hummock+oss://%s", aliyunOSS.Bucket)
	case f.isStateStoreAzureBlob():
		azureBlob := stateStore.AzureBlob
		return fmt.Sprintf("hummock+azblob://%s", azureBlob.Container)
	case f.isStateStoreHDFS():
		hdfs := stateStore.HDFS
		return fmt.Sprintf("hummock+hdfs://%s", hdfs.NameNode)
	case f.isStateStoreWebHDFS():
		webhdfs := stateStore.WebHDFS
		return fmt.Sprintf("hummock+webhdfs://%s", webhdfs.NameNode)
	case f.isStateStoreLocalDisk():
		localDisk := stateStore.LocalDisk
		return fmt.Sprintf("hummock+fs://%s", localDisk.Root)
	case f.isStateStoreHuaweiCloudOBS():
		return fmt.Sprintf("hummock+obs://%s", stateStore.HuaweiCloudOBS.Bucket)
	default:
		panic("unrecognized storage type")
	}
}

func groupSuffix(group string) string {
	if group == "" {
		return ""
	}
	return "-" + group
}

func (f *RisingWaveObjectFactory) componentName(component, group string) string {
	switch component {
	case consts.ComponentMeta:
		return f.risingwave.Name + "-meta" + groupSuffix(group)
	case consts.ComponentCompute:
		return f.risingwave.Name + "-compute" + groupSuffix(group)
	case consts.ComponentFrontend:
		return f.risingwave.Name + "-frontend" + groupSuffix(group)
	case consts.ComponentCompactor:
		return f.risingwave.Name + "-compactor" + groupSuffix(group)
	case consts.ComponentStandalone:
		return f.risingwave.Name + "-standalone"
	case consts.ComponentConfig:
		return f.risingwave.Name + "-default-config"
	default:
		panic("never reach here")
	}
}

func (f *RisingWaveObjectFactory) componentAddr(component, group string) string {
	componentName := f.componentName(component, group)
	if f.isFullKubernetesAddr() {
		return fmt.Sprintf("%s.$(POD_NAMESPACE).svc", componentName)
	}
	return componentName
}

func (f *RisingWaveObjectFactory) getObjectMetaForGeneralResources(name string, sync bool) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:            f.risingwave.Name,
			consts.LabelRisingWaveGeneration:      lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
			consts.LabelRisingWaveOperatorVersion: f.operatorVersion,
		},
	}

	objectMeta.Labels = mergeMap(objectMeta.Labels, f.getInheritedLabels())

	return objectMeta
}

func (f *RisingWaveObjectFactory) getObjectMetaForComponentLevelResources(component string, sync bool) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name:      f.componentName(component, ""),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:            f.risingwave.Name,
			consts.LabelRisingWaveComponent:       component,
			consts.LabelRisingWaveGeneration:      lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
			consts.LabelRisingWaveOperatorVersion: f.operatorVersion,
		},
	}

	objectMeta.Labels = mergeMap(objectMeta.Labels, f.getInheritedLabels())

	return objectMeta
}

func (f *RisingWaveObjectFactory) getObjectMetaForComponentGroupLevelResources(component, group string, sync bool) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name:      f.componentName(component, group),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:            f.risingwave.Name,
			consts.LabelRisingWaveComponent:       component,
			consts.LabelRisingWaveGeneration:      lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
			consts.LabelRisingWaveGroup:           group,
			consts.LabelRisingWaveOperatorVersion: f.operatorVersion,
		},
	}

	objectMeta.Labels = mergeMap(objectMeta.Labels, f.getInheritedLabels())

	return objectMeta
}

func (f *RisingWaveObjectFactory) podLabelsOrSelectorsForComponent(component string) map[string]string {
	return map[string]string{
		consts.LabelRisingWaveName:      f.risingwave.Name,
		consts.LabelRisingWaveComponent: component,
	}
}

func (f *RisingWaveObjectFactory) podLabelsOrSelectorsForComponentGroup(component, group string) map[string]string {
	return map[string]string{
		consts.LabelRisingWaveName:      f.risingwave.Name,
		consts.LabelRisingWaveComponent: component,
		consts.LabelRisingWaveGroup:     group,
	}
}

func (f *RisingWaveObjectFactory) newService(component string, serviceType corev1.ServiceType, ports []corev1.ServicePort) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: f.getObjectMetaForComponentLevelResources(component, true),
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Selector: f.podLabelsOrSelectorsForComponent(component),
			Ports:    ports,
		},
	}
}

func (f *RisingWaveObjectFactory) envsForEtcd() []corev1.EnvVar {
	credentials := f.risingwave.Spec.MetaStore.Etcd.RisingWaveEtcdCredentials

	// Empty secret indicates no authentication.
	if credentials == nil || credentials.SecretName == "" {
		return []corev1.EnvVar{}
	}

	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}

	return []corev1.EnvVar{
		// Keep the legacy environment variables for compatibility. Will remove them later.
		{
			Name: envs.EtcdUsernameLegacy,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.UsernameKeyRef,
				},
			},
		},
		{
			Name: envs.EtcdPasswordLegacy,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.PasswordKeyRef,
				},
			},
		},
		// Environment variables for etcd auth.
		{
			Name: envs.RWEtcdUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.UsernameKeyRef,
				},
			},
		},
		{
			Name: envs.RWEtcdPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.PasswordKeyRef,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForSQLite() []corev1.EnvVar {
	return nil
}

func (f *RisingWaveObjectFactory) envsForMySQL() []corev1.EnvVar {
	creds := &f.risingwave.Spec.MetaStore.MySQL.RisingWaveDBCredentials

	return []corev1.EnvVar{
		{
			Name: envs.RWMySQLUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.UsernameKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
		{
			Name: envs.RWSQLUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.UsernameKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
		{
			Name: envs.RWMySQLPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.PasswordKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
		{
			Name: envs.RWSQLPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.PasswordKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForPostgreSQL() []corev1.EnvVar {
	creds := &f.risingwave.Spec.MetaStore.PostgreSQL.RisingWaveDBCredentials

	return []corev1.EnvVar{
		{
			Name: envs.RWPostgresUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.UsernameKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
		{
			Name: envs.RWPostgresPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.PasswordKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
		{
			Name: envs.RWSQLUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.UsernameKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
		{
			Name: envs.RWSQLPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: creds.PasswordKeyRef,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: creds.SecretName,
					},
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForMetaStore() []corev1.EnvVar {
	switch {
	case f.isMetaStoreMemory():
		return nil
	case f.isMetaStoreEtcd():
		return f.envsForEtcd()
	case f.isMetaStoreSQLite():
		return f.envsForSQLite()
	case f.isMetaStoreMySQL():
		return f.envsForMySQL()
	case f.isMetaStorePostgreSQL():
		return f.envsForPostgreSQL()
	default:
		panic("unsupported meta storage type")
	}
}

func formatConnectionOptions(opts map[string]string) string {
	if len(opts) == 0 {
		return ""
	}
	val := url.Values{}
	for k, v := range opts {
		val.Add(k, v)
	}
	return "?" + val.Encode()
}

func (f *RisingWaveObjectFactory) getDataDirectory() string {
	stateStore := f.risingwave.Spec.StateStore
	// In terms of compatibility, the root path is either the path in the internal status,
	// or the deprecated data path.
	rootPath := f.risingwave.Status.Internal.StateStoreRootPath
	if rootPath == "" {
		rootPath = object.NewRisingWaveReader(f.risingwave).StateStoreRootPath()
	}

	return path.Join(rootPath, stateStore.DataDirectory)
}

func (f *RisingWaveObjectFactory) coreEnvsForMeta(image string) []corev1.EnvVar {
	metaStore := &f.risingwave.Spec.MetaStore
	version := utils.GetVersionFromImage(image)

	envVars := []corev1.EnvVar{
		{
			Name:  envs.RWStateStore,
			Value: f.hummockConnectionStr(),
		},
		{
			Name:  envs.RWDataDirectory,
			Value: f.getDataDirectory(),
		},
		{
			Name:  envs.RWDashboardHost,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.MetaDashboardPort),
		},
	}

	switch {
	case f.isMetaStoreMemory():
		envVars = append(envVars, corev1.EnvVar{
			Name:  envs.RWBackend,
			Value: "mem",
		})
	case f.isMetaStoreEtcd():
		envVars = append(envVars, []corev1.EnvVar{
			{
				Name:  envs.RWBackend,
				Value: "etcd",
			},
			{
				Name:  envs.RWEtcdEndpoints,
				Value: metaStore.Etcd.Endpoint,
			},
		}...)
		credentials := f.risingwave.Spec.MetaStore.Etcd.RisingWaveEtcdCredentials
		if credentials != nil && credentials.SecretName != "" {
			envVars = append(envVars, corev1.EnvVar{
				Name:  envs.RWEtcdAuth,
				Value: "true",
			})
		}
	case f.isMetaStoreSQLite():
		if isSQLCredentialsViaEnvSupported(version) {
			envVars = append(envVars, []corev1.EnvVar{
				{
					Name:  envs.RWBackend,
					Value: "sqlite",
				},
				{
					Name:  envs.RWSQLEndpoint,
					Value: fmt.Sprintf("%s?mode=rwc", metaStore.SQLite.Path),
				},
			}...)
		} else {
			envVars = append(envVars, []corev1.EnvVar{
				{
					Name:  envs.RWBackend,
					Value: "sql",
				},
				{
					Name:  envs.RWSQLEndpoint,
					Value: fmt.Sprintf("sqlite://%s?mode=rwc", metaStore.SQLite.Path),
				},
			}...)
		}
	case f.isMetaStoreMySQL():
		if isSQLCredentialsViaEnvSupported(version) {
			envVars = append(envVars, []corev1.EnvVar{
				{
					Name:  envs.RWBackend,
					Value: "mysql",
				},
				{
					Name: envs.RWSQLEndpoint,
					Value: fmt.Sprintf("%s:%d%s", metaStore.MySQL.Host, metaStore.MySQL.Port,
						formatConnectionOptions(metaStore.MySQL.Options)),
				},
				{
					Name:  envs.RWSQLDatabase,
					Value: metaStore.MySQL.Database,
				},
			}...)
		} else {
			envVars = append(envVars, []corev1.EnvVar{
				{
					Name:  envs.RWBackend,
					Value: "sql",
				},
				{
					Name: envs.RWSQLEndpoint,
					Value: fmt.Sprintf("mysql://$(%s):$(%s)@%s:%d/%s%s", envs.RWMySQLUsername, envs.RWMySQLPassword,
						metaStore.MySQL.Host, metaStore.MySQL.Port, metaStore.MySQL.Database,
						formatConnectionOptions(metaStore.MySQL.Options)),
				},
			}...)
		}
	case f.isMetaStorePostgreSQL():
		if isSQLCredentialsViaEnvSupported(version) {
			envVars = append(envVars, []corev1.EnvVar{
				{
					Name:  envs.RWBackend,
					Value: "postgres",
				},
				{
					Name: envs.RWSQLEndpoint,
					Value: fmt.Sprintf("%s:%d%s", metaStore.PostgreSQL.Host, metaStore.PostgreSQL.Port,
						formatConnectionOptions(metaStore.PostgreSQL.Options)),
				},
				{
					Name:  envs.RWSQLDatabase,
					Value: metaStore.PostgreSQL.Database,
				},
			}...)
		} else {
			envVars = append(envVars, []corev1.EnvVar{
				{
					Name:  envs.RWBackend,
					Value: "sql",
				},
				{
					Name: envs.RWSQLEndpoint,
					Value: fmt.Sprintf("postgres://$(%s):$(%s)@%s:%d/%s%s", envs.RWPostgresUsername,
						envs.RWPostgresPassword, metaStore.PostgreSQL.Host, metaStore.PostgreSQL.Port,
						metaStore.PostgreSQL.Database, formatConnectionOptions(metaStore.PostgreSQL.Options)),
				},
			}...)
		}
	default:
		panic("unsupported meta storage type")
	}

	return envVars
}

func (f *RisingWaveObjectFactory) envsForMetaArgs(image string) []corev1.EnvVar {
	var advertiseAddr string
	if ptr.Deref(f.risingwave.Spec.EnableAdvertisingWithIP, false) {
		advertiseAddr = fmt.Sprintf("$(POD_IP):%d", consts.MetaServicePort)
	} else {
		advertiseAddr = fmt.Sprintf("$(POD_NAME).%s:%d", f.componentAddr(consts.ComponentMeta, ""), consts.MetaServicePort)
	}

	envVars := []corev1.EnvVar{
		// Env var for local risectl access.
		{
			Name:  envs.RWMetaAddr,
			Value: fmt.Sprintf("http://127.0.0.1:%d", consts.MetaServicePort),
		},
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.MetaServicePort),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: advertiseAddr,
		},
		{
			Name:  envs.RWPrometheusHost,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.MetaMetricsPort),
		},
	}

	envVars = append(envVars, f.coreEnvsForMeta(image)...)

	return envVars
}

func (f *RisingWaveObjectFactory) envsForTLS() []corev1.EnvVar {
	tls := f.risingwave.Spec.TLS
	if tls != nil && tls.SecretName != "" {
		return []corev1.EnvVar{
			{
				Name: envs.RWSslKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: tls.SecretName,
						},
						Key: corev1.TLSPrivateKeyKey,
					},
				},
			},
			{
				Name: envs.RWSslCert,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: tls.SecretName,
						},
						Key: corev1.TLSCertKey,
					},
				},
			},
		}
	}
	return nil
}

func (f *RisingWaveObjectFactory) envsForFrontendArgs() []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.FrontendServicePort),
		},
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: fmt.Sprintf("$(POD_IP):%d", consts.FrontendServicePort),
		},
		{
			Name:  envs.RWMetaAddr,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentAddr(consts.ComponentMeta, ""), consts.MetaServicePort),
		},
		{
			Name:  envs.RWMetaAddrLegacy,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentAddr(consts.ComponentMeta, ""), consts.MetaServicePort),
		},
		{
			Name:  envs.RWPrometheusListenerAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.FrontendMetricsPort),
		},
	}

	return append(envVars, f.envsForTLS()...)
}

func (f *RisingWaveObjectFactory) envsForResourceLimits(cpuLimit int64, memLimit int64) []corev1.EnvVar {
	var envVars []corev1.EnvVar

	if cpuLimit != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  envs.RWParallelism,
			Value: strconv.FormatInt(cpuLimit, 10),
		})
	}

	if memLimit != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  envs.RWTotalMemoryBytes,
			Value: strconv.FormatInt(memLimit, 10),
		})
	}

	return envVars
}

func (f *RisingWaveObjectFactory) envsForComputeArgs(cpuLimit int64, memLimit int64) []corev1.EnvVar {
	var advertiseAddr string
	if ptr.Deref(f.risingwave.Spec.EnableAdvertisingWithIP, false) {
		advertiseAddr = fmt.Sprintf("$(POD_IP):%d", consts.ComputeServicePort)
	} else {
		advertiseAddr = fmt.Sprintf("$(POD_NAME).%s:%d", f.componentAddr(consts.ComponentCompute, ""), consts.ComputeServicePort)
	}

	envVars := []corev1.EnvVar{
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.ComputeServicePort),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: advertiseAddr,
		},
		{
			Name:  envs.RWMetaAddr,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentAddr(consts.ComponentMeta, ""), consts.MetaServicePort),
		},
		{
			Name:  envs.RWMetaAddrLegacy,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentAddr(consts.ComponentMeta, ""), consts.MetaServicePort),
		},
		{
			Name:  envs.RWPrometheusListenerAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.ComputeMetricsPort),
		},
	}

	envVars = append(envVars, f.envsForResourceLimits(cpuLimit, memLimit)...)

	return envVars
}

func (f *RisingWaveObjectFactory) envsForCompactorArgs() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.CompactorServicePort),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: fmt.Sprintf("$(POD_IP):%d", consts.CompactorServicePort),
		},
		{
			Name:  envs.RWPrometheusListenerAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", consts.CompactorMetricsPort),
		},
		{
			Name:  envs.RWMetaAddr,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentAddr(consts.ComponentMeta, ""), consts.MetaServicePort),
		},
		{
			Name:  envs.RWMetaAddrLegacy,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentAddr(consts.ComponentMeta, ""), consts.MetaServicePort),
		},
	}
}

func (f *RisingWaveObjectFactory) envsForMinIO() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore
	credentials := &stateStore.MinIO.RisingWaveMinIOCredentials

	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}

	return []corev1.EnvVar{
		{
			Name:  envs.MinIOEndpoint,
			Value: stateStore.MinIO.Endpoint,
		},
		{
			Name:  envs.MinIOBucket,
			Value: stateStore.MinIO.Bucket,
		},
		{
			Name: envs.MinIOUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.UsernameKeyRef,
				},
			},
		},
		{
			Name: envs.MinIOPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.PasswordKeyRef,
				},
			},
		},
	}
}

func envsForAWSS3(region, bucket string, credentials risingwavev1alpha1.RisingWaveS3Credentials) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  envs.AWSRegion,
			Value: region,
		},
		{
			Name:  envs.AWSS3Bucket,
			Value: bucket,
		},
	}

	if !ptr.Deref(credentials.UseServiceAccount, false) {
		secretRef := corev1.LocalObjectReference{
			Name: credentials.SecretName,
		}
		credentialsEnvVars := []corev1.EnvVar{
			{
				Name: envs.AWSAccessKeyID,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: secretRef,
						Key:                  credentials.AccessKeyRef,
					},
				},
			},
			{
				Name: envs.AWSSecretAccessKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: secretRef,
						Key:                  credentials.SecretAccessKeyRef,
					},
				},
			},
		}
		envVars = append(envVars, credentialsEnvVars...)
	}

	return envVars
}

func (f *RisingWaveObjectFactory) envsForS3() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore
	s3Spec := stateStore.S3

	if len(s3Spec.Endpoint) > 0 {
		// S3 compatible mode.
		endpoint := strings.TrimSpace(s3Spec.Endpoint)

		// Interpret the variables.
		endpoint = strings.ReplaceAll(endpoint, "${REGION}", fmt.Sprintf("$(%s)", envs.S3CompatibleRegion))
		endpoint = strings.ReplaceAll(endpoint, "${BUCKET}", fmt.Sprintf("$(%s)", envs.S3CompatibleBucket))

		if !strings.HasPrefix(endpoint, "https://") {
			endpoint = "https://" + endpoint
		}
		return envsForS3Compatible(s3Spec.Region, endpoint, s3Spec.Bucket, s3Spec.RisingWaveS3Credentials)
	}

	// AWS S3 mode.
	return envsForAWSS3(s3Spec.Region, s3Spec.Bucket, s3Spec.RisingWaveS3Credentials)
}

func envsForS3Compatible(region, endpoint, bucket string, credentials risingwavev1alpha1.RisingWaveS3Credentials) []corev1.EnvVar {
	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}

	var regionEnvVar corev1.EnvVar
	if len(region) > 0 {
		regionEnvVar = corev1.EnvVar{
			Name:  envs.S3CompatibleRegion,
			Value: region,
		}
	} else {
		regionEnvVar = corev1.EnvVar{
			Name: envs.S3CompatibleRegion,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyAWSS3Region,
				},
			},
		}
	}

	return []corev1.EnvVar{
		{
			// Disable auto region loading. Refer to the original source for more information.
			// https://github.com/awslabs/aws-sdk-rust/blob/main/sdk/aws-config/src/imds/region.rs
			// cspell:disable-next-line
			Name:  envs.AWSEC2MetadataDisabled,
			Value: "true",
		},
		{
			Name:  envs.S3CompatibleBucket,
			Value: bucket,
		},
		regionEnvVar,
		{
			Name: envs.S3CompatibleAccessKeyID,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccessKeyRef,
				},
			},
		},
		{
			Name: envs.S3CompatibleSecretAccessKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.SecretAccessKeyRef,
				},
			},
		},
		{
			Name:  envs.S3CompatibleEndpoint,
			Value: endpoint,
		},
	}
}

func (f *RisingWaveObjectFactory) envsForGCS() []corev1.EnvVar {
	gcs := f.risingwave.Spec.StateStore.GCS
	useWorkloadIdentity := ptr.Deref(gcs.UseWorkloadIdentity, false)
	if useWorkloadIdentity {
		return []corev1.EnvVar{}
	}

	credentials := gcs.RisingWaveGCSCredentials
	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}
	return []corev1.EnvVar{
		{
			Name: envs.GoogleApplicationCredentials,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.ServiceAccountCredentialsKeyRef,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForAliyunOSS() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore
	credentials := stateStore.AliyunOSS.RisingWaveAliyunOSSCredentials
	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}
	var endpoint string
	if stateStore.AliyunOSS.InternalEndpoint {
		endpoint = internalAliyunOSSEndpoint
	} else {
		endpoint = aliyunOSSEndpoint
	}

	return []corev1.EnvVar{
		{
			Name:  envs.AliyunOSSRegion,
			Value: stateStore.AliyunOSS.Region,
		},
		{
			Name: envs.AliyunOSSAccessKeyID,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccessKeyIDRef,
				},
			},
		},
		{
			Name: envs.AliyunOSSSecretAccessKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccessKeySecretRef,
				},
			},
		},
		{
			Name:  envs.AliyunOSSEndpoint,
			Value: endpoint,
		},
	}
}

func (f *RisingWaveObjectFactory) envsForAzureBlob() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore
	credentials := stateStore.AzureBlob.RisingWaveAzureBlobCredentials
	envVars := []corev1.EnvVar{
		{
			Name:  envs.AzureBlobEndpoint,
			Value: stateStore.AzureBlob.Endpoint,
		},
	}

	if !ptr.Deref(credentials.UseServiceAccount, false) {
		secretRef := corev1.LocalObjectReference{
			Name: credentials.SecretName,
		}
		envVars = append(envVars, []corev1.EnvVar{
			{
				Name: envs.AzureBlobAccountName,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: secretRef,
						Key:                  credentials.AccountNameRef,
					},
				},
			},
			{
				Name: envs.AzureBlobAccountKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: secretRef,
						Key:                  credentials.AccountKeyRef,
					},
				},
			},
		}...)
	}

	return envVars
}

func (f *RisingWaveObjectFactory) envsForHDFS() []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func (f *RisingWaveObjectFactory) envsForWebHDFS() []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func (f *RisingWaveObjectFactory) envsForLocalDisk() []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func (f *RisingWaveObjectFactory) envsForStateStore() []corev1.EnvVar {
	switch {
	case f.isStateStoreMinIO():
		return f.envsForMinIO()
	case f.isStateStoreS3() || f.isStateStoreS3Compatible():
		return f.envsForS3()
	case f.isStateStoreGCS():
		return f.envsForGCS()
	case f.isStateStoreAliyunOSS():
		return f.envsForAliyunOSS()
	case f.isStateStoreAzureBlob():
		return f.envsForAzureBlob()
	case f.isStateStoreHDFS():
		return f.envsForHDFS()
	case f.isStateStoreWebHDFS():
		return f.envsForWebHDFS()
	case f.isStateStoreLocalDisk():
		return f.envsForLocalDisk()
	case f.isStateStoreHuaweiCloudOBS():
		return f.envsForHuaweiCloudOBS()
	default:
		return nil
	}
}

func (f *RisingWaveObjectFactory) risingWaveConfigVolume(nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup) corev1.Volume {
	configSrc := &f.risingwave.Spec.Configuration.RisingWaveNodeConfiguration
	if nodeGroup.Configuration != nil {
		configSrc = nodeGroup.Configuration
	}

	if configSrc.ConfigMap != nil {
		return corev1.Volume{
			Name: risingWaveConfigVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configSrc.ConfigMap.Name,
					},
					Items: []corev1.KeyToPath{
						{
							Key:  configSrc.ConfigMap.Key,
							Path: risingwaveConfigFileName,
						},
					},
					Optional: configSrc.ConfigMap.Optional,
				},
			},
		}
	} else if configSrc.Secret != nil {
		return corev1.Volume{
			Name: risingWaveConfigVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: configSrc.Secret.Name,
					Items: []corev1.KeyToPath{
						{
							Key:  configSrc.Secret.Key,
							Path: risingwaveConfigFileName,
						},
					},
					Optional: configSrc.Secret.Optional,
				},
			},
		}
	}

	return corev1.Volume{
		Name: risingWaveConfigVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: f.componentName(consts.ComponentConfig, ""),
				},
				Items: []corev1.KeyToPath{
					{
						Key:  risingWaveConfigMapKey,
						Path: risingwaveConfigFileName,
					},
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) volumeMountForConfig() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      risingWaveConfigVolume,
		MountPath: risingwaveConfigMountPath,
		ReadOnly:  true,
	}
}

func captureInheritedLabels(risingwave *risingwavev1alpha1.RisingWave) map[string]string {
	inheritLabelPrefix, exist := risingwave.Annotations[consts.AnnotationInheritLabelPrefix]
	if !exist {
		return nil
	}

	// Parse the label prefixes (separated by comma) from the annotation value.
	prefixes := strings.Split(inheritLabelPrefix, ",")
	for i, prefix := range prefixes {
		prefixes[i] = strings.TrimSpace(prefix)
	}
	prefixes = lo.Filter(prefixes, func(s string, _ int) bool {
		return len(s) > 0 && s != "risingwave"
	})

	if len(prefixes) == 0 {
		return nil
	}

	// Match labels with naive algorithm here.
	matchLabelKey := func(s string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(s, p+"/") {
				return true
			}
		}
		return false
	}

	inheritedLabels := make(map[string]string)
	for k, v := range risingwave.Labels {
		if matchLabelKey(k) {
			inheritedLabels[k] = v
		}
	}

	if len(inheritedLabels) == 0 {
		return nil
	}

	return inheritedLabels
}

func (f *RisingWaveObjectFactory) getInheritedLabels() map[string]string {
	return f.inheritedLabels
}

func (f *RisingWaveObjectFactory) portsForMetaContainer() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.MetaServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.MetaMetricsPort,
		},
		{
			Name:          consts.PortDashboard,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.MetaDashboardPort,
		},
	}
}

func basicSetupRisingWaveContainer(container *corev1.Container, component *risingwavev1alpha1.RisingWaveComponent) {
	if component == nil {
		component = &risingwavev1alpha1.RisingWaveComponent{
			LogLevel: "INFO",
		}
	}

	// Set the default executable path.
	container.Command = []string{risingwaveExecutablePath}

	// Set the RUST_LOG to log level.
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name:  envs.RustLog,
		Value: component.LogLevel,
	}, func(e *corev1.EnvVar) bool { return e.Name == envs.RustLog })

	// Setting the system environment variables.
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name: envs.PodIP,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	}, func(env *corev1.EnvVar) bool { return env.Name == envs.PodIP })
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name: envs.PodName,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	}, func(env *corev1.EnvVar) bool { return env.Name == envs.PodName })

	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name: envs.PodNamespace,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	}, func(env *corev1.EnvVar) bool { return env.Name == envs.PodNamespace })

	// Set RUST_MIN_STACK to 4M to allow more stack space for threads.
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name:  envs.RustMinStack,
		Value: strconv.Itoa(4 << 20),
	}, func(e *corev1.EnvVar) bool { return e.Name == envs.RustMinStack })

	// Set RUST_BACKTRACE=1 if printing stack traces is enabled.
	if !ptr.Deref(component.DisallowPrintStackTraces, false) {
		container.Env = mergeListByKey(container.Env, corev1.EnvVar{
			Name:  envs.RustBacktrace,
			Value: "full",
		}, func(env *corev1.EnvVar) bool { return env.Name == envs.RustBacktrace })
	}

	// Set the RW_WORKER_THREADS based on the cpu limit.
	// Always ensure that the worker threads are at least 4 to reduce the risk of
	// unexpected blocking by synchronous operations.
	// https://github.com/risingwavelabs/risingwave/issues/16693
	if cpuLimit, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
		container.Env = mergeListByKey(container.Env, corev1.EnvVar{
			Name:  envs.RWWorkerThreads,
			Value: strconv.FormatInt(max(cpuLimit.Value(), 4), 10),
		}, func(env *corev1.EnvVar) bool { return env.Name == envs.RWWorkerThreads })
	}

	// Set the default probes.
	container.StartupProbe = &corev1.Probe{
		InitialDelaySeconds: 5,
		PeriodSeconds:       5,
		TimeoutSeconds:      5,
		FailureThreshold:    12,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(consts.PortService),
			},
		},
	}
	container.LivenessProbe = &corev1.Probe{
		InitialDelaySeconds: 2,
		PeriodSeconds:       10,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(consts.PortService),
			},
		},
	}
	container.ReadinessProbe = &corev1.Probe{
		InitialDelaySeconds: 2,
		PeriodSeconds:       10,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(consts.PortService),
			},
		},
	}

	// Set the default container settings.
	container.Stdin = false
	container.StdinOnce = false
	container.TTY = false
}

func (f *RisingWaveObjectFactory) setupMetaContainer(container *corev1.Container) {
	container.Name = "meta"
	container.Args = []string{"meta-node"}
	container.Ports = f.portsForMetaContainer()

	// merge the vars, and use the container env vars(set in template) to replace the generated default vars, if key equals.
	mergedVars := f.envsForMetaArgs(container.Image)
	for _, env := range container.Env {
		mergedVars = mergeListWhenKeyEquals(mergedVars, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}
	container.Env = mergedVars

	for _, env := range f.envsForStateStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}

	for _, env := range f.envsForMetaStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

func rollingUpdateOrDefault(rollingUpdate *risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate) risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate {
	if rollingUpdate != nil {
		return *rollingUpdate
	}
	return risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{}
}

func buildUpgradeStrategyForDeployment(strategy risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy) appsv1.DeploymentStrategy {
	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate:
		return appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: rollingUpdateOrDefault(strategy.RollingUpdate).MaxUnavailable,
			},
		}
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate:
		return appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		}
	default:
		return appsv1.DeploymentStrategy{}
	}
}

func inPlaceUpdateStrategyOrDefault(strategy *kruisepubs.InPlaceUpdateStrategy) *kruisepubs.InPlaceUpdateStrategy {
	if strategy != nil {
		return strategy
	}
	return &kruisepubs.InPlaceUpdateStrategy{}
}

func buildUpgradeStrategyForCloneSet(strategy risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy) kruiseappsv1alpha1.CloneSetUpdateStrategy {
	cloneSetUpdateStrategy := kruiseappsv1alpha1.CloneSetUpdateStrategy{}

	rollingUpdateStrategy := rollingUpdateOrDefault(strategy.RollingUpdate)
	cloneSetUpdateStrategy.MaxUnavailable = rollingUpdateStrategy.MaxUnavailable
	cloneSetUpdateStrategy.MaxSurge = rollingUpdateStrategy.MaxSurge

	cloneSetUpdateStrategy.Partition = rollingUpdateStrategy.Partition
	if strategy.InPlaceUpdateStrategy != nil {
		cloneSetUpdateStrategy.InPlaceUpdateStrategy = inPlaceUpdateStrategyOrDefault(strategy.InPlaceUpdateStrategy)
	}

	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate:
		cloneSetUpdateStrategy.Type = kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible:
		cloneSetUpdateStrategy.Type = kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly:
		cloneSetUpdateStrategy.Type = kruiseappsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType
	default:
		return kruiseappsv1alpha1.CloneSetUpdateStrategy{}
	}

	return cloneSetUpdateStrategy
}

func newPodSpecFromNodeGroupTemplate(template *risingwavev1alpha1.RisingWaveNodePodTemplate) corev1.PodTemplateSpec {
	podTemplateSpec := corev1.PodTemplateSpec{}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{
		Labels:      template.ObjectMeta.Labels,
		Annotations: template.ObjectMeta.Annotations,
	}

	podTemplateSpec.Spec = corev1.PodSpec{
		Containers: append([]corev1.Container{
			{
				Image:           template.Spec.Image,
				ImagePullPolicy: template.Spec.ImagePullPolicy,
				EnvFrom:         template.Spec.EnvFrom,
				Env:             template.Spec.Env,
				Resources:       template.Spec.Resources,
				VolumeMounts:    template.Spec.VolumeMounts,
				VolumeDevices:   template.Spec.VolumeDevices,
				SecurityContext: template.Spec.RisingWaveNodeContainer.SecurityContext,
			},
		}, template.Spec.AdditionalContainers...),
		EnableServiceLinks:            ptr.To(false),
		Volumes:                       template.Spec.Volumes,
		ActiveDeadlineSeconds:         template.Spec.ActiveDeadlineSeconds,
		TerminationGracePeriodSeconds: template.Spec.TerminationGracePeriodSeconds,
		DNSPolicy:                     template.Spec.DNSPolicy,
		NodeSelector:                  template.Spec.NodeSelector,
		ServiceAccountName:            template.Spec.ServiceAccountName,
		AutomountServiceAccountToken:  template.Spec.AutomountServiceAccountToken,
		HostPID:                       template.Spec.HostPID,
		HostIPC:                       template.Spec.HostIPC,
		ShareProcessNamespace:         template.Spec.ShareProcessNamespace,
		SecurityContext:               template.Spec.SecurityContext,
		ImagePullSecrets:              template.Spec.ImagePullSecrets,
		Affinity:                      template.Spec.Affinity,
		SchedulerName:                 template.Spec.SchedulerName,
		Tolerations:                   template.Spec.Tolerations,
		HostAliases:                   template.Spec.HostAliases,
		PriorityClassName:             template.Spec.PriorityClassName,
		Priority:                      template.Spec.Priority,
		DNSConfig:                     template.Spec.DNSConfig,
		RuntimeClassName:              template.Spec.RuntimeClassName,
		PreemptionPolicy:              template.Spec.PreemptionPolicy,
		TopologySpreadConstraints:     template.Spec.TopologySpreadConstraints,
		SetHostnameAsFQDN:             template.Spec.SetHostnameAsFQDN,
		OS:                            template.Spec.OS,
		HostUsers:                     template.Spec.HostUsers,
	}

	return podTemplateSpec
}

func (f *RisingWaveObjectFactory) buildPodTemplateFromNodeGroup(component string, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, setupRisingWaveContainer func(container *corev1.Container)) corev1.PodTemplateSpec {
	podTemplate := newPodSpecFromNodeGroupTemplate(&nodeGroup.Template)

	// Inject system labels.
	podTemplate.Labels = mergeMap(podTemplate.Labels, f.podLabelsOrSelectorsForComponentGroup(component, nodeGroup.Name))
	podTemplate.Labels = mergeMap(podTemplate.Labels, f.getInheritedLabels())

	// Inject restart at annotation.
	if nodeGroup.RestartAt != nil {
		podTemplate.Annotations = mergeMap(podTemplate.Annotations, map[string]string{
			consts.AnnotationRestartAt: nodeGroup.RestartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
		})
	}

	// Inject RisingWave's config volume.
	podTemplate.Spec.Volumes = mergeListWhenKeyEquals(podTemplate.Spec.Volumes, f.risingWaveConfigVolume(nodeGroup), func(a, b *corev1.Volume) bool {
		return a.Name == b.Name
	})

	// Run container setup for RisingWave's container.
	setupRisingWaveContainer(&podTemplate.Spec.Containers[0])

	// Keep the pod spec consistent.
	keepPodSpecConsistent(&podTemplate.Spec)

	return podTemplate
}

func (f *RisingWaveObjectFactory) newDeployment(component string, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, template *corev1.PodTemplateSpec) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: f.getObjectMetaForComponentGroupLevelResources(component, nodeGroup.Name, true),
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(nodeGroup.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectorsForComponentGroup(component, nodeGroup.Name),
			},
			Template:                *template,
			Strategy:                buildUpgradeStrategyForDeployment(nodeGroup.UpgradeStrategy),
			MinReadySeconds:         nodeGroup.MinReadySeconds,
			ProgressDeadlineSeconds: nodeGroup.ProgressDeadlineSeconds,
		},
	}
}

func (f *RisingWaveObjectFactory) newCloneSet(component string, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, template *corev1.PodTemplateSpec) *kruiseappsv1alpha1.CloneSet {
	// Inject readiness gate for in place update strategy.
	template.Spec.ReadinessGates = append(template.Spec.ReadinessGates, corev1.PodReadinessGate{
		ConditionType: kruisepubs.InPlaceUpdateReady,
	})

	return &kruiseappsv1alpha1.CloneSet{
		ObjectMeta: f.getObjectMetaForComponentGroupLevelResources(component, nodeGroup.Name, true),
		Spec: kruiseappsv1alpha1.CloneSetSpec{
			Replicas: ptr.To(nodeGroup.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectorsForComponentGroup(component, nodeGroup.Name),
			},
			Template:        *template,
			UpdateStrategy:  buildUpgradeStrategyForCloneSet(nodeGroup.UpgradeStrategy),
			MinReadySeconds: nodeGroup.MinReadySeconds,
		},
	}
}

func buildPersistentVolumeClaims(claims []risingwavev1alpha1.PersistentVolumeClaim) []corev1.PersistentVolumeClaim {
	if claims == nil {
		return nil
	}

	result := make([]corev1.PersistentVolumeClaim, 0, len(claims))
	for _, claim := range claims {
		claim := *claim.DeepCopy()
		result = append(result, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        claim.Name,
				Labels:      claim.Labels,
				Annotations: claim.Annotations,
				Finalizers:  claim.Finalizers,
			},
			Spec: claim.Spec,
		})
	}

	return result
}

func (f *RisingWaveObjectFactory) newStatefulSet(component string, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, template *corev1.PodTemplateSpec) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: f.getObjectMetaForComponentGroupLevelResources(component, nodeGroup.Name, true),
		Spec: appsv1.StatefulSetSpec{
			ServiceName:    f.componentName(component, ""),
			Replicas:       ptr.To(nodeGroup.Replicas),
			UpdateStrategy: buildUpgradeStrategyForStatefulSet(nodeGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectorsForComponentGroup(component, nodeGroup.Name),
			},
			Template:                             *template,
			PodManagementPolicy:                  appsv1.ParallelPodManagement,
			MinReadySeconds:                      nodeGroup.MinReadySeconds,
			VolumeClaimTemplates:                 buildPersistentVolumeClaims(nodeGroup.VolumeClaimTemplates),
			PersistentVolumeClaimRetentionPolicy: nodeGroup.PersistentVolumeClaimRetentionPolicy,
		},
	}
}

func (f *RisingWaveObjectFactory) newAdvancedStatefulSet(component string, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, template *corev1.PodTemplateSpec) *kruiseappsv1beta1.StatefulSet {
	// Inject readiness gate for in place update strategy.
	template.Spec.ReadinessGates = append(template.Spec.ReadinessGates, corev1.PodReadinessGate{
		ConditionType: kruisepubs.InPlaceUpdateReady,
	})

	return &kruiseappsv1beta1.StatefulSet{
		ObjectMeta: f.getObjectMetaForComponentGroupLevelResources(component, nodeGroup.Name, true),
		Spec: kruiseappsv1beta1.StatefulSetSpec{
			Replicas:       ptr.To(nodeGroup.Replicas),
			ServiceName:    f.componentName(component, ""),
			UpdateStrategy: buildUpgradeStrategyForAdvancedStatefulSet(nodeGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectorsForComponentGroup(component, nodeGroup.Name),
			},
			Template:                             *template,
			PodManagementPolicy:                  appsv1.ParallelPodManagement,
			VolumeClaimTemplates:                 buildPersistentVolumeClaims(nodeGroup.VolumeClaimTemplates),
			PersistentVolumeClaimRetentionPolicy: convertAppsV1StatefulSetPersistentVolumeClaimRetentionPolicyToKruise(nodeGroup.PersistentVolumeClaimRetentionPolicy),
		},
	}
}

func (f *RisingWaveObjectFactory) convertStandaloneSpecIntoComponent() *risingwavev1alpha1.RisingWaveComponent {
	standaloneSpec := f.risingwave.Spec.Components.Standalone
	if standaloneSpec != nil {
		return &risingwavev1alpha1.RisingWaveComponent{
			LogLevel:                 standaloneSpec.LogLevel,
			DisallowPrintStackTraces: standaloneSpec.DisallowPrintStackTraces,
		}
	}

	return &risingwavev1alpha1.RisingWaveComponent{
		LogLevel: "INFO",
	}
}

func (f *RisingWaveObjectFactory) newPodTemplateSpecFromNodeGroupByComponent(component string, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup) corev1.PodTemplateSpec {
	var containerModifier func(container *corev1.Container)
	var componentPtr *risingwavev1alpha1.RisingWaveComponent
	switch component {
	case consts.ComponentMeta:
		containerModifier = f.setupMetaContainer
		componentPtr = &f.risingwave.Spec.Components.Meta
	case consts.ComponentFrontend:
		containerModifier = f.setupFrontendContainer
		componentPtr = &f.risingwave.Spec.Components.Frontend
	case consts.ComponentCompactor:
		containerModifier = f.setupCompactorContainer
		componentPtr = &f.risingwave.Spec.Components.Compactor
	case consts.ComponentCompute:
		containerModifier = f.setupComputeContainer
		componentPtr = &f.risingwave.Spec.Components.Compute
	case consts.ComponentStandalone:
		containerModifier = f.setupStandaloneContainer
		componentPtr = f.convertStandaloneSpecIntoComponent()
	default:
		panic("invalid component")
	}

	return f.buildPodTemplateFromNodeGroup(component, nodeGroup, func(container *corev1.Container) {
		basicSetupRisingWaveContainer(container, componentPtr)
		containerModifier(container)
	})
}

func (f *RisingWaveObjectFactory) overrideFieldsOfNodeGroup(nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup) *risingwavev1alpha1.RisingWaveNodeGroup {
	if nodeGroup.Template.Spec.Image == "" {
		nodeGroup.Template.Spec.Image = f.risingwave.Spec.Image
	}
	return nodeGroup
}

func newWorkloadObjectForComponentNodeGroup[T client.Object](f *RisingWaveObjectFactory, component, group string, builder func(component string, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, template *corev1.PodTemplateSpec) T) T {
	nodeGroup := object.NewRisingWaveReader(f.risingwave).GetNodeGroup(component, group)
	template := f.newPodTemplateSpecFromNodeGroupByComponent(component, f.overrideFieldsOfNodeGroup(nodeGroup))
	workloadObj := builder(component, nodeGroup, &template)
	return mustSetControllerReference(f.risingwave, workloadObj, f.scheme)
}

func (f *RisingWaveObjectFactory) portsForFrontendContainer() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.FrontendServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.FrontendMetricsPort,
		},
	}
}

func (f *RisingWaveObjectFactory) isEmbeddedServingModeEnabled() bool {
	return ptr.Deref(f.risingwave.Spec.EnableEmbeddedServingMode, false)
}

func (f *RisingWaveObjectFactory) setupFrontendContainer(container *corev1.Container) {
	container.Name = "frontend"

	if f.isEmbeddedServingModeEnabled() {
		container.Args = []string{
			"standalone",
			fmt.Sprintf("--compute-opts=--listen-addr 0.0.0.0:%d --advertise-addr $(%s):%d --role=serving",
				consts.ComputeServicePort, envs.PodIP, consts.ComputeServicePort),
			fmt.Sprintf("--frontend-opts=--listen-addr 0.0.0.0:%d --advertise-addr $(%s):%d",
				consts.FrontendServicePort, envs.PodIP, consts.FrontendServicePort),
		}
		container.Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"bash",
						"-c",
						fmt.Sprintf("%s ctl meta unregister-workers --yes --workers ${%s}:%d",
							risingwaveExecutablePath, envs.PodIP, consts.ComputeServicePort),
					},
				},
			},
		}
	} else {
		container.Args = []string{"frontend-node"}
	}

	// merge the vars, and use the container env vars(set in template) to replace the generated default vars, if key equals.
	mergedVars := f.envsForFrontendArgs()

	if f.isEmbeddedServingModeEnabled() {
		cpuLimit := int64(math.Ceil(container.Resources.Limits.Cpu().AsApproximateFloat64()))
		memLimit, _ := container.Resources.Limits.Memory().AsInt64()
		mergedVars = append(mergedVars, f.envsForResourceLimits(cpuLimit, memLimit)...)
		mergedVars = append(mergedVars, f.envsForStateStore()...)
	}

	for _, env := range container.Env {
		mergedVars = mergeListWhenKeyEquals(mergedVars, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}
	container.Env = mergedVars
	container.Ports = f.portsForFrontendContainer()

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

func (f *RisingWaveObjectFactory) portsForComputeContainer() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.ComputeServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.ComputeMetricsPort,
		},
	}
}

func (f *RisingWaveObjectFactory) setupComputeContainer(container *corev1.Container) {
	container.Name = "compute"
	container.Args = []string{"compute-node"}

	if f.isEmbeddedServingModeEnabled() {
		container.Args = append(container.Args, "--role=streaming")
	}

	cpuLimit := int64(math.Ceil(container.Resources.Limits.Cpu().AsApproximateFloat64()))
	memLimit, _ := container.Resources.Limits.Memory().AsInt64()

	// merge the vars, and use the container env vars(set in template) to replace the generated default vars, if key equals.
	mergedVars := f.envsForComputeArgs(cpuLimit, memLimit)
	for _, env := range container.Env {
		mergedVars = mergeListWhenKeyEquals(mergedVars, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}
	container.Env = mergedVars
	container.Ports = f.portsForComputeContainer()

	for _, env := range f.envsForStateStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

func (f *RisingWaveObjectFactory) portsForCompactorContainer() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.CompactorServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.CompactorMetricsPort,
		},
	}
}

func (f *RisingWaveObjectFactory) setupCompactorContainer(container *corev1.Container) {
	container.Name = "compactor"
	container.Args = []string{"compactor-node"}

	// merge the vars, and use the container env vars(set in template) to replace the generated default vars, if key equals.
	mergedVars := f.envsForCompactorArgs()
	for _, env := range container.Env {
		mergedVars = mergeListWhenKeyEquals(mergedVars, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}
	container.Env = mergedVars
	container.Ports = f.portsForCompactorContainer()

	for _, env := range f.envsForStateStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

func (f *RisingWaveObjectFactory) argsForStandaloneMetaStore() []string {
	args := []string{
		"--backend $(RW_BACKEND)",
	}

	if f.isMetaStoreEtcd() {
		args = append(args, "--etcd-endpoints $(RW_ETCD_ENDPOINTS)")

		credentials := f.risingwave.Spec.MetaStore.Etcd.RisingWaveEtcdCredentials
		if credentials != nil && credentials.SecretName != "" {
			args = append(args, "--etcd-auth",
				"--etcd-username $(RW_ETCD_USERNAME)",
				"--etcd-password $(RW_ETCD_PASSWORD)")
		}
	}

	return args
}

func (f *RisingWaveObjectFactory) argsForStandalone() []string {
	return []string{
		"--config-path=$(RW_CONFIG_PATH)",
		fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", consts.MetaMetricsPort),
		"--meta-opts=" + strings.Join(append(
			[]string{
				fmt.Sprintf("--listen-addr 127.0.0.1:%d", consts.MetaServicePort),
				fmt.Sprintf("--advertise-addr 127.0.0.1:%d", consts.MetaServicePort),
				fmt.Sprintf("--dashboard-host 0.0.0.0:%d", consts.MetaDashboardPort),
				fmt.Sprintf("--prometheus-host 0.0.0.0:%d", consts.MetaMetricsPort),
				"--state-store $(RW_STATE_STORE)",
				"--data-directory $(RW_DATA_DIRECTORY)",
			},
			// Arguments for meta stores.
			f.argsForStandaloneMetaStore()...,
		), " "),
		"--compute-opts=" + strings.Join([]string{
			fmt.Sprintf("--listen-addr 127.0.0.1:%d", consts.ComputeServicePort),
			fmt.Sprintf("--advertise-addr 127.0.0.1:%d", consts.ComputeServicePort),
			fmt.Sprintf("--meta-address http://127.0.0.1:%d", consts.MetaServicePort),
		}, " "),
		"--frontend-opts=" + strings.Join([]string{
			fmt.Sprintf("--listen-addr 0.0.0.0:%d", consts.FrontendServicePort),
			fmt.Sprintf("--advertise-addr 127.0.0.1:%d", consts.FrontendServicePort),
			fmt.Sprintf("--meta-addr http://127.0.0.1:%d", consts.MetaServicePort),
		}, " "),
		"--compactor-opts=" + strings.Join([]string{
			fmt.Sprintf("--listen-addr 127.0.0.1:%d", consts.CompactorServicePort),
			fmt.Sprintf("--advertise-addr 127.0.0.1:%d", consts.CompactorServicePort),
			fmt.Sprintf("--meta-address http://127.0.0.1:%d", consts.MetaServicePort),
		}, " "),
	}
}

func (f *RisingWaveObjectFactory) argsForStandaloneV2() []string {
	return []string{
		"--config-path=$(RW_CONFIG_PATH)",
		fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", consts.MetaMetricsPort),
		fmt.Sprintf("--listen-addr=0.0.0.0:%d", consts.FrontendServicePort),
	}
}

func (f *RisingWaveObjectFactory) portsForStandaloneContainer() []corev1.ContainerPort {
	// TODO: either expose each metrics port, or combine them into one.
	return []corev1.ContainerPort{
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.MetaMetricsPort,
		},
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.FrontendServicePort,
		},
		{
			Name:          consts.PortDashboard,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: consts.MetaDashboardPort,
		},
	}
}

func (f *RisingWaveObjectFactory) setupStandaloneContainer(container *corev1.Container) {
	container.Name = "standalone"

	var imageVersion = utils.GetVersionFromImage(f.risingwave.Spec.Image)

	var useV2Style = false
	switch f.risingwave.Spec.StandaloneMode {
	case 0:
		// `standalone` subcommand is only supported in versions 1.3 - 1.7, and was deprecated in 1.8.
		useV2Style = !(strings.HasPrefix(imageVersion, "v1.3.") ||
			strings.HasPrefix(imageVersion, "v1.4.") ||
			strings.HasPrefix(imageVersion, "v1.5.") ||
			strings.HasPrefix(imageVersion, "v1.6.") ||
			strings.HasPrefix(imageVersion, "v1.7."))
	case 1:
		useV2Style = false
	case 2:
		useV2Style = true
	}

	if !useV2Style {
		container.Command = []string{
			risingwaveExecutablePath,
			"standalone",
		}
		container.Args = f.argsForStandalone()
	} else {
		container.Command = []string{
			risingwaveExecutablePath,
			"single-node",
		}
		container.Args = f.argsForStandaloneV2()
	}

	container.Ports = f.portsForStandaloneContainer()

	container.Env = append(container.Env, f.envsForTLS()...)

	container.Env = mergeListWhenKeyEquals(container.Env, corev1.EnvVar{
		Name:  envs.RWConfigPath,
		Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
	}, func(a, b *corev1.EnvVar) bool { return a.Name == b.Name })

	for _, env := range f.coreEnvsForMeta(container.Image) {
		container.Env = mergeListWhenKeyEquals(container.Env, env,
			func(a, b *corev1.EnvVar) bool { return a.Name == b.Name })
	}

	// Env var for local risectl access.
	container.Env = mergeListWhenKeyEquals(container.Env, corev1.EnvVar{
		Name:  envs.RWMetaAddr,
		Value: fmt.Sprintf("http://127.0.0.1:%d", consts.MetaServicePort),
	}, func(a, b *corev1.EnvVar) bool { return a.Name == b.Name })

	for _, env := range f.envsForStateStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env,
			func(a, b *corev1.EnvVar) bool { return a.Name == b.Name })
	}

	for _, env := range f.envsForMetaStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env,
			func(a, b *corev1.EnvVar) bool { return a.Name == b.Name })
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(),
		func(a, b *corev1.VolumeMount) bool { return a.MountPath == b.MountPath })
}

func (f *RisingWaveObjectFactory) convertStandaloneIntoNodeGroup() *risingwavev1alpha1.RisingWaveNodeGroup {
	standaloneSpec := f.risingwave.Spec.Components.Standalone
	if standaloneSpec == nil {
		return &risingwavev1alpha1.RisingWaveNodeGroup{
			Replicas: 1,
		}
	}

	return &risingwavev1alpha1.RisingWaveNodeGroup{
		Replicas:                             standaloneSpec.Replicas,
		RestartAt:                            standaloneSpec.RestartAt,
		UpgradeStrategy:                      standaloneSpec.UpgradeStrategy,
		MinReadySeconds:                      standaloneSpec.MinReadySeconds,
		VolumeClaimTemplates:                 standaloneSpec.VolumeClaimTemplates,
		PersistentVolumeClaimRetentionPolicy: standaloneSpec.PersistentVolumeClaimRetentionPolicy,
		ProgressDeadlineSeconds:              standaloneSpec.ProgressDeadlineSeconds,
		Template:                             standaloneSpec.Template,
	}
}

// NewStandaloneStatefulSet creates a StatefulSet for standalone component.
func (f *RisingWaveObjectFactory) NewStandaloneStatefulSet() *appsv1.StatefulSet {
	nodeGroup := f.convertStandaloneIntoNodeGroup()
	template := f.newPodTemplateSpecFromNodeGroupByComponent(consts.ComponentStandalone, f.overrideFieldsOfNodeGroup(nodeGroup))
	workloadObj := f.newStatefulSet(consts.ComponentStandalone, nodeGroup, &template)
	return mustSetControllerReference(f.risingwave, workloadObj, f.scheme)
}

// NewStandaloneAdvancedStatefulSet creates an advanced StatefulSet for standalone component.
func (f *RisingWaveObjectFactory) NewStandaloneAdvancedStatefulSet() *kruiseappsv1beta1.StatefulSet {
	nodeGroup := f.convertStandaloneIntoNodeGroup()
	template := f.newPodTemplateSpecFromNodeGroupByComponent(consts.ComponentStandalone, f.overrideFieldsOfNodeGroup(nodeGroup))
	workloadObj := f.newAdvancedStatefulSet(consts.ComponentStandalone, nodeGroup, &template)
	return mustSetControllerReference(f.risingwave, workloadObj, f.scheme)
}

// NewStandaloneService creates a Service for standalone component.
func (f *RisingWaveObjectFactory) NewStandaloneService() *corev1.Service {
	standaloneSvc := f.newService(consts.ComponentStandalone, corev1.ServiceTypeClusterIP, []corev1.ServicePort{
		{
			Name:       consts.PortService,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.FrontendServicePort,
			TargetPort: intstr.FromString(consts.PortService),
		},
		{
			Name:       consts.PortMetrics,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.MetaMetricsPort,
			TargetPort: intstr.FromString(consts.PortMetrics),
		},
		{
			Name:       consts.PortDashboard,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.MetaDashboardPort,
			TargetPort: intstr.FromString(consts.PortDashboard),
		},
	})

	return mustSetControllerReference(f.risingwave, standaloneSvc, f.scheme)
}

// NewMetaService creates a new Service for the meta.
func (f *RisingWaveObjectFactory) NewMetaService() *corev1.Service {
	metaSvc := f.newService(consts.ComponentMeta, corev1.ServiceTypeClusterIP, []corev1.ServicePort{
		{
			Name:       consts.PortService,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.MetaServicePort,
			TargetPort: intstr.FromString(consts.PortService),
		},
		{
			Name:       consts.PortMetrics,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.MetaMetricsPort,
			TargetPort: intstr.FromString(consts.PortMetrics),
		},
		{
			Name:       consts.PortDashboard,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.MetaDashboardPort,
			TargetPort: intstr.FromString(consts.PortDashboard),
		},
	})

	// Set the ClusterIP to None to make it a headless service.
	metaSvc.Spec.ClusterIP = corev1.ClusterIPNone

	return mustSetControllerReference(f.risingwave, metaSvc, f.scheme)
}

// NewFrontendService creates a new Service for the frontend.
func (f *RisingWaveObjectFactory) NewFrontendService() *corev1.Service {
	svcPorts := []corev1.ServicePort{
		{
			Name:       consts.PortService,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.FrontendServicePort,
			TargetPort: intstr.FromString(consts.PortService),
		},
	}

	// Disable metrics port when RisingWave is in standalone mode. This makes it
	// easier to write a single telemetry configuration to cover all running RisingWaves.
	if !ptr.Deref(f.risingwave.Spec.EnableStandaloneMode, false) {
		svcPorts = append(svcPorts, corev1.ServicePort{
			Name:       consts.PortMetrics,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.FrontendMetricsPort,
			TargetPort: intstr.FromString(consts.PortMetrics),
		})
	}

	frontendSvc := f.newService(consts.ComponentFrontend, f.risingwave.Spec.FrontendServiceType, svcPorts)

	// Hijack selector if it's standalone mode.
	if object.NewRisingWaveReader(f.risingwave).IsStandaloneModeEnabled() {
		frontendSvc.Spec.Selector[consts.LabelRisingWaveComponent] = consts.ComponentStandalone
	}

	// Inject additional metadata.
	frontendSvc.ObjectMeta.Labels = mergeMap(frontendSvc.ObjectMeta.Labels, f.risingwave.Spec.AdditionalFrontendServiceMetadata.Labels)
	frontendSvc.ObjectMeta.Annotations = mergeMap(frontendSvc.ObjectMeta.Annotations, f.risingwave.Spec.AdditionalFrontendServiceMetadata.Annotations)

	return mustSetControllerReference(f.risingwave, frontendSvc, f.scheme)
}

// NewComputeService creates a new Service for the compute nodes.
func (f *RisingWaveObjectFactory) NewComputeService() *corev1.Service {
	computeSvc := f.newService(consts.ComponentCompute, corev1.ServiceTypeClusterIP, []corev1.ServicePort{
		{
			Name:       consts.PortService,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.ComputeServicePort,
			TargetPort: intstr.FromString(consts.PortService),
		},
		{
			Name:       consts.PortMetrics,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.ComputeMetricsPort,
			TargetPort: intstr.FromString(consts.PortMetrics),
		},
	})

	// Set the ClusterIP to None to make it a headless service.
	computeSvc.Spec.ClusterIP = corev1.ClusterIPNone

	return mustSetControllerReference(f.risingwave, computeSvc, f.scheme)
}

// NewCompactorService creates a new Service for the compactor.
func (f *RisingWaveObjectFactory) NewCompactorService() *corev1.Service {
	compactorSvc := f.newService(consts.ComponentCompactor, corev1.ServiceTypeClusterIP, []corev1.ServicePort{
		{
			Name:       consts.PortService,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.CompactorServicePort,
			TargetPort: intstr.FromString(consts.PortService),
		},
		{
			Name:       consts.PortMetrics,
			Protocol:   corev1.ProtocolTCP,
			Port:       consts.CompactorMetricsPort,
			TargetPort: intstr.FromString(consts.PortMetrics),
		},
	})

	// Set the ClusterIP to None to make it a headless service.
	compactorSvc.Spec.ClusterIP = corev1.ClusterIPNone

	return mustSetControllerReference(f.risingwave, compactorSvc, f.scheme)
}

// NewMetaStatefulSet creates a new StatefulSet for the meta component and specified group.
func (f *RisingWaveObjectFactory) NewMetaStatefulSet(group string) *appsv1.StatefulSet {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentMeta, group, f.newStatefulSet)
}

func convertAppsV1StatefulSetPersistentVolumeClaimRetentionPolicyToKruise(retentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy) *kruiseappsv1beta1.StatefulSetPersistentVolumeClaimRetentionPolicy {
	if retentionPolicy == nil {
		return nil
	}
	return &kruiseappsv1beta1.StatefulSetPersistentVolumeClaimRetentionPolicy{
		WhenDeleted: kruiseappsv1beta1.PersistentVolumeClaimRetentionPolicyType(retentionPolicy.WhenDeleted),
		WhenScaled:  kruiseappsv1beta1.PersistentVolumeClaimRetentionPolicyType(retentionPolicy.WhenScaled),
	}
}

// NewMetaAdvancedStatefulSet creates a new OpenKruise StatefulSet for the meta component and specified group.
func (f *RisingWaveObjectFactory) NewMetaAdvancedStatefulSet(group string) *kruiseappsv1beta1.StatefulSet {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentMeta, group, f.newAdvancedStatefulSet)
}

// NewFrontendDeployment creates a new Deployment for the frontend component and specified group.
func (f *RisingWaveObjectFactory) NewFrontendDeployment(group string) *appsv1.Deployment {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentFrontend, group, f.newDeployment)
}

// NewFrontendCloneSet creates a new CloneSet for the frontend component and specified group.
func (f *RisingWaveObjectFactory) NewFrontendCloneSet(group string) *kruiseappsv1alpha1.CloneSet {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentFrontend, group, f.newCloneSet)
}

// NewCompactorDeployment creates a new Deployment for the compactor component and specified group.
func (f *RisingWaveObjectFactory) NewCompactorDeployment(group string) *appsv1.Deployment {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentCompactor, group, f.newDeployment)
}

// NewCompactorCloneSet creates a new CloneSet for the compactor component and specified group.
func (f *RisingWaveObjectFactory) NewCompactorCloneSet(group string) *kruiseappsv1alpha1.CloneSet {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentCompactor, group, f.newCloneSet)
}

func buildUpgradeStrategyForStatefulSet(strategy risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy) appsv1.StatefulSetUpdateStrategy {
	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate:
		return appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
			RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
				MaxUnavailable: rollingUpdateOrDefault(strategy.RollingUpdate).MaxUnavailable,
			},
		}
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate:
		return appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
			RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
				MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "100%"},
			},
		}
	default:
		return appsv1.StatefulSetUpdateStrategy{}
	}
}

func buildUpgradeStrategyForAdvancedStatefulSet(strategy risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy) kruiseappsv1beta1.StatefulSetUpdateStrategy {
	advancedStatefulSetUpgradeStrategy := kruiseappsv1beta1.StatefulSetUpdateStrategy{}
	advancedStatefulSetUpgradeStrategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
	advancedStatefulSetUpgradeStrategy.RollingUpdate = &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
		MaxUnavailable: rollingUpdateOrDefault(strategy.RollingUpdate).MaxUnavailable,
	}
	if strategy.InPlaceUpdateStrategy != nil {
		advancedStatefulSetUpgradeStrategy.RollingUpdate.InPlaceUpdateStrategy = strategy.InPlaceUpdateStrategy.DeepCopy()
	}

	if rollingUpdateOrDefault(strategy.RollingUpdate).Partition != nil {
		// Change a percentage to an integer, partition only accepts int pointers for advanced stateful sets
		if rollingUpdateOrDefault(strategy.RollingUpdate).Partition.Type != intstr.Int {
			intValue, err := strconv.Atoi(strings.Replace((strategy.RollingUpdate).Partition.StrVal, "%", "", -1))
			if err != nil {
				panic(err)
			}
			advancedStatefulSetUpgradeStrategy.RollingUpdate.Partition = ptr.To(int32(intValue))
		} else {
			advancedStatefulSetUpgradeStrategy.RollingUpdate.Partition = ptr.To(rollingUpdateOrDefault(strategy.RollingUpdate).Partition.IntVal)
		}
	}

	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate:
		advancedStatefulSetUpgradeStrategy.RollingUpdate.PodUpdatePolicy = kruiseappsv1beta1.RecreatePodUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible:
		advancedStatefulSetUpgradeStrategy.RollingUpdate.PodUpdatePolicy = kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly:
		advancedStatefulSetUpgradeStrategy.RollingUpdate.PodUpdatePolicy = kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType
	default:
		return kruiseappsv1beta1.StatefulSetUpdateStrategy{}
	}

	return advancedStatefulSetUpgradeStrategy
}

// NewComputeStatefulSet creates a new StatefulSet for the compute component and specified group.
func (f *RisingWaveObjectFactory) NewComputeStatefulSet(group string) *appsv1.StatefulSet {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentCompute, group, f.newStatefulSet)
}

// NewComputeAdvancedStatefulSet creates a new OpenKruise StatefulSet for the compute component and specified group.
func (f *RisingWaveObjectFactory) NewComputeAdvancedStatefulSet(group string) *kruiseappsv1beta1.StatefulSet {
	return newWorkloadObjectForComponentNodeGroup(f, consts.ComponentCompute, group, f.newAdvancedStatefulSet)
}

// NewConfigConfigMap creates a new ConfigMap with the specified string value for risingwave.toml.
func (f *RisingWaveObjectFactory) NewConfigConfigMap(val string) *corev1.ConfigMap {
	risingwaveConfigConfigMap := &corev1.ConfigMap{
		ObjectMeta: f.getObjectMetaForComponentLevelResources(consts.ComponentConfig, false), // not synced
		Data: map[string]string{
			risingWaveConfigMapKey: val,
		},
	}
	return mustSetControllerReference(f.risingwave, risingwaveConfigConfigMap, f.scheme)
}

// NewServiceMonitor creates a new ServiceMonitor.
func (f *RisingWaveObjectFactory) NewServiceMonitor() *prometheusv1.ServiceMonitor {
	const (
		interval      = 15 * time.Second
		scrapeTimeout = 15 * time.Second
	)

	serviceMonitor := &prometheusv1.ServiceMonitor{
		ObjectMeta: f.getObjectMetaForGeneralResources("risingwave-"+f.risingwave.Name, true),
		Spec: prometheusv1.ServiceMonitorSpec{
			JobLabel: "risingwave/" + f.risingwave.Name,
			TargetLabels: []string{
				consts.LabelRisingWaveName,
				consts.LabelRisingWaveComponent,
				consts.LabelRisingWaveGroup,
			},
			Endpoints: []prometheusv1.Endpoint{
				{
					Port:          consts.PortMetrics,
					Interval:      prometheusv1.Duration(fmt.Sprintf("%.0fs", interval.Seconds())),
					ScrapeTimeout: prometheusv1.Duration(fmt.Sprintf("%.0fs", scrapeTimeout.Seconds())),
					// we need to drop some metrics which maybe will produce too many series.
					MetricRelabelConfigs: []prometheusv1.RelabelConfig{
						{
							SourceLabels: []prometheusv1.LabelName{"__name__"},
							Action:       "drop",
							Regex:        "batch_.+",
						},
						{
							SourceLabels: []prometheusv1.LabelName{"__name__"},
							Action:       "drop",
							Regex:        "stream_exchange_.+",
						},
					},
				},
			},
			Selector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      consts.LabelRisingWaveName,
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{f.risingwave.Name},
					},
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, serviceMonitor, f.scheme)
}

func (f *RisingWaveObjectFactory) envsForHuaweiCloudOBS() []corev1.EnvVar {
	obs := f.risingwave.Spec.StateStore.HuaweiCloudOBS
	credentials := obs.RisingWaveHuaweiCloudOBSCredentials
	secretRef := corev1.LocalObjectReference{Name: credentials.SecretName}

	return []corev1.EnvVar{
		{
			Name:  envs.HuaweiCloudOBSRegion,
			Value: obs.Region,
		},
		{
			Name:  envs.HuaweiCloudOBSEndpoint,
			Value: huaweiCloudOBSEndpoint,
		},
		{
			Name: envs.HuaweiCloudOBSAccessKeyID,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccessKeyIDRef,
				},
			},
		},
		{
			Name: envs.HuaweiCloudOBSSecretAccessKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccessKeySecretRef,
				},
			},
		},
	}
}

// NewRisingWaveObjectFactory creates a new RisingWaveObjectFactory.
func NewRisingWaveObjectFactory(risingwave *risingwavev1alpha1.RisingWave, scheme *runtime.Scheme, operatorVersion string) *RisingWaveObjectFactory {
	return &RisingWaveObjectFactory{
		risingwave:      risingwave,
		scheme:          scheme,
		inheritedLabels: captureInheritedLabels(risingwave),
		operatorVersion: operatorVersion,
	}
}
