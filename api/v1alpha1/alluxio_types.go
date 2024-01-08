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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlluxioSpec defines the desired state of Alluxio
type AlluxioSpec struct {
	Image           *ImageSpec           `json:"image"`
	SecurityContext *SecurityContextSpec `json:"securityContext,omitempty"`
	Properties      map[string]string    `json:"properties,omitempty"`
	Master          map[string]RoleGroupSpec 
	Worker          *WorkerSpec          `json:"worker,omitempty"`
	JobMaster       *JobMasterSpec       `json:"jobMaster,omitempty"`
	JobWorker       *JobWorkerSpec       `json:"jobWorker,omitempty"`

	TieredStore  []*TieredStore    `json:"tieredStore,omitempty"`
	ShortCircuit *ShortCircuitSpec `json:"shortCircuit,omitempty"`
}

func (alluxio *Alluxio) GetLabels() map[string]string {
	return map[string]string{
		"app":      alluxio.GetName(),
		"provider": "zncdata",
	}
}

type RoleGroupSpec struct {
	Replicas  int32 `json:"replicas"`
	
}

type MasterSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"master-only", "--no-format"}
	Args []string `json:"args,omitempty"`
	// +kubebuilder:validation:Optional
	Properties *map[string]string `json:"properties,omitempty"`

	Resources *MasterResourceSpec `json:"resources,omitempty"`

	Ports *MasterPortsSpec `json:"ports,omitempty"`
}

type MasterPortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=19200
	Embedded int32 `json:"embedded,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=19998
	RpcPort int32 `json:"rpcPort,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=19999
	WebPort int32 `json:"debugPort,omitempty"`
}

type MasterResourceSpec struct {
	Limits   *MasterResourcesLimitSpec  `json:"limits,omitempty"`
	Requests *MasterResourceRequestSpec `json:"requests,omitempty"`
}

type MasterResourcesLimitSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4000m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4Gi"
	Memory string `json:"memory,omitempty"`
}

type MasterResourceRequestSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500Mi"
	Memory string `json:"memory,omitempty"`
}

//// MasterArgs returns the args for the master process.
//func (master *MasterSpec) MasterArgs() []string {
//	if master.Args != nil {
//		return master.Args
//	} else {
//		return []string{
//			"master-only",
//			"--no-format",
//		}
//	}
//}

type SecurityContextSpec struct {
	// The UID to run the entrypoint of the container process.
	// Defaults to user specified in image metadata if unspecified.
	// May also be set in SecurityContext.  If set in both SecurityContext and
	// PodSecurityContext, the value specified in SecurityContext takes precedence
	// for that container.
	// Note that this field cannot be set when spec.os.name is windows.
	// +optional
	// +kubebuilder:default=1000
	RunAsUser *int64 `json:"runAsUser,omitempty"`
	// The GID to run the entrypoint of the container process.
	// Uses runtime default if unset.
	// May also be set in SecurityContext.  If set in both SecurityContext and
	// PodSecurityContext, the value specified in SecurityContext takes precedence
	// for that container.
	// Note that this field cannot be set when spec.os.name is windows.
	// +optional
	// +kubebuilder:default=1000
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
	// A special supplemental group that applies to all containers in a pod.
	// Some volume types allow the Kubelet to change the ownership of that volume
	// to be owned by the pod:
	//
	// 1. The owning GID will be the FSGroup
	// 2. The setgid bit is set (new files created in the volume will be owned by FSGroup)
	// 3. The permission bits are OR'd with rw-rw----
	//
	// If unset, the Kubelet will not modify the ownership and permissions of any volume.
	// Note that this field cannot be set when spec.os.name is windows.
	// +optional
	// +kubebuilder:default=1000
	FSGroup *int64 `json:"fsGroup,omitempty"`
}

type WorkerSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"worker", "--no-format"}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	Properties *map[string]string `json:"properties,omitempty"`

	Resources *ResourceSpec `json:"resources,omitempty"`

	// +kube:validation:Optional
	Ports *WorkPortsSpec `json:"ports,omitempty"`
}

type WorkPortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=29999
	Rpc int32 `json:"rpc,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30000
	Web int32 `json:"web,omitempty"`
}

type ResourceSpec struct {
	Limits   *WorkResourcesLimitSpec  `json:"limits,omitempty"`
	Requests *WorkResourceRequestSpec `json:"requests,omitempty"`
}

type WorkResourcesLimitSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4000m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4Gi"
	Memory string `json:"memory,omitempty"`
}

type WorkResourceRequestSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500Mi"
	Memory string `json:"memory,omitempty"`
}

//// WorkArgs returns the args for the worker process.
//func (work *WorkSpec) WorkArgs() []string {
//	if work.Args != nil {
//		return work.Args
//	} else {
//		return []string{
//			"worker",
//			"--no-format",
//		}
//	}
//}

type JobMasterSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"job-master"}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	Properties *map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *JobMasterResourceSpec `json:"resources,omitempty"`

	Ports *JobMasterPortsSpec `json:"ports,omitempty"`
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

type JobMasterResourceSpec struct {
	Limits   *JobMasterResourcesLimitSpec  `json:"limits,omitempty"`
	Requests *JobMasterResourceRequestSpec `json:"requests,omitempty"`
}

type JobMasterResourcesLimitSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4000m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4Gi"
	Memory string `json:"memory,omitempty"`
}

type JobMasterResourceRequestSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500Mi"
	Memory string `json:"memory,omitempty"`
}

//// JobMasterArgs returns the args for the job master process.
//func (jobMaster *JobMasterSpec) JobMasterArgs() []string {
//	if jobMaster.Args != nil {
//		return jobMaster.Args
//	} else {
//		return []string{
//			"job-master",
//		}
//	}
//}

type JobWorkerSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"job-worker"}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	Properties *map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *JobWorkerResourceSpec `json:"resources,omitempty"`

	Ports *JobWorkerPortsSpec `json:"ports,omitempty"`
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

type JobWorkerResourceSpec struct {
	Limits   *JobWorkerResourcesLimitSpec  `json:"limits,omitempty"`
	Requests *JobWorkerResourceRequestSpec `json:"requests,omitempty"`
}

type JobWorkerResourcesLimitSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4000m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="4Gi"
	Memory string `json:"memory,omitempty"`
}

type JobWorkerResourceRequestSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500m"
	CPU string `json:"cpu,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="500Mi"
	Memory string `json:"memory,omitempty"`
}

//// JobWorkerArgs returns the args for the job worker process.
//func (jobWorker *JobWorkerSpec) JobWorkerArgs() []string {
//	if jobWorker.Args != nil {
//		return jobWorker.Args
//	} else {
//		return []string{
//			"job-worker",
//		}
//	}
//}
//

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
	Name       string  `json:"name"`
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
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="uuid"
	Policy string `json:"policy"`

	// +kubebuilder:validation:hostPath,persistentVolumeClaim
	// +kubebuilder:default="hostPath"
	VolumeType string `json:"volumeType"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="100Mi"
	Size string `json:"size"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/tmp/"
	HosePath string `json:"path"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Alluxio is the Schema for the alluxios API
type Alluxio struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlluxioSpec   `json:"spec,omitempty"`
	Status AlluxioStatus `json:"status,omitempty"`
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
