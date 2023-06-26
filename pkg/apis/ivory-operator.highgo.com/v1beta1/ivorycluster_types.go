/*
 Copyright 2021 - 2023 Highgo Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package v1beta1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IvoryClusterSpec defines the desired state of IvoryCluster
type IvoryClusterSpec struct {
	// +optional
	Metadata *Metadata `json:"metadata,omitempty"`

	// Specifies a data source for bootstrapping the IvorySQL cluster.
	// +optional
	DataSource *DataSource `json:"dataSource,omitempty"`

	// IvorySQL backup configuration
	// +kubebuilder:validation:Required
	Backups Backups `json:"backups"`

	// The secret containing the Certificates and Keys to encrypt IvorySQL
	// traffic will need to contain the server TLS certificate, TLS key and the
	// Certificate Authority certificate with the data keys set to tls.crt,
	// tls.key and ca.crt, respectively. It will then be mounted as a volume
	// projection to the '/pgconf/tls' directory. For more information on
	// Kubernetes secret projections, please see
	// https://k8s.io/docs/concepts/configuration/secret/#projection-of-secret-keys-to-specific-paths
	// NOTE: If CustomTLSSecret is provided, CustomReplicationClientTLSSecret
	// MUST be provided and the ca.crt provided must be the same.
	// +optional
	CustomTLSSecret *corev1.SecretProjection `json:"customTLSSecret,omitempty"`

	// The secret containing the replication client certificates and keys for
	// secure connections to the IvorySQL server. It will need to contain the
	// client TLS certificate, TLS key and the Certificate Authority certificate
	// with the data keys set to tls.crt, tls.key and ca.crt, respectively.
	// NOTE: If CustomReplicationClientTLSSecret is provided, CustomTLSSecret
	// MUST be provided and the ca.crt provided must be the same.
	// +optional
	CustomReplicationClientTLSSecret *corev1.SecretProjection `json:"customReplicationTLSSecret,omitempty"`

	// DatabaseInitSQL defines a ConfigMap containing custom SQL that will
	// be run after the cluster is initialized. This ConfigMap must be in the same
	// namespace as the cluster.
	// +optional
	DatabaseInitSQL *DatabaseInitSQL `json:"databaseInitSQL,omitempty"`
	// Whether or not the IvorySQL cluster should use the defined default
	// scheduling constraints. If the field is unset or false, the default
	// scheduling constraints will be used in addition to any custom constraints
	// provided.
	// +optional
	DisableDefaultPodScheduling *bool `json:"disableDefaultPodScheduling,omitempty"`

	// The image name to use for IvorySQL containers. When omitted, the value
	// comes from an operator environment variable. For standard IvorySQL images,
	// the format is RELATED_IMAGE_IVORY_{ivoryVersion},
	// e.g. RELATED_IMAGE_IVORY_13. For PostGIS enabled IvorySQL images,
	// the format is RELATED_IMAGE_IVORY_{ivoryVersion}_GIS_{postGISVersion},
	// e.g. RELATED_IMAGE_IVORY_13_GIS_3.1.
	// +optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=1
	Image string `json:"image,omitempty"`

	// ImagePullPolicy is used to determine when Kubernetes will attempt to
	// pull (download) container images.
	// More info: https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy
	// +kubebuilder:validation:Enum={Always,Never,IfNotPresent}
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// The image pull secrets used to pull from a private registry
	// Changing this value causes all running pods to restart.
	// https://k8s.io/docs/tasks/configure-pod-container/pull-image-private-registry/
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Specifies one or more sets of IvorySQL pods that replicate data for
	// this cluster.
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=2
	InstanceSets []IvoryInstanceSetSpec `json:"instances"`

	// Whether or not the IvorySQL cluster is being deployed to an OpenShift
	// environment. If the field is unset, the operator will automatically
	// detect the environment.
	// +optional
	OpenShift *bool `json:"openshift,omitempty"`

	// +optional
	Patroni *PatroniSpec `json:"patroni,omitempty"`

	// Suspends the rollout and reconciliation of changes made to the
	// IvoryCluster spec.
	// +optional
	Paused *bool `json:"paused,omitempty"`

	// The port on which IvorySQL should listen.
	// +optional
	// +kubebuilder:default=5432
	// +kubebuilder:validation:Minimum=1024
	Port *int32 `json:"port,omitempty"`

	// The major version of PostgreSQL installed in the PostgreSQL image
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=15
	// +operator-sdk:csv:customresourcedefinitions:type=spec,order=1
	PostgresVersion int `json:"postgresVersion"`

	// The PostGIS extension version installed in the IvorySQL image.
	// When image is not set, indicates a PostGIS enabled image will be used.
	// +optional
	PostGISVersion string `json:"postGISVersion,omitempty"`

	// The specification of a proxy that connects to IvorySQL.
	// +optional
	Proxy *IvoryProxySpec `json:"proxy,omitempty"`

	// The specification of a user interface that connects to IvorySQL.
	// +optional
	UserInterface *UserInterfaceSpec `json:"userInterface,omitempty"`

	// The specification of monitoring tools that connect to IvorySQL
	// +optional
	Monitoring *MonitoringSpec `json:"monitoring,omitempty"`

	// Specification of the service that exposes the IvorySQL primary instance.
	// +optional
	Service *ServiceSpec `json:"service,omitempty"`

	// Whether or not the IvorySQL cluster should be stopped.
	// When this is true, workloads are scaled to zero and CronJobs
	// are suspended.
	// Other resources, such as Services and Volumes, remain in place.
	// +optional
	Shutdown *bool `json:"shutdown,omitempty"`

	// Run this cluster as a read-only copy of an existing cluster or archive.
	// +optional
	Standby *IvoryStandbySpec `json:"standby,omitempty"`

	// A list of group IDs applied to the process of a container. These can be
	// useful when accessing shared file systems with constrained permissions.
	// More info: https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#security-context
	// +optional
	SupplementalGroups []int64 `json:"supplementalGroups,omitempty"`

	// Users to create inside IvorySQL and the databases they should access.
	// The default creates one user that can access one database matching the
	// IvoryCluster name. An empty list creates no users. Removing a user
	// from this list does NOT drop the user nor revoke their access.
	// +listType=map
	// +listMapKey=name
	// +optional
	Users []IvoryUserSpec `json:"users,omitempty"`

	Config IvoryAdditionalConfig `json:"config,omitempty"`
}

// DataSource defines data sources for a new IvoryCluster.
type DataSource struct {
	// Defines a pgBackRest cloud-based data source that can be used to pre-populate the
	// the IvorySQL data directory for a new IvorySQL cluster using a pgBackRest restore.
	// The PGBackRest field is incompatible with the IvoryCluster field: only one
	// data source can be used for pre-populating a new IvorySQL cluster
	// +optional
	PGBackRest *PGBackRestDataSource `json:"pgbackrest,omitempty"`

	// Defines a pgBackRest data source that can be used to pre-populate the IvorySQL data
	// directory for a new IvorySQL cluster using a pgBackRest restore.
	// The PGBackRest field is incompatible with the IvoryCluster field: only one
	// data source can be used for pre-populating a new IvorySQL cluster
	// +optional
	IvoryCluster *IvoryClusterDataSource `json:"ivoryCluster,omitempty"`

	// Defines any existing volumes to reuse for this IvoryCluster.
	// +optional
	Volumes *DataSourceVolumes `json:"volumes,omitempty"`
}

// DataSourceVolumes defines any existing volumes to reuse for this IvoryCluster.
type DataSourceVolumes struct {
	// Defines the existing pgData volume and directory to use in the current
	// IvoryCluster.
	// +optional
	PGDataVolume *DataSourceVolume `json:"pgDataVolume,omitempty"`

	// Defines the existing pg_wal volume and directory to use in the current
	// IvoryCluster. Note that a defined pg_wal volume MUST be accompanied by
	// a pgData volume.
	// +optional
	PGWALVolume *DataSourceVolume `json:"pgWALVolume,omitempty"`

	// Defines the existing pgBackRest repo volume and directory to use in the
	// current IvoryCluster.
	// +optional
	PGBackRestVolume *DataSourceVolume `json:"pgBackRestVolume,omitempty"`
}

// DataSourceVolume defines the PVC name and data diretory path for an existing cluster volume.
type DataSourceVolume struct {
	// The existing PVC name.
	PVCName string `json:"pvcName"`

	// The existing directory. When not set, a move Job is not created for the
	// associated volume.
	// +optional
	Directory string `json:"directory,omitempty"`
}

// DatabaseInitSQL defines a ConfigMap containing custom SQL that will
// be run after the cluster is initialized. This ConfigMap must be in the same
// namespace as the cluster.
type DatabaseInitSQL struct {
	// Name is the name of a ConfigMap
	// +required
	Name string `json:"name"`

	// Key is the ConfigMap data key that points to a SQL string
	// +required
	Key string `json:"key"`
}

// IvoryClusterDataSource defines a data source for bootstrapping IvorySQL clusters using a
// an existing IvoryCluster.
type IvoryClusterDataSource struct {

	// The name of an existing IvoryCluster to use as the data source for the new IvoryCluster.
	// Defaults to the name of the IvoryCluster being created if not provided.
	// +optional
	ClusterName string `json:"clusterName,omitempty"`

	// The namespace of the cluster specified as the data source using the clusterName field.
	// Defaults to the namespace of the IvoryCluster being created if not provided.
	// +optional
	ClusterNamespace string `json:"clusterNamespace,omitempty"`

	// The name of the pgBackRest repo within the source IvoryCluster that contains the backups
	// that should be utilized to perform a pgBackRest restore when initializing the data source
	// for the new IvoryCluster.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^repo[1-4]
	RepoName string `json:"repoName"`

	// Command line options to include when running the pgBackRest restore command.
	// https://pgbackrest.org/command.html#command-restore
	// +optional
	Options []string `json:"options,omitempty"`

	// Resource requirements for the pgBackRest restore Job.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Scheduling constraints of the pgBackRest restore Job.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Priority class name for the pgBackRest restore Job pod. Changing this
	// value causes IvorySQL to restart.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`

	// Tolerations of the pgBackRest restore Job.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// Default defines several key default values for a Ivory cluster.
func (s *IvoryClusterSpec) Default() {
	for i := range s.InstanceSets {
		s.InstanceSets[i].Default(i)
	}

	if s.Patroni == nil {
		s.Patroni = new(PatroniSpec)
	}
	s.Patroni.Default()

	if s.Port == nil {
		s.Port = new(int32)
		*s.Port = 5432
	}

	if s.Proxy != nil {
		s.Proxy.Default()
	}

	if s.UserInterface != nil {
		s.UserInterface.Default()
	}
}

// Backups defines a IvorySQL archive configuration
type Backups struct {

	// pgBackRest archive configuration
	// +kubebuilder:validation:Required
	PGBackRest PGBackRestArchive `json:"pgbackrest"`
}

// IvoryClusterStatus defines the observed state of IvoryCluster
type IvoryClusterStatus struct {

	// Identifies the databases that have been installed into IvorySQL.
	DatabaseRevision string `json:"databaseRevision,omitempty"`

	// Current state of IvorySQL instances.
	// +listType=map
	// +listMapKey=name
	// +optional
	InstanceSets []IvoryInstanceSetStatus `json:"instances,omitempty"`

	// +optional
	Patroni PatroniStatus `json:"patroni,omitempty"`

	// Status information for pgBackRest
	// +optional
	PGBackRest *PGBackRestStatus `json:"pgbackrest,omitempty"`

	// Stores the current IvorySQL major version following a successful
	// major IvorySQL upgrade.
	// +optional
	PostgresVersion int `json:"postgresVersion"`

	// Current state of the IvorySQL proxy.
	// +optional
	Proxy IvoryProxyStatus `json:"proxy,omitempty"`

	// The instance that should be started first when bootstrapping and/or starting a
	// IvoryCluster.
	// +optional
	StartupInstance string `json:"startupInstance,omitempty"`

	// The instance set associated with the startupInstance
	// +optional
	StartupInstanceSet string `json:"startupInstanceSet,omitempty"`

	// Current state of the IvorySQL user interface.
	// +optional
	UserInterface *IvoryUserInterfaceStatus `json:"userInterface,omitempty"`

	// Identifies the users that have been installed into IvorySQL.
	UsersRevision string `json:"usersRevision,omitempty"`

	// Current state of IvorySQL cluster monitoring tool configuration
	// +optional
	Monitoring MonitoringStatus `json:"monitoring,omitempty"`

	// DatabaseInitSQL state of custom database initialization in the cluster
	// +optional
	DatabaseInitSQL *string `json:"databaseInitSQL,omitempty"`

	// observedGeneration represents the .metadata.generation on which the status was based.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// conditions represent the observations of ivorycluster's current state.
	// Known .status.conditions.type are: "PersistentVolumeResizing",
	// "Progressing", "ProxyAvailable"
	// +optional
	// +listType=map
	// +listMapKey=type
	// +operator-sdk:csv:customresourcedefinitions:type=status,xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// IvoryClusterStatus condition types.
const (
	PersistentVolumeResizing = "PersistentVolumeResizing"
	IvoryClusterProgressing  = "Progressing"
	ProxyAvailable           = "ProxyAvailable"
)

type IvoryInstanceSetSpec struct {
	// +optional
	Metadata *Metadata `json:"metadata,omitempty"`

	// This value goes into the name of an appsv1.StatefulSet, the hostname of
	// a corev1.Pod, and label values. The pattern below is IsDNS1123Label
	// wrapped in "()?" to accommodate the empty default.
	//
	// The Pods created by a StatefulSet have a "controller-revision-hash" label
	// comprised of the StatefulSet name, a dash, and a 10-character hash.
	// The length below is derived from limitations on label values:
	//
	//   63 (max) ≥ len(cluster) + 1 (dash)
	//                + len(set) + 1 (dash) + 4 (id)
	//                + 1 (dash) + 10 (hash)
	//
	// See: https://issue.k8s.io/64023

	// Name that associates this set of IvorySQL pods. This field is optional
	// when only one instance set is defined. Each instance set in a cluster
	// must have a unique name. The combined length of this and the cluster name
	// must be 46 characters or less.
	// +optional
	// +kubebuilder:default=""
	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?)?$`
	Name string `json:"name"`

	// Scheduling constraints of a IvorySQL pod. Changing this value causes
	// IvorySQL to restart.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Custom sidecars for IvorySQL instance pods. Changing this value causes
	// IvorySQL to restart.
	// +optional
	Containers []corev1.Container `json:"containers,omitempty"`

	// Defines a PersistentVolumeClaim for IvorySQL data.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes
	// +kubebuilder:validation:Required
	DataVolumeClaimSpec corev1.PersistentVolumeClaimSpec `json:"dataVolumeClaimSpec"`

	// Priority class name for the IvorySQL pod. Changing this value causes
	// IvorySQL to restart.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`

	// Number of desired IvorySQL pods.
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Minimum number of pods that should be available at a time.
	// Defaults to one when the replicas field is greater than one.
	// +optional
	MinAvailable *intstr.IntOrString `json:"minAvailable,omitempty"`

	// Compute resources of a IvorySQL container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Configuration for instance sidecar containers
	// +optional
	Sidecars *InstanceSidecars `json:"sidecars,omitempty"`

	// Tolerations of a IvorySQL pod. Changing this value causes IvorySQL to restart.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Topology spread constraints of a IvorySQL pod. Changing this value causes
	// IvorySQL to restart.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/
	// +optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	// Defines a separate PersistentVolumeClaim for IvorySQL's write-ahead log.
	// More info: https://www.ivorysql.org/docs/current/wal.html
	// +optional
	WALVolumeClaimSpec *corev1.PersistentVolumeClaimSpec `json:"walVolumeClaimSpec,omitempty"`

	// The list of tablespaces volumes to mount for this ivorycluster
	// This field requires enabling TablespaceVolumes feature gate
	// +listType=map
	// +listMapKey=name
	// +optional
	TablespaceVolumes []TablespaceVolume `json:"tablespaceVolumes,omitempty"`
}

type TablespaceVolume struct {
	// This value goes into
	// a. the name of a corev1.PersistentVolumeClaim,
	// b. a label value, and
	// c. a path name.
	// So it must match both IsDNS1123Subdomain and IsValidLabelValue;
	// and be valid as a file path.

	// The name for the tablespace, used as the path name for the volume.
	// Must be unique in the instance set since they become the directory names.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^[a-z][a-z0-9]*$`
	// +kubebuilder:validation:Type=string
	Name string `json:"name"`

	// Defines a PersistentVolumeClaim for a tablespace.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes
	// +kubebuilder:validation:Required
	DataVolumeClaimSpec corev1.PersistentVolumeClaimSpec `json:"dataVolumeClaimSpec"`
}

// InstanceSidecars defines the configuration for instance sidecar containers
type InstanceSidecars struct {
	// Defines the configuration for the replica cert copy sidecar container
	// +optional
	ReplicaCertCopy *Sidecar `json:"replicaCertCopy,omitempty"`
}

// Default sets the default values for an instance set spec, including the name
// suffix and number of replicas.
func (s *IvoryInstanceSetSpec) Default(i int) {
	if s.Name == "" {
		s.Name = fmt.Sprintf("%02d", i)
	}
	if s.Replicas == nil {
		s.Replicas = new(int32)
		*s.Replicas = 1
	}
}

type IvoryInstanceSetStatus struct {
	Name string `json:"name"`

	// Total number of ready pods.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Total number of pods.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Total number of pods that have the desired specification.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
}

// IvoryProxySpec is a union of the supported IvorySQL proxies.
type IvoryProxySpec struct {

	// Defines a PgBouncer proxy and connection pooler.
	PGBouncer *PGBouncerPodSpec `json:"pgBouncer"`
}

// Default sets the defaults for any proxies that are set.
func (s *IvoryProxySpec) Default() {
	if s.PGBouncer != nil {
		s.PGBouncer.Default()
	}
}

type IvoryProxyStatus struct {
	PGBouncer PGBouncerPodStatus `json:"pgBouncer,omitempty"`
}

// IvoryStandbySpec defines if/how the cluster should be a hot standby.
type IvoryStandbySpec struct {
	// Whether or not the IvorySQL cluster should be read-only. When this is
	// true, WAL files are applied from a pgBackRest repository or another
	// IvorySQL server.
	// +optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// The name of the pgBackRest repository to follow for WAL files.
	// +optional
	// +kubebuilder:validation:Pattern=^repo[1-4]
	RepoName string `json:"repoName,omitempty"`

	// Network address of the IvorySQL server to follow via streaming replication.
	// +optional
	Host string `json:"host,omitempty"`

	// Network port of the IvorySQL server to follow via streaming replication.
	// +optional
	// +kubebuilder:validation:Minimum=1024
	Port *int32 `json:"port,omitempty"`
}

// UserInterfaceSpec is a union of the supported IvorySQL user interfaces.
type UserInterfaceSpec struct {

	// Defines a pgAdmin user interface.
	PGAdmin *PGAdminPodSpec `json:"pgAdmin"`
}

// Default sets the defaults for any user interfaces that are set.
func (s *UserInterfaceSpec) Default() {
	if s.PGAdmin != nil {
		s.PGAdmin.Default()
	}
}

// IvoryUserInterfaceStatus is a union of the supported IvorySQL user
// interface statuses.
type IvoryUserInterfaceStatus struct {

	// The state of the pgAdmin user interface.
	PGAdmin PGAdminPodStatus `json:"pgAdmin,omitempty"`
}

type IvoryAdditionalConfig struct {
	Files []corev1.VolumeProjection `json:"files,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +operator-sdk:csv:customresourcedefinitions:resources={{ConfigMap,v1},{Secret,v1},{Service,v1},{CronJob,v1beta1},{Deployment,v1},{Job,v1},{StatefulSet,v1},{PersistentVolumeClaim,v1}}

// IvoryCluster is the Schema for the ivoryclusters API
type IvoryCluster struct {
	// ObjectMeta.Name is a DNS subdomain.
	// - https://docs.k8s.io/concepts/overview/working-with-objects/names/#dns-subdomain-names
	// - https://releases.k8s.io/v1.21.0/staging/src/k8s.io/apiextensions-apiserver/pkg/registry/customresource/validator.go#L60

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// NOTE(cbandy): Every IvoryCluster needs a Spec, but it is optional here
	// so ObjectMeta can be managed independently.

	Spec   IvoryClusterSpec   `json:"spec,omitempty"`
	Status IvoryClusterStatus `json:"status,omitempty"`
}

// Default implements "sigs.k8s.io/controller-runtime/pkg/webhook.Defaulter" so
// a webhook can be registered for the type.
// - https://book.kubebuilder.io/reference/webhook-overview.html
func (c *IvoryCluster) Default() {
	if len(c.APIVersion) == 0 {
		c.APIVersion = GroupVersion.String()
	}
	if len(c.Kind) == 0 {
		c.Kind = "IvoryCluster"
	}
	c.Spec.Default()
}

// +kubebuilder:object:root=true

// IvoryClusterList contains a list of IvoryCluster
type IvoryClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IvoryCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IvoryCluster{}, &IvoryClusterList{})
}

// MonitoringSpec is a union of the supported IvorySQL Monitoring tools
type MonitoringSpec struct {
	// +optional
	PGMonitor *PGMonitorSpec `json:"pgmonitor,omitempty"`
}

// MonitoringStatus is the current state of IvorySQL cluster monitoring tool
// configuration
type MonitoringStatus struct {
	// +optional
	ExporterConfiguration string `json:"exporterConfiguration,omitempty"`
}

// PGMonitorSpec defines the desired state of the pgMonitor tool suite
type PGMonitorSpec struct {
	// +optional
	Exporter *ExporterSpec `json:"exporter,omitempty"`
}

type ExporterSpec struct {

	// Projected volumes containing custom IvorySQL Exporter configuration.  Currently supports
	// the customization of IvorySQL Exporter queries. If a "queries.yml" file is detected in
	// any volume projected using this field, it will be loaded using the "extend.query-path" flag:
	// https://github.com/prometheus-community/postgres_exporter#flags
	// Changing the values of field causes IvorySQL and the exporter to restart.
	// +optional
	Configuration []corev1.VolumeProjection `json:"configuration,omitempty"`

	// Projected secret containing custom TLS certificates to encrypt output from the exporter
	// web server
	// +optional
	CustomTLSSecret *corev1.SecretProjection `json:"customTLSSecret,omitempty"`

	// The image name to use for highgo-ivory-exporter containers. The image may
	// also be set using the RELATED_IMAGE_PGEXPORTER environment variable.
	// +optional
	Image string `json:"image,omitempty"`

	// Changing this value causes IvorySQL and the exporter to restart.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

func NewIvoryCluster() *IvoryCluster {
	cluster := &IvoryCluster{}
	cluster.SetGroupVersionKind(GroupVersion.WithKind("IvoryCluster"))
	return cluster
}
