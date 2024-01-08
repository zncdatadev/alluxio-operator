package controller

import (
	"context"
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	appV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *AlluxioReconciler) reconcileMasterService(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) error {
	logger := log.FromContext(ctx)
	svc := makeMasterService(ctx, instance, schema)
	if svc == nil {
		return nil
	}

	logger.Info("Creating/Updating master service")
	if err := CreateOrUpdate(ctx, r.Client, svc); err != nil {
		logger.Error(err, "Failed to create/update master service")
		return err
	}

	return nil
}

func (r *AlluxioReconciler) reconcileWorkerService(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) error {
	logger := log.FromContext(ctx)
	svc := makeWorkerService(ctx, instance, schema)
	if svc == nil {
		return nil
	}

	logger.Info("Creating/Updating worker service")
	if err := CreateOrUpdate(ctx, r.Client, svc); err != nil {
		logger.Error(err, "Failed to create/update worker service")
		return err
	}

	return nil
}

func makeMasterStatefulSet(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) *appV1.StatefulSet {
	logger := log.FromContext(ctx)

	var masterPorts []corev1.ContainerPort
	masterPortsValue := reflect.ValueOf(instance.Spec.Worker.Ports)
	masterPortsType := masterPortsValue.Type()
	for i := 0; i < masterPortsValue.NumField(); i++ {
		masterPorts = append(masterPorts, corev1.ContainerPort{
			Name:     masterPortsType.Field(i).Name,
			HostPort: masterPortsValue.Index(i).Interface().(int32),
		})
	}

	var jobMasterPorts []corev1.ContainerPort
	jobMasterPortsValue := reflect.ValueOf(instance.Spec.JobMaster.Ports)
	jobMasterPortsType := jobMasterPortsValue.Type()
	for i := 0; i < jobMasterPortsValue.NumField(); i++ {
		jobMasterPorts = append(jobMasterPorts, corev1.ContainerPort{
			Name:     jobMasterPortsType.Field(i).Name,
			HostPort: jobMasterPortsValue.Index(i).Interface().(int32),
		})
	}
	app := &appV1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + "-master",
			Namespace: instance.Namespace,
		},
		Spec: appV1.StatefulSetSpec{
			Replicas:    instance.Spec.Master.Replicas,
			ServiceName: instance.GetName() + "svc-master-headless",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": instance.GetName() + "-master",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": instance.GetName() + "-master",
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  instance.Spec.SecurityContext.RunAsUser,
						RunAsGroup: instance.Spec.SecurityContext.RunAsGroup,
					},
					Containers: []corev1.Container{
						{
							Name:    "master",
							Image:   instance.Spec.Image.Repository + ":" + instance.Spec.Image.Tag,
							Ports:   masterPorts,
							Command: []string{"tini", "--", "/entrypoint.sh"},
							Args:    instance.Spec.Master.Args,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.Master.Resources.Limits.CPU),
									"mem": k8sResource.MustParse(instance.Spec.Master.Resources.Limits.Memory),
								},
								Requests: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.Master.Resources.Requests.CPU),
									"mem": k8sResource.MustParse(instance.Spec.Master.Resources.Requests.Memory),
								},
							},
						},
						{
							Name:    "job-master",
							Image:   instance.Spec.Image.Repository + ":" + instance.Spec.Image.Tag,
							Ports:   jobMasterPorts,
							Command: []string{"tini", "--", "/entrypoint.sh"},
							Args:    instance.Spec.JobMaster.Args,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.JobMaster.Resources.Limits.CPU),
									"mem": k8sResource.MustParse(instance.Spec.JobMaster.Resources.Limits.Memory),
								},
								Requests: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.JobMaster.Resources.Requests.CPU),
									"mem": k8sResource.MustParse(instance.Spec.JobMaster.Resources.Requests.Memory),
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, app, schema); err != nil {
		logger.Error(err, "Failed to set owner for master statefulset")
		return nil
	}

	return app
}

