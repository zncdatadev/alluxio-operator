package controller

//func (r *AlluxioReconciler) makeMasterStatefulSet(instance *stackv1alpha1.Alluxio) []*appsv1.StatefulSet {
//	var statefulSets []*appsv1.StatefulSet
//
//	if instance.Spec.Master.RoleGroups != nil {
//		for roleGroupName, _ := range instance.Spec.Master.RoleGroups {
//			roleGroup := instance.Spec.Master.GetRoleGroup(roleGroupName)
//			statefulSet := r.makeMasterStatefulSetForRoleGroup(instance, roleGroupName, roleGroup, r.Scheme)
//			if statefulSet != nil {
//				statefulSets = append(statefulSets, statefulSet)
//			}
//		}
//	}
//
//	return statefulSets
//}
//
//func (r *AlluxioReconciler) makeMasterStatefulSetForRoleGroup(instance *stackv1alpha1.Alluxio, roleGroupName string, roleGroup *stackv1alpha1.RoleGroupMasterSpec, schema *runtime.Scheme) *appsv1.StatefulSet {
//	labels := instance.GetLabels()
//
//	additionalLabels := make(map[string]string)
//
//	if roleGroup != nil && roleGroup.MatchLabels != nil {
//		for k, v := range roleGroup.MatchLabels {
//			additionalLabels[k] = v
//		}
//	}
//
//	mergedLabels := make(map[string]string)
//	for key, value := range labels {
//		mergedLabels[key] = value
//	}
//	for key, value := range additionalLabels {
//		mergedLabels[key] = value
//	}
//
//	//roleConfig := instance.Spec.Master.RoleConfig
//	//jobMaster := roleConfig.JobMaster
//	//var image stackv1alpha1.ImageSpec
//	//var podSecurityContext *corev1.PodSecurityContext
//	//var args []string
//	//var jobArgs []string
//	//var jobResources corev1.ResourceRequirements
//	//var hostPID bool
//	//var hostNetwork bool
//	//var dnsPolicy corev1.DNSPolicy
//	//var shareProcessNamespace bool
//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.HostPID {
//	//	hostPID = roleGroup.Config.HostPID
//	//} else if roleConfig.HostPID {
//	//	hostPID = roleConfig.HostPID
//	//}
//	//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.HostNetwork {
//	//	hostNetwork = roleGroup.Config.HostNetwork
//	//} else if roleConfig.HostNetwork {
//	//	hostNetwork = roleConfig.HostNetwork
//	//}
//	//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.DnsPolicy != "" {
//	//	dnsPolicy = roleGroup.Config.DnsPolicy
//	//} else if roleConfig.DnsPolicy != "" {
//	//	dnsPolicy = roleConfig.DnsPolicy
//	//} else {
//	//	if hostNetwork {
//	//		dnsPolicy = corev1.DNSClusterFirstWithHostNet
//	//	} else {
//	//		dnsPolicy = corev1.DNSClusterFirst
//	//	}
//	//}
//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.ShareProcessNamespace {
//	//	shareProcessNamespace = roleGroup.Config.ShareProcessNamespace
//	//} else if roleConfig.ShareProcessNamespace {
//	//	shareProcessNamespace = roleConfig.ShareProcessNamespace
//	//}
//	//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.Image != nil {
//	//	image = *roleGroup.Config.Image
//	//} else {
//	//	image = *instance.Spec.Image
//	//}
//	//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.SecurityContext != nil {
//	//	securityContext := roleGroup.Config.SecurityContext
//	//	podSecurityContext = &corev1.PodSecurityContext{
//	//		RunAsUser:  securityContext.RunAsUser,
//	//		RunAsGroup: securityContext.RunAsGroup,
//	//		FSGroup:    securityContext.FSGroup,
//	//	}
//	//} else if instance.Spec.SecurityContext != nil {
//	//	securityContext := instance.Spec.SecurityContext
//	//	podSecurityContext = &corev1.PodSecurityContext{
//	//		RunAsUser:  securityContext.RunAsUser,
//	//		RunAsGroup: securityContext.RunAsGroup,
//	//		FSGroup:    securityContext.FSGroup,
//	//	}
//	//}
//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.Args != nil {
//	//	args = roleGroup.Config.Args
//	//} else {
//	//	args = instance.Spec.Master.Args
//	//}
//	//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.JobMaster != nil && roleGroup.Config.JobMaster.Args != nil {
//	//	jobArgs = roleGroup.Config.JobMaster.Args
//	//} else if roleConfig != nil && jobMaster != nil && jobMaster.Args != nil {
//	//	jobArgs = jobMaster.Args
//	//}
//	//
//	//if roleGroup != nil && roleGroup.Config != nil && roleGroup.Config.JobMaster != nil && roleGroup.Config.JobMaster.Resources != nil {
//	//	jobResources = *roleGroup.Config.JobMaster.Resources
//	//} else if jobMaster.Resources != nil {
//	//	jobResources = *jobMaster.Resources
//	//}
//
//	//var masterPorts []corev1.ContainerPort
//	//var masterPortsValue reflect.Value
//	////if roleGroup.Config.Ports != nil {
//	////	masterPortsValue = reflect.ValueOf(roleGroup.Config.Ports)
//	////} else if instance.GetMasterPorts() != nil {
//	////	masterPortsValue = reflect.ValueOf(instance.GetMasterPorts())
//	////}
//	////
//	//masterPortsValue = reflect.ValueOf(instance.GetMasterPorts())
//	//// check is Ptr
//	//if masterPortsValue.Kind() == reflect.Ptr && !masterPortsValue.IsNil() {
//	//	masterPortsValue = masterPortsValue.Elem()
//	//}
//	//
//	//for i := 0; i < masterPortsValue.NumField(); i++ {
//	//	field := masterPortsValue.Field(i)
//	//	masterPorts = append(masterPorts, corev1.ContainerPort{
//	//		Name:     field.Type().Name(),
//	//		HostPort: field.Interface().(int32),
//	//	})
//	//}
//
//	//var jobMasterPorts []corev1.ContainerPort
//	//var jobMasterPortsValue reflect.Value
//	//if roleGroup.Config.JobMaster.Ports != nil {
//	//	jobMasterPortsValue = reflect.ValueOf(roleGroup.Config.JobMaster.Ports)
//	//} else if instance.Spec.Master.RoleConfig.JobMaster.Ports != nil {
//	//	jobMasterPortsValue = reflect.ValueOf(jobMaster.Ports)
//	//}
//	//
//	//jobMasterPortsValue = reflect.ValueOf(instance.GetJobMasterPorts())
//	//
//	//if jobMasterPortsValue.Kind() == reflect.Ptr && !jobMasterPortsValue.IsNil() {
//	//	jobMasterPortsValue = jobMasterPortsValue.Elem()
//	//}
//	//
//	//for i := 0; i < jobMasterPortsValue.NumField(); i++ {
//	//	field := jobMasterPortsValue.Field(i)
//	//	jobMasterPorts = append(jobMasterPorts, corev1.ContainerPort{
//	//		Name:     field.Type().Name(),
//	//		HostPort: field.Interface().(int32),
//	//	})
//	//}
//
//	var isUfsLocal, isEmbedded, needJournalVolume bool
//	var volumeMounts []corev1.VolumeMount
//	var volumes []corev1.Volume
//	var volumeClaimTemplates []corev1.PersistentVolumeClaim
//
//	if instance.Spec.ClusterConfig != nil && instance.Spec.ClusterConfig.Journal != nil {
//		// Set isUfsLocal and isEmbedded based on Journal's type and UfsType
//		if instance.Spec.ClusterConfig.Journal.Type == "UFS" && instance.Spec.ClusterConfig.Journal.UfsType == "local" {
//			isUfsLocal = true
//		}
//		if instance.Spec.ClusterConfig.Journal.Type == "EMBEDDED" {
//			isEmbedded = true
//		}
//
//		// Set needJournalVolume if isUfsLocal or isEmbedded is true
//		if isUfsLocal || isEmbedded {
//			needJournalVolume = true
//		}
//
//		// Create and add volumeMount if needJournalVolume is true
//		if needJournalVolume {
//			volumeMount := corev1.VolumeMount{
//				Name:      "alluxio-journal",
//				MountPath: instance.Spec.ClusterConfig.Journal.Folder,
//			}
//			volumeMounts = append(volumeMounts, volumeMount)
//		}
//
//		// Create and add volume if needJournalVolume is true and VolumeType is "emptyDir"
//		if needJournalVolume && instance.Spec.ClusterConfig.Journal.VolumeType == "emptyDir" {
//			volume := corev1.Volume{
//				Name: "alluxio-journal",
//				VolumeSource: corev1.VolumeSource{
//					EmptyDir: &corev1.EmptyDirVolumeSource{},
//				},
//			}
//			volumes = append(volumes, volume)
//		}
//
//		// Create and add PersistentVolumeClaim if needJournalVolume is true and VolumeType is "persistentVolumeClaim"
//		if needJournalVolume && instance.Spec.ClusterConfig.Journal.VolumeType == "persistentVolumeClaim" {
//			accessMode := corev1.PersistentVolumeAccessMode(instance.Spec.ClusterConfig.Journal.AccessMode)
//			pvc := corev1.PersistentVolumeClaim{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "alluxio-journal",
//				},
//				Spec: corev1.PersistentVolumeClaimSpec{
//					StorageClassName: &instance.Spec.ClusterConfig.Journal.StorageClass,
//					AccessModes:      []corev1.PersistentVolumeAccessMode{accessMode},
//					Resources: corev1.VolumeResourceRequirements{
//						Requests: corev1.ResourceList{
//							corev1.ResourceStorage: resource.MustParse(instance.Spec.ClusterConfig.Journal.Size),
//						},
//					},
//				},
//			}
//			volumeClaimTemplates = append(volumeClaimTemplates, pvc)
//		}
//	}
//
//	roleGroup := instance.Spec.Master.GetRoleGroup(roleGroupName)
//
//	app := &appsv1.StatefulSet{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      instance.GetNameWithSuffix("master-" + roleGroupName),
//			Namespace: instance.Namespace,
//		},
//		Spec: appsv1.StatefulSetSpec{
//			Replicas:    &roleGroup.Replicas,
//			ServiceName: instance.GetName() + "svc-master-headless",
//			Selector: &metav1.LabelSelector{
//				MatchLabels: mergedLabels,
//			},
//			Template: corev1.PodTemplateSpec{
//				ObjectMeta: metav1.ObjectMeta{
//					Labels: mergedLabels,
//				},
//				Spec: corev1.PodSpec{
//					SecurityContext:       roleGroup.SecurityContext,
//					HostPID:               group.Config.HostPID,
//					HostNetwork:           hostNetwork,
//					DNSPolicy:             dnsPolicy,
//					ShareProcessNamespace: &shareProcessNamespace,
//					Containers: []corev1.Container{
//						{
//							Name:            instance.GetNameWithSuffix("master"),
//							Image:           image.Repository + ":" + image.Tag,
//							ImagePullPolicy: image.PullPolicy,
//							Ports:           masterPorts,
//							Command:         []string{"tini", "--", "/entrypoint.sh"},
//							Args:            args,
//							Resources:       *roleGroup.Config.Resources,
//							VolumeMounts:    volumeMounts,
//						},
//						{
//							Name:            instance.GetNameWithSuffix("job-master"),
//							Image:           image.Repository + ":" + image.Tag,
//							ImagePullPolicy: image.PullPolicy,
//							Ports:           jobMasterPorts,
//							Command:         []string{"tini", "--", "/entrypoint.sh"},
//							Args:            jobArgs,
//							Resources:       jobResources,
//						},
//					},
//					Volumes: volumes,
//				},
//			},
//			VolumeClaimTemplates: volumeClaimTemplates,
//		},
//	}
//
//	err := controllerutil.SetControllerReference(instance, app, schema)
//	if err != nil {
//		r.Log.Error(err, "Failed to set controller reference for master statefulset")
//		return nil
//	}
//
//	return app
//}
