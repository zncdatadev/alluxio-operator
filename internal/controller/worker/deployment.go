package worker

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/alluxio-operator/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeploymentReconciler struct {
	common.DeploymentStyleReconciler[*stackv1alpha1.Alluxio, *stackv1alpha1.WorkerRoleGroupSpec]
}

// NewDeployment New a StatefulSet
func NewDeployment(
	scheme *runtime.Scheme,
	instance *stackv1alpha1.Alluxio,
	client client.Client,
	groupName string,
	mergedLabels map[string]string,
	mergedCfg *stackv1alpha1.WorkerRoleGroupSpec,
	replicas int32,
) *DeploymentReconciler {
	return &DeploymentReconciler{
		DeploymentStyleReconciler: *common.NewDeploymentStyleReconciler[*stackv1alpha1.Alluxio,
			*stackv1alpha1.WorkerRoleGroupSpec](
			scheme,
			instance,
			client,
			groupName,
			mergedLabels,
			mergedCfg,
			replicas),
	}
}

func (d *DeploymentReconciler) GetConditions() *[]metav1.Condition {
	return &d.Instance.Status.Conditions
}

func (d *DeploymentReconciler) Build() (client.Object, error) {

	groupName := d.GroupName
	mergedGroupCfg := d.MergedCfg
	mergedConfigSpec := mergedGroupCfg.Config
	instance := d.Instance

	isShortCircuitEnabled := d.isShortCircuitEnabled()
	needDomainSocketVolume := d.isNeedDomainSocketVolume(isShortCircuitEnabled)
	envVars := d.createEnvVars(instance, d.MergedCfg)
	envFrom := d.createEnvFrom(groupName)
	volumes, volumeMounts := d.createVolumesAndMounts(needDomainSocketVolume, groupName, instance)
	image := instance.Spec.Image

	//todo: webhook opt
	workerPorts := common.GetWorkerPorts(mergedGroupCfg)
	jobWorkerPorts := common.GetJobWorkerPorts(mergedGroupCfg)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createDeploymentName(instance.GetName(), string(common.Worker), groupName),
			Namespace: instance.Namespace,
			Labels:    d.MergedLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &mergedGroupCfg.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: d.MergedLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: d.MergedLabels,
				},
				Spec: corev1.PodSpec{
					SecurityContext:       mergedConfigSpec.SecurityContext,
					HostPID:               *mergedConfigSpec.HostPID,
					HostNetwork:           *mergedConfigSpec.HostNetwork,
					DNSPolicy:             mergedConfigSpec.DnsPolicy,
					ShareProcessNamespace: mergedConfigSpec.ShareProcessNamespace,
					Containers: []corev1.Container{
						{
							Name:            instance.GetNameWithSuffix("worker"),
							Image:           image.Repository + ":" + image.Tag,
							ImagePullPolicy: image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: workerPorts.Web,
								},
								{
									Name:          "rpc",
									ContainerPort: workerPorts.Rpc,
								},
							},
							Command:      []string{"tini", "--", "/entrypoint.sh"},
							Args:         d.getWorkerArgs(),
							Resources:    *common.ConvertToResourceRequirements(mergedConfigSpec.Resources),
							VolumeMounts: d.createWorkerVolumeMounts(volumeMounts),
						},
						{
							Name:            instance.GetNameWithSuffix("job-worker"),
							Image:           image.Repository + ":" + image.Tag,
							ImagePullPolicy: image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "job-rpc",
									ContainerPort: jobWorkerPorts.Rpc,
								},
								{
									Name:          "job-data",
									ContainerPort: jobWorkerPorts.Data,
								},
								{
									Name:          "job-web",
									ContainerPort: jobWorkerPorts.Web,
								},
							},
							Command:      []string{"tini", "--", "/entrypoint.sh"},
							Args:         d.getJobWorkerArgs(),
							Resources:    *common.ConvertToResourceRequirements(mergedConfigSpec.JobWorker.Resources),
							VolumeMounts: d.createJobWorkerVolumeMounts(volumeMounts),
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if len(mergedConfigSpec.ExtraContainers) > 0 {
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, mergedConfigSpec.ExtraContainers...)
	}
	d.schedulePod(dep)
	return dep, nil
}

// get worker args
func (d *DeploymentReconciler) getWorkerArgs() []string {
	args := d.MergedCfg.Config.Args
	if len(args) == 0 {
		return []string{"worker-only", "--no-format"}
	}
	return args
}

// get job worker args
func (d *DeploymentReconciler) getJobWorkerArgs() []string {
	args := d.MergedCfg.Config.JobWorker.Args
	if len(args) == 0 {
		return []string{"job-worker"}
	}
	return args
}

