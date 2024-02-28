package master

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/alluxio-operator/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSetReconciler struct {
	common.DeploymentStyleReconciler[*stackv1alpha1.Alluxio, *stackv1alpha1.MasterRoleGroupSpec]
}

// NewStatefulSet New a StatefulSetReconciler
func NewStatefulSet(
	scheme *runtime.Scheme,
	instance *stackv1alpha1.Alluxio,
	client client.Client,
	mergedLabels map[string]string,
	mergedCfg *stackv1alpha1.MasterRoleGroupSpec,
	replicas int32,
) *StatefulSetReconciler {
	return &StatefulSetReconciler{
		DeploymentStyleReconciler: *common.NewDeploymentStyleReconciler[*stackv1alpha1.Alluxio,
			*stackv1alpha1.MasterRoleGroupSpec](
			scheme,
			instance,
			client,
			mergedLabels,
			mergedCfg,
			replicas),
	}
}

func (s *StatefulSetReconciler) GetConditions() *[]metav1.Condition {
	return &s.Instance.Status.Conditions
}

// CommandOverride commandOverride only deployment and statefulset need to implement this method
// todo: set the same command for all containers currently
func (s *StatefulSetReconciler) CommandOverride(obj client.Object) {
	statefulSet := obj.(*appsv1.StatefulSet)
	containers := statefulSet.Spec.Template.Spec.Containers
	if cmdOverride := s.MergedCfg.CommandArgsOverrides; cmdOverride != nil {
		for i := range containers {
			containers[i].Command = cmdOverride
		}
	}
}

// EnvOverride only deployment and statefulset need to implement this method
// todo: set the same env for all containers currently
func (s *StatefulSetReconciler) EnvOverride(obj client.Object) {
	statefulSet := obj.(*appsv1.StatefulSet)
	containers := statefulSet.Spec.Template.Spec.Containers
	if envOverride := s.MergedCfg.EnvOverrides; envOverride != nil {
		for i := range containers {
			envVars := containers[i].Env
			common.OverrideEnvVars(envVars, s.MergedCfg.EnvOverrides)
		}
	}
}

func (s *StatefulSetReconciler) Build(data common.ResourceBuilderData) (client.Object, error) {
	instance := s.Instance
	mergedGroupCfg := s.MergedCfg
	mergedConfigSpec := mergedGroupCfg.Config

	isUfsLocal := isUfsLocal(instance.Spec.ClusterConfig)
	isEmbedded := isEmbedded(instance.Spec.ClusterConfig)
	isSingleMaster := isSingleMaster(mergedGroupCfg)
	needJournalVolume := needJournalVolume(isUfsLocal, isEmbedded)
	isHaEmbedded := isHaEmbedded(isEmbedded, mergedGroupCfg.Replicas)
	image := instance.Spec.Image
	journal := common.GetJournal(instance.Spec.ClusterConfig)
	app := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createMasterStatefulSetName(s.Instance.GetName(), s.RoleName, data.GroupName),
			Namespace: instance.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &mergedGroupCfg.Replicas,
			ServiceName: instance.GetName() + "svc-master-headless",
			Selector: &metav1.LabelSelector{
				MatchLabels: data.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: data.Labels,
				},
				Spec: corev1.PodSpec{
					SecurityContext:       mergedConfigSpec.SecurityContext,
					HostPID:               *mergedConfigSpec.HostPID,
					HostNetwork:           *mergedConfigSpec.HostNetwork,
					DNSPolicy:             mergedConfigSpec.DnsPolicy,
					ShareProcessNamespace: mergedConfigSpec.ShareProcessNamespace,
					Containers: []corev1.Container{
						{
							Name:            instance.GetNameWithSuffix("master"),
							Image:           image.Repository + ":" + image.Tag,
							ImagePullPolicy: image.PullPolicy,
							Env:             s.createEnvVars(isHaEmbedded, isSingleMaster, mergedGroupCfg),
							EnvFrom:         createEnvFrom(instance, data.GroupName),
							Ports:           createMasterPorts(mergedGroupCfg, isHaEmbedded),
							Command:         []string{"tini", "--", "/entrypoint.sh"},
							Args:            s.getMasterCmdArgs(mergedGroupCfg),
							Resources:       *common.ConvertToResourceRequirements(mergedConfigSpec.Resources),
							VolumeMounts:    createVolumeMount(needJournalVolume, journal),
						},
						{
							Name:            instance.GetNameWithSuffix("job-master"),
							Image:           image.Repository + ":" + image.Tag,
							ImagePullPolicy: image.PullPolicy,
							Env:             s.createEnvVars(isHaEmbedded, isSingleMaster, mergedGroupCfg),
							EnvFrom:         createEnvFrom(instance, data.GroupName),
							Ports:           createJobMasterPorts(mergedGroupCfg, isHaEmbedded),
							Command:         []string{"tini", "--", "/entrypoint.sh"},
							Args:            s.getJobMasterCmdArgs(mergedGroupCfg),
							Resources:       *common.ConvertToResourceRequirements(mergedConfigSpec.JobMaster.Resources),
						},
					},
					Volumes: createVolumes(needJournalVolume, journal),
				},
			},
			VolumeClaimTemplates: createVolumeClaimTemplates(needJournalVolume, journal),
		},
	}
	s.schedulePod(app)
	return app, nil
}

