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

func (r *AlluxioReconciler) makeWorkerDeploymentForRoleGroup(instance *stackv1alpha1.Alluxio, roleGroupName string, roleGroup *stackv1alpha1.RoleGroupWorkerSpec, schema *runtime.Scheme) *appsv1.Deployment {
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
				Name: instance.GetNameWithSuffix("config"),
			},
		},
	})

	roleGroup = instance.Spec.Worker.GetRoleGroup(instance.Spec, roleGroupName)

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

	//var workerPorts []corev1.ContainerPort
	//var workerPortsValue reflect.Value
	//workerPortsValue = reflect.ValueOf(roleGroup.Ports)
	//// check is Ptr
	//if workerPortsValue.Kind() == reflect.Ptr && !workerPortsValue.IsNil() {
	//	workerPortsValue = workerPortsValue.Elem()
	//}
	//
	//for i := 0; i < workerPortsValue.NumField(); i++ {
	//	field := workerPortsValue.Field(i)
	//	workerPorts = append(workerPorts, corev1.ContainerPort{
	//		Name:     field.Type().Name(),
	//		HostPort: field.Interface().(int32),
	//	})
	//}

	//var jobWorkerPorts []corev1.ContainerPort
	//var jobWorkerPortsValue reflect.Value
	//var jobWrokerPortsType reflect.Type
	//jobWorkerPortsValue = reflect.ValueOf(roleGroup.Ports)
	//jobWrokerPortsType = reflect.TypeOf(roleGroup.Ports)
	//
	//// check is Ptr
	//if jobWorkerPortsValue.Kind() == reflect.Ptr && !jobWorkerPortsValue.IsNil() {
	//	jobWorkerPortsValue = jobWorkerPortsValue.Elem()
	//}
	//
	//for i := 0; i < jobWorkerPortsValue.NumField(); i++ {
	//	field := jobWorkerPortsValue.Field(i)
	//	jobWorkerPorts = append(jobWorkerPorts, corev1.ContainerPort{
	//		Name:     jobWrokerPortsType.Field(i).Name,
	//		HostPort: field.Interface().(int32),
	//	})
	//}

	volumes := makeTieredStoreVolumes(instance)
	volumes = append(makeShortCircuitVolumes(instance), volumes...)
	volumeMounts := makeTieredStoreVolumeMounts(instance)
	var needDomainSocketVolume bool

	if instance.Spec.ClusterConfig != nil && instance.Spec.ClusterConfig.ShortCircuit != nil && instance.Spec.ClusterConfig.ShortCircuit.Enabled {
		if instance.Spec.ClusterConfig.ShortCircuit.Policy == "uuid" {
			needDomainSocketVolume = true
		}
	}
	if needDomainSocketVolume {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "alluxio-domain",
			MountPath: "/opt/domain",
		})
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
		for _, roleGroup := range instance.Spec.Worker.RoleGroups {
			pvc := r.makeWorkerPVCForRoleGroup(instance, roleGroup, r.Scheme)
			if pvc != nil {
				pvcs = append(pvcs, pvc)
			}
		}
	}

	return pvcs, nil
}

func (r *AlluxioReconciler) makeWorkerPVCForRoleGroup(instance *stackv1alpha1.Alluxio, roleGroup *stackv1alpha1.RoleGroupWorkerSpec, schema *runtime.Scheme) *corev1.PersistentVolumeClaim {
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
			Name:      shortCircuit.PvcName,
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
			Selector: &metav1.LabelSelector{
				MatchLabels: mergedLabels,
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