func (r *AlluxioReconciler) reconcileMasterStatefulSet(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) error {
	logger := log.FromContext(ctx)
	app := makeMasterStatefulSet(ctx, instance, schema)
	if app == nil {
		return nil
	}

	logger.Info("Creating/Updating master statefulset")
	if err := CreateOrUpdate(ctx, r.Client, app); err != nil {
		logger.Error(err, "Failed to create/update master statefulset")
		return err
	}
	return nil
}

func makeWorkerDaemonSet(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) *appV1.DaemonSet {
	logger := log.FromContext(ctx)

	volumes := makeTieredStoreVolumes(instance)
	volumeMounts := makeTieredStoreVolumeMounts(instance)

	var workPorts []corev1.ContainerPort
	workerPortsValue := reflect.ValueOf(instance.Spec.Worker.Ports)
	workerPortsType := workerPortsValue.Type()
	for i := 0; i < workerPortsValue.NumField(); i++ {
		workPorts = append(workPorts, corev1.ContainerPort{
			Name:     workerPortsType.Field(i).Name,
			HostPort: workerPortsValue.Index(i).Interface().(int32),
		})
	}

	var jobWorkerPorts []corev1.ContainerPort
	jobWorkerPortsValue := reflect.ValueOf(instance.Spec.JobWorker.Ports)
	jobWorkerPortsType := jobWorkerPortsValue.Type()
	for i := 0; i < jobWorkerPortsValue.NumField(); i++ {
		jobWorkerPorts = append(jobWorkerPorts, corev1.ContainerPort{
			Name:     jobWorkerPortsType.Field(i).Name,
			HostPort: jobWorkerPortsValue.Index(i).Interface().(int32),
		})
	}

	app := &appV1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + "-worker",
			Namespace: instance.Namespace,
		},
		Spec: appV1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": instance.GetName() + "-worker",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": instance.GetName() + "-worker",
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  instance.Spec.SecurityContext.RunAsUser,
						RunAsGroup: instance.Spec.SecurityContext.RunAsGroup,
					},
					Containers: []corev1.Container{
						{
							Name:    "worker",
							Image:   instance.Spec.Image.Repository + ":" + instance.Spec.Image.Tag,
							Ports:   workPorts,
							Command: []string{"tini", "--", "/entrypoint.sh"},
							Args:    instance.Spec.Worker.Args,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.Worker.Resources.Limits.CPU),
									"mem": k8sResource.MustParse(instance.Spec.Worker.Resources.Limits.Memory),
								},
								Requests: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.Worker.Resources.Requests.CPU),
									"mem": k8sResource.MustParse(instance.Spec.Worker.Resources.Requests.Memory),
								},
							},
						},
						{
							Name:    "job-worker",
							Image:   instance.Spec.Image.Repository + ":" + instance.Spec.Image.Tag,
							Ports:   jobWorkerPorts,
							Command: []string{"tini", "--", "/entrypoint.sh"},
							Args:    instance.Spec.JobWorker.Args,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.JobWorker.Resources.Limits.CPU),
									"mem": k8sResource.MustParse(instance.Spec.JobWorker.Resources.Limits.Memory),
								},
								Requests: corev1.ResourceList{
									"cpu": k8sResource.MustParse(instance.Spec.JobWorker.Resources.Requests.CPU),
									"mem": k8sResource.MustParse(instance.Spec.JobWorker.Resources.Requests.Memory),
								},
							},
							VolumeMounts: volumeMounts,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, app, schema); err != nil {
		logger.Error(err, "Failed to set owner for worker daemonset")
		return nil
	}

	return app
}

func (r *AlluxioReconciler) reconcileWorkerDaemonSet(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) error {
	logger := log.FromContext(ctx)
	app := makeWorkerDaemonSet(ctx, instance, schema)
	if app == nil {
		return nil
	}

	logger.Info("Creating/Updating worker daemonset")
	if err := CreateOrUpdate(ctx, r.Client, app); err != nil {
		logger.Error(err, "Failed to create/update worker daemonset")
		return err
	}
	return nil
}