// schedulePod is used to schedule pod, such as affinity, tolerations, nodeSelector
func (d *DeploymentReconciler) schedulePod(obj *appsv1.Deployment) {
	mergedGroupCfg := d.MergedCfg
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

// CommandOverride commandOverride only deployment and statefulset need to implement this method
// todo: set the same command for all containers currently
func (d *DeploymentReconciler) CommandOverride(obj client.Object) {
	dep := obj.(*appsv1.Deployment)
	containers := dep.Spec.Template.Spec.Containers
	if cmdOverride := d.MergedCfg.CommandArgsOverrides; cmdOverride != nil {
		for i := range containers {
			containers[i].Command = cmdOverride
		}
	}
}

// EnvOverride only deployment and statefulset need to implement this method
// todo: set the same env for all containers currently
func (d *DeploymentReconciler) EnvOverride(obj client.Object) {
	dep := obj.(*appsv1.Deployment)
	containers := dep.Spec.Template.Spec.Containers
	if envOverride := d.MergedCfg.EnvOverrides; envOverride != nil {
		for i := range containers {
			envVars := containers[i].Env
			common.OverrideEnvVars(&envVars, d.MergedCfg.EnvOverrides)
		}
	}
}

func (d *DeploymentReconciler) RoleGroupConfig() *stackv1alpha1.WorkerConfigSpec {
	return d.MergedCfg.Config
}

func (d *DeploymentReconciler) EnableWorkerLogging() bool {
	return d.RoleGroupConfig() != nil &&
		d.RoleGroupConfig().Logging != nil &&
		d.RoleGroupConfig().Logging.Metastore != nil
}

// EnableJobWorkerLogging is job worker enabled logging
func (d *DeploymentReconciler) EnableJobWorkerLogging() bool {
	return d.MergedCfg.Config.JobWorker.Logging != nil && d.MergedCfg.Config.JobWorker.Logging.Metastore != nil
}

func (d *DeploymentReconciler) LogOverride(resource client.Object) {
	statefulSet := resource.(*appsv1.Deployment)
	volumes := statefulSet.Spec.Template.Spec.Volumes
	keyToPath := d.createWorkerVolumeKeyToPath()
	if len(keyToPath) > 0 {
		log4jVolume := corev1.Volume{
			Name: common.Log4jVolumeName(),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: common.CreateRoleGroupLoggingConfigMapName(d.Instance.GetName(), string(common.Worker),
							d.GroupName),
					},
					Items: keyToPath,
				},
			},
		}
		volumes = append(volumes, log4jVolume)
		statefulSet.Spec.Template.Spec.Volumes = volumes
	}
}

// is short circuit enabled
func (d *DeploymentReconciler) isShortCircuitEnabled() bool {
	return d.Instance.Spec.ClusterConfig != nil && d.Instance.Spec.ClusterConfig.GetShortCircuit().Enabled
}

// is need domain socket volume
func (d *DeploymentReconciler) isNeedDomainSocketVolume(isShortCircuitEnabled bool) bool {
	return isShortCircuitEnabled && d.Instance.Spec.ClusterConfig.GetShortCircuit().Policy == "uuid"
}

// create env form
func (d *DeploymentReconciler) createEnvFrom(groupName string) []corev1.EnvFromSource {
	var envFrom []corev1.EnvFromSource
	envFrom = append(envFrom, corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: common.CreateMasterConfigMapName(d.Instance.GetName(), groupName),
			},
		},
	})
	return envFrom
}

// create env vars
func (d *DeploymentReconciler) createEnvVars(
	instance *stackv1alpha1.Alluxio,
	mergedGroupCfg *stackv1alpha1.WorkerRoleGroupSpec) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	envVars = append(envVars, corev1.EnvVar{
		Name: "ALLUXIO_WORKER_HOSTNAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	})

	if instance != nil && instance.Spec.Worker != nil {
		envVarsMap := make(map[string]string)

		if mergedGroupCfg != nil && mergedGroupCfg.Config.EnvVars != nil {
			for key, value := range mergedGroupCfg.Config.EnvVars {
				envVarsMap[key] = value
			}
		}

		for key, value := range envVarsMap {
			envVars = append(envVars, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}

	if mergedGroupCfg != nil && mergedGroupCfg.Config.HostNetwork != nil && !*mergedGroupCfg.Config.HostNetwork {
		envVars = append(envVars, corev1.EnvVar{
			Name: "ALLUXIO_WORKER_CONTAINER_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		})
	}
	return envVars
}

// create volumes and volume mounts
func (d *DeploymentReconciler) createVolumesAndMounts(
	needDomainSocketVolume bool,
	groupName string, instance *stackv1alpha1.Alluxio) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := MakeTieredStoreVolumes(instance)
	volumeMounts := MakeTieredStoreVolumeMounts(instance)
	if needDomainSocketVolume {
		volumes = append(MakeShortCircuitVolumes(instance, groupName), volumes...)
		volumeMounts = append(MakeShortCircuitVolumeMounts(), volumeMounts...)
	}
	return volumes, volumeMounts
}

func (d *DeploymentReconciler) createWorkerVolumeMounts(exists []corev1.VolumeMount) []corev1.VolumeMount {
	if d.EnableWorkerLogging() {
		exists = append(exists, common.CreateAlluxioLoggerVolumeMounts())
	}
	return exists
}

func (d *DeploymentReconciler) createJobWorkerVolumeMounts(exists []corev1.VolumeMount) []corev1.VolumeMount {
	if d.EnableJobWorkerLogging() {
		exists = append(exists, common.CreateAlluxioLoggerVolumeMounts())
	}
	return exists
}
func (d *DeploymentReconciler) createWorkerVolumeKeyToPath() []corev1.KeyToPath {
	var res []corev1.KeyToPath
	if d.EnableWorkerLogging() {
		res = append(res, corev1.KeyToPath{
			Key:  common.CreateLoggerConfigMapKey(common.WorkerLogger),
			Path: common.Log4jCfgName,
		})
	}
	if d.EnableJobWorkerLogging() {
		res = append(res, corev1.KeyToPath{
			Key:  common.CreateLoggerConfigMapKey(common.JobWorkerLogger),
			Path: common.Log4jCfgName,
		})
	}
	return res
}
