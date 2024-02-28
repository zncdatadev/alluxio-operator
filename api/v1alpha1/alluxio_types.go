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
	MasterEmbedded    int32 = 19200
	MasterRpcPort     int32 = 19998
	MasterWebPort     int32 = 19999
	JobMasterRpcPort  int32 = 20001
	JobMasterWebPort  int32 = 20002
	JobMasterEmbedded int32 = 20003
	WorkerRpcPort     int32 = 29999
	WorkerWebPort     int32 = 30000
	JobWorkerRpcPort  int32 = 30001
	JobWorkerDataPort int32 = 30002
	JobWorkerWebPort  int32 = 30003
)

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

// AlluxioSpec defines the desired state of Alluxio
type AlluxioSpec struct {
	// +kubebuilder:validation:Required
	ClusterConfig *ClusterConfigSpec `json:"clusterConfig,omitempty"`

	// +kubebuilder:validation:Optional
	Image *ImageSpec `json:"image,omitempty"`

	// +kubebuilder:validation:Required
	Master *MasterSpec `json:"master,omitempty"`

	// +kubebuilder:validation:Required
	Worker *WorkerSpec `json:"worker,omitempty"`
}

type MasterSpec struct {
	// +kubebuilder:validation:Optional
	Config *MasterConfigSpec `json:"config,omitempty"`

	RoleGroups map[string]*MasterRoleGroupSpec `json:"roleGroups,omitempty"`

	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
	// +kubebuilder:validation:Optional
	CommandArgsOverrides []string `json:"commandArgsOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	ConfigOverrides *ConfigOverridesSpec `json:"configOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	EnvOverrides map[string]string `json:"envOverrides,omitempty"`
}

type WorkerSpec struct {
	// +kubebuilder:validation:Optional
	Config *WorkerConfigSpec `json:"config,omitempty"`

	RoleGroups map[string]*WorkerRoleGroupSpec `json:"roleGroups,omitempty"`

	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`
	// +kubebuilder:validation:Optional
	CommandArgsOverrides []string `json:"commandArgsOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	ConfigOverrides *ConfigOverridesSpec `json:"configOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	EnvOverrides map[string]string `json:"envOverrides,omitempty"`
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

func (r *Alluxio) GetNameWithSuffix(suffix string) string {
	// return sparkHistory.GetName() + rand.String(5) + suffix
	return r.GetName() + "-" + suffix
}

type WorkerPortsSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=29999
	Rpc int32 `json:"rpc,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=30000
	Web int32 `json:"web,omitempty"`
}

type JobMasterSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"job-master"}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	Properties map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *ResourcesSpec `json:"resources,omitempty"`

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

type JobWorkerSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"job-worker"}
	Args []string `json:"args,omitempty"`

	// +kubebuilder:validation:Optional
	Properties map[string]string `json:"properties,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *ResourcesSpec `json:"resources,omitempty"`

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

type PodDisruptionBudgetSpec struct {
	// +kubebuilder:validation:Optional
	MinAvailable int32 `json:"minAvailable,omitempty"`

	// +kubebuilder:validation:Optional
	MaxUnavailable int32 `json:"maxUnavailable,omitempty"`
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
		VolumeType:   "persistentVolumeClaim",
		Size:         "100Mi",
		PvcName:      "alluxio-worker-domain-socket",
		StorageClass: "local-path",
		AccessMode:   "ReadWriteOnce",
		HosePath:     "/tmp/alluxio-domain",
	}
}

func (clusterConfig *ClusterConfigSpec) GetJournal() JournalSpec {
	return JournalSpec{
		Type:         "UFS",
		Folder:       "/journal",
		UfsType:      "local",
		VolumeType:   "persistentVolumeClaim",
		Size:         "1Gi",
		StorageClass: "local-path",
		AccessMode:   "ReadWriteOnce",
		Medium:       "",
		RunFormat:    false,
	}
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

func init() {
	SchemeBuilder.Register(&Alluxio{}, &AlluxioList{})
}
