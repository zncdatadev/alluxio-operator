/*
Copyright 2023 zncdata-labs.

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

package v1alpha1

import (
	"github.com/zncdata-labs/operator-go/pkg/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	MasterEmbedded    = 19200
	MasterRpcPort     = 19998
	MasterWebPort     = 19999
	JobMasterRpcPort  = 20001
	JobMasterWebPort  = 20002
	JobMasterEmbedded = 20003
	WorkerRpcPort     = 29999
	WorkerWebPort     = 30000
	JobWorkerRpcPort  = 30001
	JobWorkerDataPort = 30002
	JobWorkerWebPort  = 30003
)

// AlluxioSpec defines the desired state of Alluxio
type AlluxioSpec struct {
	// +kubebuilder:validation:Required
	ClusterConfig *ClusterConfigSpec `json:"clusterConfig,omitempty"`

	// +kubebuilder:validation:Required
	Master *MasterSpec `json:"master,omitempty"`

	// +kubebuilder:validation:Required
	Worker *WorkerSpec `json:"worker,omitempty"`
}

func (r *Alluxio) GetNameWithSuffix(suffix string) string {
	// return sparkHistory.GetName() + rand.String(5) + suffix
	return r.GetName() + "-" + suffix
}

type ClusterConfigSpec struct {
	// +kubebuilder:validation:Optional
	Image *ImageSpec `json:"image,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// +kubebuilder:validation:Optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations *corev1.Toleration `json:"tolerations,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	EnvVars map[string]string `json:"envVars,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	MasterPorts *MasterPortsSpec `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	WorkerPorts *WorkerPortsSpec `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraContainers []corev1.Container `json:"extraContainers,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumes []corev1.Volume `json:"extraVolumes,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts,omitempty"`

	// +kubebuilder:validation:Optional
	JobMaster *JobMasterSpec `json:"jobMaster,omitempty"`

	// +kubebuilder:validation:Optional
	JobWorker *JobWorkerSpec `json:"jobWorker,omitempty"`

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
}

func (clusterConfig *ClusterConfigSpec) GetSecurityContext() *corev1.PodSecurityContext {
	if clusterConfig != nil && clusterConfig.SecurityContext != nil {
		return clusterConfig.SecurityContext
	}
	runAsUser := int64(1000)
	runAsGroup := int64(1000)
	fsGroup := int64(1000)
	return &corev1.PodSecurityContext{
		RunAsUser:  &runAsUser,
		RunAsGroup: &runAsGroup,
		FSGroup:    &fsGroup,
	}
}

type MasterSpec struct {
	// +kubebuilder:validation:Optional
	RoleConfig *RoleGroupMasterSpec `json:"roleConfig,omitempty"`

	// +kubebuilder:validation:Optional
	RoleGroups map[string]*RoleGroupMasterSpec `json:"roleGroups,omitempty"`
}

//func (master *MasterSpec) GetRoleGroup(name string) *RoleGroupMasterSpec {
//	if master.RoleGroups == nil {
//		return nil
//	}
//
//	roleGroup := master.RoleGroups[name]
//	roleConfig := master.RoleConfig
//
//	var instance Alluxio
//	var image ImageSpec
//	var podSecurityContext *corev1.PodSecurityContext
//	var args []string
//	var jobArgs []string
//	var jobResources corev1.ResourceRequirements
//	var hostPID bool
//	var hostNetwork bool
//	var dnsPolicy corev1.DNSPolicy
//	var shareProcessNamespace bool
//
//	if roleGroup != nil && roleGroup.HostPID {
//		hostPID = roleGroup.HostPID
//	} else if roleConfig.HostPID {
//		hostPID = roleConfig.HostPID
//	}
//
//	if roleGroup != nil && roleGroup.HostNetwork {
//		hostNetwork = roleGroup.HostNetwork
//	} else if roleConfig.HostNetwork {
//		hostNetwork = roleConfig.HostNetwork
//	}
//
//	if roleGroup != nil && roleGroup.DnsPolicy != "" {
//		dnsPolicy = roleGroup.DnsPolicy
//	} else if roleConfig.DnsPolicy != "" {
//		dnsPolicy = roleConfig.DnsPolicy
//	} else {
//		if hostNetwork {
//			dnsPolicy = corev1.DNSClusterFirstWithHostNet
//		} else {
//			dnsPolicy = corev1.DNSClusterFirst
//		}
//	}
//
//	if roleGroup != nil && roleGroup.ShareProcessNamespace {
//		shareProcessNamespace = roleGroup.ShareProcessNamespace
//	} else if roleConfig.ShareProcessNamespace {
//		shareProcessNamespace = roleConfig.ShareProcessNamespace
//	}
//
//	if roleGroup != nil && roleGroup.Image != nil {
//		image = *roleGroup.Image
//	} else {
//		image = *instance.Spec.ClusterConfig.Image
//	}
//
//	if roleGroup != nil && roleGroup.SecurityContext != nil {
//		securityContext := roleGroup.SecurityContext
//		podSecurityContext = &corev1.PodSecurityContext{
//			RunAsUser:  securityContext.RunAsUser,
//			RunAsGroup: securityContext.RunAsGroup,
//			FSGroup:    securityContext.FSGroup,
//		}
//	} else if instance.Spec.ClusterConfig.SecurityContext != nil {
//		securityContext := instance.Spec.ClusterConfig.SecurityContext
//		podSecurityContext = &corev1.PodSecurityContext{
//			RunAsUser:  securityContext.RunAsUser,
//			RunAsGroup: securityContext.RunAsGroup,
//			FSGroup:    securityContext.FSGroup,
//		}
//	}
//
//	if roleGroup != nil && roleGroup.Args != nil {
//		args = roleGroup.Args
//	} else {
//		args = instance.Spec.Master.Args
//	}
//
//	if roleGroup != nil && roleGroup.JobMaster != nil && roleGroup.JobMaster.Args != nil {
//		jobArgs = roleGroup.JobMaster.Args
//	} else if roleConfig != nil && roleConfig.JobMaster != nil && roleConfig.JobMaster.Args != nil {
//		jobArgs = roleConfig.JobMaster.Args
//	}
//
//	if roleGroup != nil && roleGroup.JobMaster != nil && roleGroup.JobMaster.Resources != nil {
//		jobResources = *roleGroup.JobMaster.Resources
//	} else if roleConfig.JobMaster.Resources != nil {
//		jobResources = *roleConfig.JobMaster.Resources
//	}
//
//	var masterEmbedded int32
//	var masterRpcPort int32
//	var masterWebPort int32
//	if roleGroup.Ports != nil {
//		masterEmbedded = roleGroup.Ports.Embedded
//		masterRpcPort = roleGroup.Ports.Rpc
//		masterWebPort = roleGroup.Ports.Web
//	} else if master.GetMasterPorts() != nil {
//		masterEmbedded = master.GetMasterPorts().Embedded
//		masterRpcPort = master.GetMasterPorts().Rpc
//		masterWebPort = master.GetMasterPorts().Web
//	}
//
//	var jobMasterEmbedded int32
//	var jobMasterRpcPort int32
//	var jobMasterWebPort int32
//	if roleGroup.JobMaster.Ports != nil {
//		jobMasterEmbedded = roleGroup.JobMaster.Ports.Embedded
//		jobMasterRpcPort = roleGroup.JobMaster.Ports.Rpc
//		jobMasterWebPort = roleGroup.JobMaster.Ports.Web
//	} else if master.RoleConfig.JobMaster.Ports != nil {
//		jobMasterEmbedded = master.GetJobMasterPorts(clusterconfig, ).Embedded
//		jobMasterRpcPort = master.GetJobMasterPorts().Rpc
//		jobMasterWebPort = master.GetJobMasterPorts().Web
//	}
//
//	mergedRoleGroup := &RoleGroupMasterSpec{
//		HostPID:               hostPID,
//		HostNetwork:           hostNetwork,
//		DnsPolicy:             dnsPolicy,
//		ShareProcessNamespace: shareProcessNamespace,
//		Image:                 &image,
//		Ports: &MasterPortsSpec{
//			Embedded: masterEmbedded,
//			Rpc:      masterRpcPort,
//			Web:      masterWebPort,
//		},
//		SecurityContext: podSecurityContext,
//		Args:            args,
//		JobMaster: &JobMasterSpec{
//			Args:      jobArgs,
//			Resources: &jobResources,
//			Ports: &JobMasterPortsSpec{
//				Embedded: jobMasterEmbedded,
//				Rpc:      jobMasterRpcPort,
//				Web:      jobMasterWebPort,
//			},
//		},
//	}
//
//	return mergedRoleGroup
//}

type RoleConfigMasterSpec struct {
	// +kubebuilder:validation:Optional
	JobMaster *JobMasterSpec `json:"jobMaster,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	Count int32 `json:"count,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	JvmOptions []string `json:"jvmOptions,omitempty"`

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
	// +kubebuilder:default={}
	Properties map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	EnvVars map[string]string `json:"envVars,omitempty"`
}

func (master *MasterSpec) GetHostPID() bool {
	if master.RoleConfig.HostPID {
		return master.RoleConfig.HostPID
	}
	return false
}

func (master *MasterSpec) GetHostNetwork() bool {
	if master.RoleConfig.HostNetwork {
		return master.RoleConfig.HostNetwork
	}
	return false
}

func (master *MasterSpec) GetDnsPolicy() corev1.DNSPolicy {
	if master.GetHostNetwork() {
		return corev1.DNSClusterFirstWithHostNet
	}
	return corev1.DNSClusterFirst
}

func (master *MasterSpec) GetShareProcessNamespace() bool {
	if master.RoleConfig.ShareProcessNamespace {
		return master.RoleConfig.ShareProcessNamespace
	}
	return false

}

type RoleGroupMasterSpec struct {
	// +kubebuilder:validation:Optional
	Image *ImageSpec `json:"image,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// +kubebuilder:validation:Optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations *corev1.Toleration `json:"tolerations,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	EnvVars map[string]string `json:"envVars,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	Ports *MasterPortsSpec `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraContainers []corev1.Container `json:"extraContainers,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumes []corev1.Volume `json:"extraVolumes,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts,omitempty"`

	// +kubebuilder:validation:Optional
	JobMaster *JobMasterSpec `json:"jobMaster,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	JvmOptions []string `json:"jvmOptions,omitempty"`

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
	// +kubebuilder:default={}
	Properties map[string]string `json:"properties,omitempty"`
}

type MasterPortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=19200
	Embedded int32 `json:"embedded,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=19998
	Rpc int32 `json:"rpc,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=19999
	Web int32 `json:"web,omitempty"`
}

//func (master *MasterSpec) GetMasterPorts() *MasterPortsSpec {
//	if master.Ports == nil {
//		return &MasterPortsSpec{
//			Embedded: int32(MasterEmbedded),
//			Rpc:      int32(MasterRpcPort),
//			Web:      int32(MasterWebPort),
//		}
//	}
//	return master.Ports
//}

type WorkerSpec struct {
	// +kubebuilder:validation:Optional
	RoleConfig *RoleGroupWorkerSpec `json:"roleConfig,omitempty"`

	// +kubebuilder:validation:Optional
	RoleGroups map[string]*RoleGroupWorkerSpec `json:"roleGroups,omitempty"`
}

func (worker *WorkerSpec) GetRoleGroup(instance AlluxioSpec, name string) *RoleGroupWorkerSpec {
	if worker.RoleGroups == nil {
		return nil
	}

	clusterConfig := instance.ClusterConfig
	roleGroup := worker.RoleGroups[name]
	roleConfig := worker.RoleConfig

	var image ImageSpec
	var podSecurityContext *corev1.PodSecurityContext
	var args []string
	var jobArgs []string
	var jobResources corev1.ResourceRequirements
	var envVars map[string]string
	var extraContainers []corev1.Container
	var resources corev1.ResourceRequirements
	var replica int32

	hostPID := worker.GetHostPID(roleGroup, roleConfig)
	hostNetwork := worker.GetHostNetwork(roleGroup, roleConfig)
	dnsPolicy := worker.GetDnsPolicy(roleGroup, roleConfig)
	shareProcessNamespace := worker.GetShareProcessNamespace(roleGroup, roleConfig)

	if roleGroup != nil {
		if roleGroup.Image != nil {
			image = *roleGroup.Image
		} else if roleConfig.Image != nil {
			image = *roleConfig.Image
		} else if clusterConfig.Image != nil {
			image = *clusterConfig.Image
		}

		if roleGroup.SecurityContext != nil {
			podSecurityContext = roleGroup.SecurityContext
		} else {
			podSecurityContext = &corev1.PodSecurityContext{
				RunAsUser:  instance.ClusterConfig.GetSecurityContext().RunAsUser,
				RunAsGroup: instance.ClusterConfig.GetSecurityContext().RunAsGroup,
				FSGroup:    instance.ClusterConfig.GetSecurityContext().FSGroup,
			}
		}

		if roleGroup.Replicas != nil {
			replica = *roleGroup.Replicas
		} else if roleConfig.Replicas != nil {
			replica = *roleConfig.Replicas
		} else if clusterConfig.Replicas != nil {
			replica = *clusterConfig.Replicas
		} else {
			replica = 1
		}

		if roleGroup.Resources != nil {
			resources = *roleGroup.Resources
		} else if roleConfig.Resources != nil {
			resources = *roleConfig.Resources
		}

		if roleGroup.Args != nil {
			args = roleGroup.Args
		} else if roleConfig.Args != nil {
			args = roleConfig.Args
		}

		if roleGroup.EnvVars != nil {
			envVars = roleGroup.EnvVars
		} else {
			envVars = instance.Worker.RoleConfig.EnvVars
		}

		if roleGroup.JobWorker != nil {
			if roleGroup.JobWorker.Args != nil {
				jobArgs = roleGroup.JobWorker.Args
			}

			if roleGroup.JobWorker.Resources != nil {
				jobResources = *roleGroup.JobWorker.Resources
			}
		}
	}

	if roleGroup != nil && roleGroup.ExtraContainers != nil {
		extraContainers = roleGroup.ExtraContainers
	} else if instance.Worker.RoleConfig.ExtraContainers != nil {
		extraContainers = instance.Worker.RoleConfig.ExtraContainers
	}

	workerPorts := worker.GetWorkerPorts(clusterConfig, roleGroup, roleConfig)
	workerRpcPort := workerPorts.Rpc
	workerWebPort := workerPorts.Web

	jobWorkerPorts := worker.GetJobWorkerPorts(clusterConfig, roleGroup, roleConfig)
	jobWorkerRpcPort := jobWorkerPorts.Rpc
	jobWorkerDataPort := jobWorkerPorts.Data
	jobWorkerWebPort := jobWorkerPorts.Web

	mergedRoleGroup := &RoleGroupWorkerSpec{
		HostPID:               &hostPID,
		HostNetwork:           &hostNetwork,
		DnsPolicy:             dnsPolicy,
		ShareProcessNamespace: &shareProcessNamespace,
		Image:                 &image,
		Replicas:              &replica,
		Ports: &WorkerPortsSpec{
			Rpc: workerRpcPort,
			Web: workerWebPort,
		},
		Resources:       &resources,
		SecurityContext: podSecurityContext,
		Args:            args,
		EnvVars:         envVars,
		ExtraContainers: extraContainers,
		JobWorker: &JobWorkerSpec{
			Args:      jobArgs,
			Resources: &jobResources,
			Ports: &JobWorkerPortsSpec{
				Data: jobWorkerDataPort,
				Rpc:  jobWorkerRpcPort,
				Web:  jobWorkerWebPort,
			},
		},
	}

	return mergedRoleGroup
}

func (worker *WorkerSpec) GetHostPID(RoleGroup *RoleGroupWorkerSpec, RoleConfig *RoleGroupWorkerSpec) bool {
	if RoleGroup != nil && RoleGroup.HostPID != nil {
		return *RoleGroup.HostPID
	} else {
		return *RoleConfig.HostPID
	}
}

func (worker *WorkerSpec) GetHostNetwork(RoleGroup *RoleGroupWorkerSpec, RoleConfig *RoleGroupWorkerSpec) bool {
	if RoleGroup != nil && RoleGroup.HostNetwork != nil {
		return *RoleGroup.HostNetwork
	} else {
		return *RoleConfig.HostNetwork
	}
}

func (worker *WorkerSpec) GetDnsPolicy(RoleGroup *RoleGroupWorkerSpec, RoleConfig *RoleGroupWorkerSpec) corev1.DNSPolicy {
	if worker.GetHostNetwork(RoleGroup, RoleConfig) {
		return corev1.DNSClusterFirstWithHostNet
	}
	return corev1.DNSClusterFirst
}

func (worker *WorkerSpec) GetShareProcessNamespace(RoleGroup *RoleGroupWorkerSpec, RoleConfig *RoleGroupWorkerSpec) bool {
	if RoleGroup != nil && RoleGroup.ShareProcessNamespace != nil {
		return *RoleGroup.ShareProcessNamespace
	} else {
		return *RoleConfig.ShareProcessNamespace
	}
}

type RoleGroupWorkerSpec struct {
	// +kubebuilder:validation:Optional
	Image *ImageSpec `json:"image,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// +kubebuilder:validation:Optional
	MatchLabels map[string]string `json:"matchLabels,omitempty"`

	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuilder:validation:Optional
	Tolerations *corev1.Toleration `json:"tolerations,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	EnvVars map[string]string `json:"envVars,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Args []string `json:"args,omitempty"`

	// +kube:validation:Optional
	Ports *WorkerPortsSpec `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraContainers []corev1.Container `json:"extraContainers,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumes []corev1.Volume `json:"extraVolumes,omitempty"`

	// +kubebuilder:validation:Optional
	ExtraVolumeMounts []corev1.VolumeMount `json:"extraVolumeMounts,omitempty"`

	// +kubebuilder:validation:Optional
	JobWorker *JobWorkerSpec `json:"jobWorker,omitempty"`

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
}

type WorkerPortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=29999
	Rpc int32 `json:"rpc,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30000
	Web int32 `json:"web,omitempty"`
}

func (worker *WorkerSpec) GetWorkerPorts(ClusterConfig *ClusterConfigSpec, RoleGroup *RoleGroupWorkerSpec, RoleConfig *RoleGroupWorkerSpec) *WorkerPortsSpec {
	if RoleGroup != nil && RoleGroup.Ports != nil {
		return RoleGroup.Ports
	} else if RoleConfig != nil && RoleConfig.Ports != nil {
		return RoleConfig.Ports
	} else if ClusterConfig != nil && ClusterConfig.WorkerPorts != nil {
		return ClusterConfig.WorkerPorts
	} else {
		return &WorkerPortsSpec{
			Rpc: int32(WorkerRpcPort),
			Web: int32(WorkerWebPort),
		}
	}
}

type JobMasterSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"job-master"}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Properties map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Optional
	Ports *JobMasterPortsSpec `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	JvmOptions []string `json:"jvmOptions,omitempty"`
}

type JobMasterPortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=20001
	Rpc int32 `json:"rpc,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=20002
	Web int32 `json:"web,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=20003
	Embedded int32 `json:"embedded,omitempty"`
}

func (master *MasterSpec) GetJobMasterPorts(ClusterConfig *ClusterConfigSpec, RoleGroup *RoleGroupMasterSpec, RoleConfig *RoleGroupMasterSpec) *JobMasterPortsSpec {
	if RoleGroup != nil && RoleGroup.JobMaster != nil && RoleGroup.JobMaster.Ports != nil {
		return RoleGroup.JobMaster.Ports
	} else if RoleConfig != nil && RoleConfig.JobMaster != nil && RoleConfig.JobMaster.Ports != nil {
		return RoleConfig.JobMaster.Ports
	} else if ClusterConfig != nil && ClusterConfig.JobWorker != nil && ClusterConfig.JobWorker.Ports != nil {
		return ClusterConfig.JobMaster.Ports
	} else {
		return &JobMasterPortsSpec{
			Rpc:      int32(JobMasterRpcPort),
			Web:      int32(JobMasterWebPort),
			Embedded: int32(JobMasterEmbedded),
		}
	}
}

type JobWorkerSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"job-worker"}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	Properties map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:Optional
	Ports *JobWorkerPortsSpec `json:"ports,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	JvmOptions []string `json:"jvmOptions,omitempty"`
}

type JobWorkerPortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30001
	Rpc int32 `json:"rpc,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30002
	Data int32 `json:"data,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30003
	Web int32 `json:"web,omitempty"`
}

func (worker *WorkerSpec) GetJobWorkerPorts(ClusterConfig *ClusterConfigSpec, RoleGroup *RoleGroupWorkerSpec, RoleConfig *RoleGroupWorkerSpec) *JobWorkerPortsSpec {
	if RoleGroup != nil && RoleGroup.JobWorker != nil && RoleGroup.JobWorker.Ports != nil {
		return RoleGroup.JobWorker.Ports
	} else if RoleConfig != nil && RoleConfig.JobWorker != nil && RoleConfig.JobWorker.Ports != nil {
		return RoleConfig.JobWorker.Ports
	} else if ClusterConfig != nil && ClusterConfig.JobWorker != nil && ClusterConfig.JobWorker.Ports != nil {
		return ClusterConfig.JobWorker.Ports
	} else {
		return &JobWorkerPortsSpec{
			Rpc:  int32(JobWorkerRpcPort),
			Data: int32(JobWorkerDataPort),
			Web:  int32(JobWorkerWebPort),
		}
	}
}

type ImageSpec struct {

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=alluxio/alluxio
	Repository string `json:"repository"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=latest
	Tag string `json:"tag"`

	// +kubebuilder:validation:enum=Always;Never;IfNotPresent
	// +kubebuilder:default=IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy"`
}

type TieredStore struct {
	Level      int32   `json:"level"`
	Alias      string  `json:"alias"`
	MediumType string  `json:"mediumType"`
	Path       string  `json:"path"`
	Type       string  `json:"type"`
	Quota      string  `json:"quota"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
}

type ShortCircuitSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="uuid"
	Policy string `json:"policy,omitempty"`

	// +kubebuilder:validation:hostPath,persistentVolumeClaim
	// +kubebuilder:default="hostPath"
	VolumeType string `json:"volumeType,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="100Mi"
	Size string `json:"size,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="alluxio-worker-domain-socket"
	PvcName string `json:"pvcName,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="standard"
	StorageClass string `json:"storageClass,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="ReadWriteOnce"
	AccessMode string `json:"accessMode,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/tmp/"
	HosePath string `json:"path,omitempty"`
}

type JournalSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="UFS"
	Type string `json:"type,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/journal"
	Folder string `json:"folder,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="local"
	UfsType string `json:"ufsType,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="persistentVolumeClaim"
	VolumeType string `json:"volumeType,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="1Gi"
	Size string `json:"size,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="standard"
	StorageClass string `json:"storageClass,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="ReadWriteOnce"
	AccessMode string `json:"accessMode,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	Medium string `json:"medium,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	RunFormat bool `json:"runFormat,omitempty"`
}

func (clusterConfig *ClusterConfigSpec) GetShortCircuit() ShortCircuitSpec {
	return ShortCircuitSpec{
		Enabled:      true,
		Policy:       "uuid",
		VolumeType:   "hostPath",
		Size:         "100Mi",
		PvcName:      "alluxio-worker-domain-socket",
		StorageClass: "standard",
		AccessMode:   "ReadWriteOnce",
		HosePath:     "/tmp/",
	}
}

func (clusterConfig *ClusterConfigSpec) GetJournal() JournalSpec {
	return JournalSpec{
		Type:         "UFS",
		Folder:       "/journal",
		UfsType:      "local",
		VolumeType:   "persistentVolumeClaim",
		Size:         "1Gi",
		StorageClass: "standard",
		AccessMode:   "ReadWriteOnce",
		Medium:       "",
		RunFormat:    false,
	}
}

// AlluxioStatus defines the observed state of Alluxio
type AlluxioStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"condition,omitempty"`
	// +kubebuilder:validation:Optional
	URLs []StatusURL `json:"urls,omitempty"`
}

type StatusURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// SetStatusCondition updates the status condition using the provided arguments.
// If the condition already exists, it updates the condition; otherwise, it appends the condition.
// If the condition status has changed, it updates the condition's LastTransitionTime.
func (r *Alluxio) SetStatusCondition(condition metav1.Condition) {
	r.Status.SetStatusCondition(condition)
}

// InitStatusConditions initializes the status conditions to the provided conditions.
func (r *Alluxio) InitStatusConditions() {
	r.Status.InitStatus(r)
	r.Status.InitStatusConditions()
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Alluxio is the Schema for the alluxios API
type Alluxio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlluxioSpec   `json:"spec,omitempty"`
	Status status.Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AlluxioList contains a list of Alluxio
type AlluxioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Alluxio `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Alluxio{}, &AlluxioList{})
}
