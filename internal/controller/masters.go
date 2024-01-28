package controller

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *AlluxioReconciler) makeMasterStatefulSet(instance *stackv1alpha1.Alluxio) []*appsv1.StatefulSet {
	var statefulSets []*appsv1.StatefulSet

	if instance.Spec.Master.RoleGroups != nil {
		for roleGroupName, roleGroup := range instance.Spec.Master.RoleGroups {
			statefulSet := r.makeMasterStatefulSetForRoleGroup(instance, roleGroupName, roleGroup, r.Scheme)
			if statefulSet != nil {
				statefulSets = append(statefulSets, statefulSet)
			}
		}
	}

	return statefulSets
}

func (r *AlluxioReconciler) makeMasterStatefulSetForRoleGroup(instance *stackv1alpha1.Alluxio, roleGroupName string, roleGroup *stackv1alpha1.RoleMasterSpec, schema *runtime.Scheme) *appsv1.StatefulSet {
	labels := instance.GetLabels()

	additionalLabels := make(map[string]string)
	roleGroup = instance.Spec.Master.GetRoleGroup(instance, roleGroupName)

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

	var isUfsLocal, isEmbedded, needJournalVolume, isHaEmbedded, isSingleMaster bool
	var volumeMounts []corev1.VolumeMount
	var volumes []corev1.Volume
	var volumeClaimTemplates []corev1.PersistentVolumeClaim
	var envVarsMap = make(map[string]string)
	var envVars []corev1.EnvVar
	var envFrom []corev1.EnvFromSource

	// Set isUfsLocal and isEmbedded based on Journal's type and UfsType
	if instance.Spec.ClusterConfig.GetJournal().Type == "UFS" && instance.Spec.ClusterConfig.GetJournal().UfsType == "local" {
		isUfsLocal = true
	}
	if instance.Spec.ClusterConfig.GetJournal().Type == "EMBEDDED" {
		isEmbedded = true
	}

	if *roleGroup.Replicas == 1 {
		isSingleMaster = true
	}

	// Set needJournalVolume if isUfsLocal or isEmbedded is true
	if isUfsLocal || isEmbedded {
		needJournalVolume = true
	}

	if isEmbedded && *roleGroup.Replicas > 1 {
		isHaEmbedded = true
	}

	journal := instance.Spec.ClusterConfig.GetJournal()
	// Create and add volumeMount if needJournalVolume is true
	if needJournalVolume {
		volumeMount := corev1.VolumeMount{
			Name:      "alluxio-journal",
			MountPath: journal.Folder,
		}
		volumeMounts = append(volumeMounts, volumeMount)
	}

	// Create and add volume if needJournalVolume is true and VolumeType is "emptyDir"
	if needJournalVolume && journal.VolumeType == "emptyDir" {
		volume := corev1.Volume{
			Name: "alluxio-journal",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		volumes = append(volumes, volume)
	}

	// Create and add PersistentVolumeClaim if needJournalVolume is true and VolumeType is "persistentVolumeClaim"
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

	envFrom = append(envFrom, corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: instance.GetNameWithSuffix("config" + "-" + roleGroupName),
			},
		},
	})

	if roleGroup.EnvVars != nil {
		for key, value := range roleGroup.EnvVars {
			envVarsMap[key] = value
		}

		for key, value := range envVarsMap {
			envVars = append(envVars, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}

	masterPorts := []corev1.ContainerPort{
		{
			Name:          "web",
			ContainerPort: roleGroup.Ports.Web,
		},
		{
			Name:          "rpc",
			ContainerPort: roleGroup.Ports.Rpc,
		},
	}

	jobMasterPorts := []corev1.ContainerPort{
		{
			Name:          "job-web",
			ContainerPort: roleGroup.JobMaster.Ports.Web,
		},
		{
			Name:          "job-rpc",
			ContainerPort: roleGroup.JobMaster.Ports.Rpc,
		},
	}

	if isHaEmbedded {
		masterPorts = append(masterPorts, corev1.ContainerPort{
			Name:          "embedded",
			ContainerPort: roleGroup.Ports.Embedded,
		})
		jobMasterPorts = append(jobMasterPorts, corev1.ContainerPort{
			Name:          "job-embedded",
			ContainerPort: roleGroup.JobMaster.Ports.Embedded,
		})
	}

	app := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix("master-" + roleGroupName),
			Namespace: instance.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    roleGroup.Replicas,
			ServiceName: instance.GetName() + "svc-master-headless",
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
							Name:            instance.GetNameWithSuffix("master"),
							Image:           roleGroup.Image.Repository + ":" + roleGroup.Image.Tag,
							ImagePullPolicy: roleGroup.Image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports:           masterPorts,
							Command:         []string{"tini", "--", "/entrypoint.sh"},
							Args:            roleGroup.Args,
							Resources:       *roleGroup.Resources,
							VolumeMounts:    volumeMounts,
						},
						{
							Name:            instance.GetNameWithSuffix("job-master"),
							Image:           roleGroup.Image.Repository + ":" + roleGroup.Image.Tag,
							ImagePullPolicy: roleGroup.Image.PullPolicy,
							Env:             envVars,
							EnvFrom:         envFrom,
							Ports:           jobMasterPorts,
							Command:         []string{"tini", "--", "/entrypoint.sh"},
							Args:            roleGroup.JobMaster.Args,
							Resources:       *roleGroup.JobMaster.Resources,
						},
					},
					Volumes: volumes,
				},
			},
			VolumeClaimTemplates: volumeClaimTemplates,
		},
	}

	err := controllerutil.SetControllerReference(instance, app, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for master statefulset")
		return nil
	}

	return app
}
