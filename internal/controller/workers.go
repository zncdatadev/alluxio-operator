package controller

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *AlluxioReconciler) makeWorkerDeployment(instance *stackv1alpha1.Alluxio) []*appsv1.Deployment {
	var deployments []*appsv1.Deployment

	if instance.Spec.Worker.RoleGroups != nil {
		for roleGroupName, roleGroup := range instance.Spec.Worker.RoleGroups {
			dep := r.makeWorkerDeploymentForRoleGroup(instance, roleGroupName, roleGroup, r.Scheme)
			if dep != nil {
				deployments = append(deployments, dep)
			}
		}
	}

	return deployments
}

func (r *AlluxioReconciler) makeWorkerDeploymentForRoleGroup(instance *stackv1alpha1.Alluxio, roleGroupName string, roleGroup *stackv1alpha1.RoleWorkerSpec, schema *runtime.Scheme) *appsv1.Deployment {
	labels := instance.GetLabels()

	additionalLabels := make(map[string]string)

	if roleGroup != nil && roleGroup.MatchLabels != nil {
		for k, v := range roleGroup.MatchLabels {
			additionalLabels[k] = v
		}
	}

	mergedLabels := make(map[string]string)
	for key, value := range labels {
		mergedLabels[key] = value
	}
	for key, value := range additionalLabels {
		mergedLabels[key] = value
	}

	var envVars []corev1.EnvVar
	var envFrom []corev1.EnvFromSource
	var shortCircuitEnabled bool
	var needDomainSocketVolume bool

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
				Name: instance.GetNameWithSuffix("config" + "-" + roleGroupName),
			},
		},
	})

	roleGroup = instance.Spec.Worker.GetRoleGroup(instance, roleGroupName)

	if instance != nil && instance.Spec.Worker != nil {
		envVarsMap := make(map[string]string)

		if roleGroup != nil && roleGroup.EnvVars != nil {
			for key, value := range roleGroup.EnvVars {
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

	if roleGroup != nil && roleGroup.HostNetwork != nil && !*roleGroup.HostNetwork {
		envVars = append(envVars, corev1.EnvVar{
			Name: "ALLUXIO_WORKER_CONTAINER_HOSTNAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		})
	}

	volumes := makeTieredStoreVolumes(instance)

	volumeMounts := makeTieredStoreVolumeMounts(instance)
	if needDomainSocketVolume {
		volumes = append(makeShortCircuitVolumes(instance, roleGroupName), volumes...)
		volumeMounts = append(makeShortCircuitVolumeMounts(), volumeMounts...)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix("worker-" + roleGroupName),
			Namespace: instance.Namespace,
			Labels:    mergedLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: roleGroup.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: mergedLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: mergedLabels,
				},
				Spec: corev1.PodSpec{
					SecurityContext:       roleGroup.SecurityContext,
					HostPID:               *roleGroup.HostPID,
					HostNetwork:           *roleGroup.HostNetwork,
					DNSPolicy:             roleGroup.DnsPolicy,
					ShareProcessNamespace: *&roleGroup.ShareProcessNamespace,
					Containers: []corev1.Container{
						{
							Name:            instance.GetNameWithSuffix("worker"),
							Image:           roleGroup.Image.Repository + ":" + roleGroup.Image.Tag,
							ImagePullPolicy: roleGroup.Image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: roleGroup.Ports.Web,
								},
								{
									Name:          "rpc",
									ContainerPort: roleGroup.Ports.Rpc,
								},
							},

							Command:      []string{"tini", "--", "/entrypoint.sh"},
							Args:         roleGroup.Args,
							Resources:    *roleGroup.Resources,
							VolumeMounts: volumeMounts,
						},
						{
							Name:            instance.GetNameWithSuffix("job-worker"),
							Image:           roleGroup.Image.Repository + ":" + roleGroup.Image.Tag,
							ImagePullPolicy: roleGroup.Image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports: []corev1.ContainerPort{
								{
									Name:          "job-rpc",
									ContainerPort: roleGroup.JobWorker.Ports.Rpc,
								},
								{
									Name:          "job-data",
									ContainerPort: roleGroup.JobWorker.Ports.Data,
								},
								{
									Name:          "job-web",
									ContainerPort: roleGroup.JobWorker.Ports.Web,
								},
							},
							Command:      []string{"tini", "--", "/entrypoint.sh"},
							Args:         roleGroup.JobWorker.Args,
							Resources:    *roleGroup.JobWorker.Resources,
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if len(roleGroup.ExtraContainers) > 0 {
		dep.Spec.Template.Spec.Containers = append(dep.Spec.Template.Spec.Containers, roleGroup.ExtraContainers...)
	}

	err := ctrl.SetControllerReference(instance, dep, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for worker deployment")
		return nil
	}
	return dep
}

func (r *AlluxioReconciler) makeWorkerPVCs(instance *stackv1alpha1.Alluxio) ([]*corev1.PersistentVolumeClaim, error) {
	var pvcs []*corev1.PersistentVolumeClaim

	if instance.Spec.Worker.RoleGroups != nil {
		for roleGroupName, roleGroup := range instance.Spec.Worker.RoleGroups {
			pvc := r.makeWorkerPVCForRoleGroup(instance, roleGroupName, roleGroup, r.Scheme)
			if pvc != nil {
				pvcs = append(pvcs, pvc)
			}
		}
	}

	return pvcs, nil
}

func (r *AlluxioReconciler) makeWorkerPVCForRoleGroup(instance *stackv1alpha1.Alluxio, roleGroupName string, roleGroup *stackv1alpha1.RoleWorkerSpec, schema *runtime.Scheme) *corev1.PersistentVolumeClaim {
	labels := instance.GetLabels()

	additionalLabels := make(map[string]string)

	if roleGroup != nil && roleGroup.MatchLabels != nil {
		for k, v := range roleGroup.MatchLabels {
			additionalLabels[k] = v
		}
	}

	mergedLabels := make(map[string]string)
	for key, value := range labels {
		mergedLabels[key] = value
	}
	for key, value := range additionalLabels {
		mergedLabels[key] = value
	}

	shortCircuit := instance.Spec.ClusterConfig.GetShortCircuit()

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      shortCircuit.PvcName + "-" + roleGroupName,
			Namespace: instance.Namespace,
			Labels:    mergedLabels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &shortCircuit.StorageClass,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(shortCircuit.AccessMode)},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: k8sResource.MustParse(shortCircuit.Size),
				},
			},
		},
	}

	err := ctrl.SetControllerReference(instance, pvc, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for worker pvc")
		return nil
	}
	return pvc
}
