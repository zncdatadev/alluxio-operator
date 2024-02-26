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
			mergedLabels,
			mergedCfg,
			replicas),
	}
}

func (d *DeploymentReconciler) GetConditions() *[]metav1.Condition {
	return &d.Instance.Status.Conditions
}

func (d *DeploymentReconciler) Build(data common.ResourceBuilderData) (client.Object, error) {
	var envVars []corev1.EnvVar
	var envFrom []corev1.EnvFromSource
	var shortCircuitEnabled bool
	var needDomainSocketVolume bool

	roleGroupName := data.GroupName
	mergedGroupCfg := d.MergedCfg
	mergedConfigSpec := mergedGroupCfg.Config
	instance := d.Instance

	if instance.Spec.ClusterConfig != nil && instance.Spec.ClusterConfig.GetShortCircuit().Enabled {
		shortCircuitEnabled = true
	}

	if shortCircuitEnabled && instance.Spec.ClusterConfig.GetShortCircuit().Policy == "uuid" {
		needDomainSocketVolume = true
	}

	envVars = append(envVars, corev1.EnvVar{
		Name: "ALLUXIO_WORKER_HOSTNAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	})

	envFrom = append(envFrom, corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: common.CreateMasterConfigMapName(instance.GetName(), roleGroupName),
			},
		},
	})

	roleGroup := instance.Spec.Worker.RoleGroups[roleGroupName]

	if instance != nil && instance.Spec.Worker != nil {
		envVarsMap := make(map[string]string)

		if roleGroup != nil && roleGroup.Config.EnvVars != nil {
			for key, value := range roleGroup.Config.EnvVars {
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

	if roleGroup != nil && roleGroup.Config.HostNetwork != nil && !*roleGroup.Config.HostNetwork {
		envVars = append(envVars, corev1.EnvVar{
			Name: "ALLUXIO_WORKER_CONTAINER_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		})
	}
	volumes := MakeTieredStoreVolumes(instance)

	volumeMounts := MakeTieredStoreVolumeMounts(instance)
	if needDomainSocketVolume {
		volumes = append(MakeShortCircuitVolumes(instance, roleGroupName), volumes...)
		volumeMounts = append(MakeShortCircuitVolumeMounts(), volumeMounts...)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createDeploymentName(instance.GetName(), d.RoleName, roleGroupName),
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
							Image:           mergedConfigSpec.Image.Repository + ":" + mergedConfigSpec.Image.Tag,
							ImagePullPolicy: mergedConfigSpec.Image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: mergedConfigSpec.Ports.Web,
								},
								{
									Name:          "rpc",
									ContainerPort: mergedConfigSpec.Ports.Rpc,
								},
							},

							Command:      []string{"tini", "--", "/entrypoint.sh"},
							Args:         mergedConfigSpec.Args,
							Resources:    *common.ConvertToResourceRequirements(mergedConfigSpec.Resources),
							VolumeMounts: volumeMounts,
						},
						{
							Name:            instance.GetNameWithSuffix("job-worker"),
							Image:           mergedConfigSpec.Image.Repository + ":" + mergedConfigSpec.Image.Tag,
							ImagePullPolicy: mergedConfigSpec.Image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "job-rpc",
									ContainerPort: mergedConfigSpec.JobWorker.Ports.Rpc,
								},
								{
									Name:          "job-data",
									ContainerPort: mergedConfigSpec.JobWorker.Ports.Data,
								},
								{
									Name:          "job-web",
									ContainerPort: mergedConfigSpec.JobWorker.Ports.Web,
								},
							},
							Command:      []string{"tini", "--", "/entrypoint.sh"},
							Args:         mergedConfigSpec.JobWorker.Args,
							Resources:    *mergedConfigSpec.JobWorker.Resources,
							VolumeMounts: volumeMounts,
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

// schedulePod is used to schedule pod, such as affinity, tolerations, nodeSelector
func (d *DeploymentReconciler) schedulePod(obj *appsv1.Deployment) {
	mergedGroupCfg := d.MergedCfg
	mergedConfigSpec := mergedGroupCfg.Config
	if mergedGroupCfg != nil {
		if mergedConfigSpec.Affinity != nil {
			obj.Spec.Template.Spec.Affinity = mergedConfigSpec.Affinity
		}

		if mergedConfigSpec.Tolerations != nil {
			obj.Spec.Template.Spec.Tolerations = mergedConfigSpec.Tolerations
		}

		if mergedConfigSpec.NodeSelector != nil {
			obj.Spec.Template.Spec.NodeSelector = mergedConfigSpec.NodeSelector
		}
	}
}

// commandOverride only deployment and statefulset need to implement this method
// todo: set the same command for all containers currently
func (d *DeploymentReconciler) commandOverride(obj client.Object) {
	statefulSet := obj.(*appsv1.StatefulSet)
	containers := statefulSet.Spec.Template.Spec.Containers
	if cmdOverride := d.MergedCfg.CommandArgsOverrides; cmdOverride != nil {
		for i := range containers {
			containers[i].Command = cmdOverride
		}
	}
}

// EnvOverride only deployment and statefulset need to implement this method
// todo: set the same env for all containers currently
func (d *DeploymentReconciler) EnvOverride(obj client.Object) {
	statefulSet := obj.(*appsv1.StatefulSet)
	containers := statefulSet.Spec.Template.Spec.Containers
	if envOverride := d.MergedCfg.EnvOverrides; envOverride != nil {
		for i := range containers {
			envVars := containers[i].Env
			common.OverrideEnvVars(envVars, d.MergedCfg.EnvOverrides)
		}
	}
}
