package v1alpha1

import corev1 "k8s.io/api/core/v1"

type ClusterConfigSpec struct {

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	HostPID bool `json:"hostPID,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// +kubebuilder:validation:Optional
	DnsPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	ShareProcessNamespace bool `json:"shareProcessNamespace,omitempty"`

	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	// +kubebuilder:validation:Optional
	TieredStore []*TieredStore `json:"tieredStore,omitempty"`

	// +kubebuilder:validation:Optional
	ShortCircuit *ShortCircuitSpec `json:"shortCircuit,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"alluxio.security.stale.channel.purge.interval": "365d"}
	Properties map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"-XX:+UseContainerSupport"}
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// +kubebuilder:validation:Optional
	Journal *JournalSpec `json:"journal,omitempty"`
	// +kubebuilder:validation:Optional
	Listener *ListenerSpec `json:"listener,omitempty"`
}

type ListenerSpec struct {
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// +kubebuilder:validation:enum=ClusterIP;NodePort;LoadBalancer;ExternalName
	// +kubebuilder:default=ClusterIP
	Type corev1.ServiceType `json:"type,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=9083
	Port int32 `json:"port,omitempty"`
}

type ConfigOverridesSpec struct {
	OverrideConfig map[string]string `json:"overrideConfig,omitempty"`
}

type MasterConfigSpec struct {
	// +kubebuilder:validation:Optional
	Resources *ResourcesSpec `json:"resources,omitempty"`
	// +kubebuilder:validation:Optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
	// +kubebuilder:validation:Optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext"`

	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`

	// +kubebuilder:validation:Optional
	Logging *ContainerLoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	EnvVars map[string]string `json:"envVars,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraContainers []corev1.Container `json:"extraContainers,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumes []corev1.Volume `json:"extraVolumes,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	HostPID *bool `json:"hostPID,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	HostNetwork *bool `json:"hostNetwork,omitempty"`

	// +kubebuilder:validation:Optional
	DnsPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Properties map[string]string `json:"properties,omitempty"`
	// +kubebuilder:validation:Optional
	Ports *MasterPortsSpec `json:"ports,omitempty"`
	// +kubebuilder:validation:Optional
	JobMaster *JobMasterSpec `json:"jobMaster,omitempty"`
}

type WorkerConfigSpec struct {
	// +kubebuilder:validation:Optional
	Resources *ResourcesSpec `json:"resources,omitempty"`
	// +kubebuilder:validation:Optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
	// +kubebuilder:validation:Optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext"`

	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`

	// +kubebuilder:validation:Optional
	Logging *ContainerLoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	EnvVars map[string]string `json:"envVars,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraContainers []corev1.Container `json:"extraContainers,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumes []corev1.Volume `json:"extraVolumes,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	JvmOptions []string `json:"jvmOptions,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	HostPID *bool `json:"hostPID,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	HostNetwork *bool `json:"hostNetwork,omitempty"`

	// +kubebuilder:validation:Optional
	DnsPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Properties map[string]string `json:"properties,omitempty"`

	Resource ResourcesSpec `json:"resource,omitempty"`

	// +kubebuilder:validation:Optional
	Ports *WorkerPortsSpec `json:"ports,omitempty"`
	// +kubebuilder:validation:Optional
	JobWorker *JobWorkerSpec `json:"jobWorker,omitempty"`
}

type BaseGroupSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas int32 `json:"replicas,omitempty"`
	// +kubebuilder:validation:Optional
	CommandArgsOverrides []string `json:"commandArgsOverrides,omitempty"`
	// +kubebuilder:validation:Optional
	ConfigOverrides *ConfigOverridesSpec `json:"configOverrides,omitempty"`
	// +kubebuilder:validation:Optional
	EnvOverrides map[string]string `json:"envOverrides,omitempty"`
	//// +kubebuilder:validation:Optional
	//PodOverride corev1.PodSpec `json:"podOverride,omitempty"`
}

type MasterRoleGroupSpec struct {
	BaseGroupSpec `json:",inline"`
	Config        *MasterConfigSpec `json:"config,omitempty"`
}

type WorkerRoleGroupSpec struct {
	BaseGroupSpec `json:",inline"`
	Config        *WorkerConfigSpec `json:"config,omitempty"`
}

type DatabaseSpec struct {
	// +kubebuilder:validation=Optional
	Reference string `json:"reference"`

	// +kubebuilder:validation=Optional
	Inline *DatabaseInlineSpec `json:"inline,omitempty"`
}

// DatabaseInlineSpec defines the inline database spec.
type DatabaseInlineSpec struct {
	// +kubebuilder:validation:Enum=mysql;postgres
	// +kubebuilder:default="postgres"
	Driver string `json:"driver,omitempty"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default="hive"
	DatabaseName string `json:"databaseName,omitempty"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default="hive"
	Username string `json:"username,omitempty"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default="hive"
	Password string `json:"password,omitempty"`

	// +kubebuilder:validation=Required
	Host string `json:"host,omitempty"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default=5432
	Port int32 `json:"port,omitempty"`
}

type S3BucketSpec struct {
	// S3 bucket name with S3Bucket
	// +kubebuilder:validation=Optional
	Reference *string `json:"reference"`

	// +kubebuilder:validation=Optional
	Inline *S3BucketInlineSpec `json:"inline,omitempty"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default=20
	MaxConnect int `json:"maxConnect"`

	// +kubebuilder:validation=Optional
	PathStyleAccess bool `json:"pathStyle_access"`
}

type S3BucketInlineSpec struct {

	// +kubeBuilder:validation=Required
	Bucket string `json:"bucket"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default="us-east-1"
	Region string `json:"region,omitempty"`

	// +kubebuilder:validation=Required
	Endpoints string `json:"endpoints"`

	// +kubebuilder:validation=Optional
	// +kubebuilder:default=false
	SSL bool `json:"ssl,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	PathStyle bool `json:"pathStyle,omitempty"`

	// +kubebuilder:validation=Optional
	AccessKey string `json:"accessKey,omitempty"`

	// +kubebuilder:validation=Optional
	SecretKey string `json:"secretKey,omitempty"`
}