// schedulePod is used to schedule pod, such as affinity, tolerations, nodeSelector
func (s *StatefulSetReconciler) schedulePod(obj *appsv1.StatefulSet) {
	mergedGroupCfg := s.MergedCfg
	if mergedGroupCfg != nil && mergedGroupCfg.Config != nil {
		if affinity := mergedGroupCfg.Config.Affinity; affinity != nil {
			obj.Spec.Template.Spec.Affinity = affinity
		}
		if toleration := mergedGroupCfg.Config.Tolerations; toleration != nil {
			obj.Spec.Template.Spec.Tolerations = toleration
		}
		if nodeSelector := mergedGroupCfg.Config.NodeSelector; nodeSelector != nil {
			obj.Spec.Template.Spec.NodeSelector = nodeSelector
		}
	}
}

func (s *StatefulSetReconciler) getMasterCmdArgs(cfg *stackv1alpha1.MasterRoleGroupSpec) []string {
	args := cfg.Config.Args
	if len(args) == 0 {
		return []string{"master-only", "--no-format"}
	}
	return args
}

func (s *StatefulSetReconciler) getJobMasterCmdArgs(cfg *stackv1alpha1.MasterRoleGroupSpec) []string {
	args := cfg.Config.JobMaster.Args
	if len(args) == 0 {
		return []string{"job-master"}
	}
	return args
}

// isUfsLocal is UFS local based on Journal's type and UfsType
func isUfsLocal(clusterCfg *stackv1alpha1.ClusterConfigSpec) bool {
	if clusterCfg.GetJournal().Type == "UFS" && clusterCfg.GetJournal().UfsType == "local" {
		return true
	}
	return false
}

// is Embedded based on Journal's type
func isEmbedded(clusterCfg *stackv1alpha1.ClusterConfigSpec) bool {
	return clusterCfg.GetJournal().Type == "EMBEDDED"
}

// isSingleMaster based on Replicas
func isSingleMaster(mergedGroupCfg *stackv1alpha1.MasterRoleGroupSpec) bool {
	return mergedGroupCfg.Replicas == 1
}

// is need Journal Volume based on isUfsLocal and isEmbedded
func needJournalVolume(isUfsLocal bool, isEmbedded bool) bool {
	if isUfsLocal || isEmbedded {
		return true
	}
	return false
}

// isHaEmbedded based on isEmbedded and Replicas
func isHaEmbedded(isEmbedded bool, replicas int32) bool {
	if isEmbedded && replicas > 1 {
		return true
	}
	return false
}

// create and add volumeMount if needJournalVolume is true
func createVolumeMount(needJournalVolume bool, journal *stackv1alpha1.JournalSpec) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount
	if needJournalVolume {
		vm := corev1.VolumeMount{
			Name:      "alluxio-journal",
			MountPath: journal.Folder,
		}
		volumeMounts = append(volumeMounts, vm)
	}
	return volumeMounts
}

// Create and add volume if needJournalVolume is true and VolumeType is "emptyDir"
func createVolumes(needJournalVolume bool, journal *stackv1alpha1.JournalSpec) []corev1.Volume {
	var volumes []corev1.Volume
	if needJournalVolume && journal.VolumeType == "emptyDir" {
		volume := corev1.Volume{
			Name: "alluxio-journal",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		volumes = append(volumes, volume)
	}
	return volumes
}

// Create and add PersistentVolumeClaim if needJournalVolume is true and VolumeType is "persistentVolumeClaim"
func createVolumeClaimTemplates(needJournalVolume bool, journal *stackv1alpha1.JournalSpec) []corev1.PersistentVolumeClaim {
	var volumeClaimTemplates []corev1.PersistentVolumeClaim
	if needJournalVolume && journal.VolumeType == "persistentVolumeClaim" {
		accessMode := corev1.PersistentVolumeAccessMode(journal.AccessMode)
		pvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "alluxio-journal",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				StorageClassName: &journal.StorageClass,
				AccessModes:      []corev1.PersistentVolumeAccessMode{accessMode},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(journal.Size),
					},
				},
			},
		}
		volumeClaimTemplates = append(volumeClaimTemplates, pvc)
	}
	return volumeClaimTemplates
}

// Create and add envVars
func (s *StatefulSetReconciler) createEnvVars(
	isHaEmbedded bool,
	isSingleMaster bool,
	mergedGroupCfg *stackv1alpha1.MasterRoleGroupSpec) []corev1.EnvVar {
	var envVarsMap = make(map[string]string)
	var envVars []corev1.EnvVar
	if isHaEmbedded {
		envVars = append(envVars, corev1.EnvVar{
			Name: "ALLUXIO_MASTER_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		})
	} else if isSingleMaster {
		envVars = append(envVars, corev1.EnvVar{
			Name: "ALLUXIO_MASTER_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		})
	}
	mergedConfigSpec := mergedGroupCfg.Config
	if mergedConfigSpec.EnvVars != nil {
		for key, value := range mergedConfigSpec.EnvVars {
			envVarsMap[key] = value
		}
		for key, value := range envVarsMap {
			envVars = append(envVars, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}
	return envVars
}

// create env from
func createEnvFrom(instance *stackv1alpha1.Alluxio, groupName string) []corev1.EnvFromSource {
	var envFrom []corev1.EnvFromSource
	envFrom = append(envFrom, corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: common.CreateMasterConfigMapName(instance.GetName(), groupName),
			},
		},
	})
	return envFrom
}

// create master ports
func createMasterPorts(mergedConfigSpec *stackv1alpha1.MasterRoleGroupSpec, isHaEmbedded bool) []corev1.ContainerPort {
	materPort := getMasterPorts(mergedConfigSpec)
	masterPorts := []corev1.ContainerPort{
		{
			Name:          "web",
			ContainerPort: materPort.Web,
		},
		{
			Name:          "rpc",
			ContainerPort: materPort.Rpc,
		},
	}
	if isHaEmbedded {
		masterPorts = append(masterPorts, corev1.ContainerPort{
			Name:          "embedded",
			ContainerPort: materPort.Embedded,
		})
	}
	return masterPorts
}

// create job master ports
func createJobMasterPorts(mergedConfigSpec *stackv1alpha1.MasterRoleGroupSpec, isHaEmbedded bool) []corev1.ContainerPort {
	jobMasterPort := getJobMasterPorts(mergedConfigSpec)
	jobMasterPorts := []corev1.ContainerPort{
		{
			Name:          "job-web",
			ContainerPort: jobMasterPort.Web,
		},
		{
			Name:          "job-rpc",
			ContainerPort: jobMasterPort.Rpc,
		},
	}
	if isHaEmbedded {
		jobMasterPorts = append(jobMasterPorts, corev1.ContainerPort{
			Name:          "job-embedded",
			ContainerPort: jobMasterPort.Embedded,
		})
	}
	return jobMasterPorts
}
