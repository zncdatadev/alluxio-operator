package controller

import (
	"context"
	"fmt"
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/operator-go/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

func (r *AlluxioReconciler) reconcileMasterStatefulSet(ctx context.Context, instance *stackv1alpha1.Alluxio) error {

	masterStatefulSet := r.makeMasterStatefulSet(instance)
	for _, app := range masterStatefulSet {
		if app == nil {
			continue
		}

		if err := CreateOrUpdate(ctx, r.Client, app); err != nil {
			r.Log.Error(err, "Failed to create or update master Statefulset", "statefulset", app.Name)
			return err
		}
	}
	return nil
}

func (r *AlluxioReconciler) reconcileWorkerDeployment(ctx context.Context, instance *stackv1alpha1.Alluxio) error {
	workerDeployment := r.makeWorkerDeployment(instance)
	for _, dep := range workerDeployment {
		if dep == nil {
			continue
		}

		if err := CreateOrUpdate(ctx, r.Client, dep); err != nil {
			r.Log.Error(err, "Failed to create or update master Deployment", "deployment", dep.Name)
			return err
		}
	}
	return nil
}

func (r *AlluxioReconciler) reconcilePVC(ctx context.Context, instance *stackv1alpha1.Alluxio) error {
	obj, err := r.makeWorkerPVCs(instance)
	if err != nil {
		return err
	}

	for _, pvc := range obj {
		if err := CreateOrUpdate(ctx, r.Client, pvc); err != nil {
			r.Log.Error(err, "Failed to create or update PVC")
			return err
		}
	}

	return nil
}

func (r *AlluxioReconciler) makeServices(instance *stackv1alpha1.Alluxio) ([]*corev1.Service, error) {
	var services []*corev1.Service

	if instance.Spec.Master.RoleGroups != nil {
		for roleGroupName, roleGroup := range instance.Spec.Master.RoleGroups {
			svc, err := r.makeMasterServiceForRoleGroup(instance, roleGroupName, roleGroup, r.Scheme)
			if err != nil {
				return nil, err
			}
			services = append(services, svc)
		}
	}

	return services, nil
}

func (r *AlluxioReconciler) makeMasterServiceForRoleGroup(instance *stackv1alpha1.Alluxio, roleGroupName string, roleGroup *stackv1alpha1.RoleMasterSpec, schema *runtime.Scheme) (*corev1.Service, error) {
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

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix("master-" + roleGroupName + "-0"),
			Namespace: instance.Namespace,
			Labels:    mergedLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "rpc",
					Port: roleGroup.Ports.Rpc,
				},
				{
					Name: "web",
					Port: roleGroup.Ports.Web,
				},
				{
					Name: "embedded",
					Port: roleGroup.Ports.Embedded,
				},
				{
					Name: "job-rpc",
					Port: roleGroup.JobMaster.Ports.Rpc,
				},
				{
					Name: "job-web",
					Port: roleGroup.JobMaster.Ports.Web,
				},
				{
					Name: "job-embedded",
					Port: roleGroup.JobMaster.Ports.Embedded,
				},
			},
			Selector:  mergedLabels,
			ClusterIP: "None",
		},
	}
	err := ctrl.SetControllerReference(instance, svc, schema)
	if err != nil {
		r.Log.Error(err, "Failed to set controller reference for service")
		return nil, errors.Wrap(err, "Failed to set controller reference for service")
	}
	return svc, nil
}

func (r *AlluxioReconciler) reconcileService(ctx context.Context, instance *stackv1alpha1.Alluxio) error {
	services, err := r.makeServices(instance)
	if err != nil {
		return err
	}

	for _, svc := range services {
		if svc == nil {
			continue
		}

		if err := CreateOrUpdate(ctx, r.Client, svc); err != nil {
			r.Log.Error(err, "Failed to create or update service", "service", svc.Name)
			return err
		}
	}

	return nil
}

func (r *AlluxioReconciler) makeConfigMapForRoleGroup(instance *stackv1alpha1.Alluxio, mgn, wgn string, masterRoleGroup *stackv1alpha1.RoleMasterSpec, workerRoleGroup *stackv1alpha1.RoleWorkerSpec) (*corev1.ConfigMap, error) {

	//split := strings.Split(roleGroupName, "-")
	//mgn := split[0]
	//wgn := split[1]
	masterRoleGroup = instance.Spec.Master.GetRoleGroup(instance, mgn)
	workerRoleGroup = instance.Spec.Worker.GetRoleGroup(instance, wgn)
	journal := instance.Spec.ClusterConfig.GetJournal()
	shortCircuit := instance.Spec.ClusterConfig.GetShortCircuit()

	var MasterCount int32
	var isSingleMaster, isHaEmbedded bool

	if masterRoleGroup != nil && masterRoleGroup.Replicas != nil {
		MasterCount = *masterRoleGroup.Replicas
	}

	isSingleMaster = MasterCount == 1

	if journal.Type == "EMBEDDED" && MasterCount > 1 {
		isHaEmbedded = true
	} else {
		isHaEmbedded = false
	}

	// ALLUXIO_JAVA_OPTS
	alluxioJavaOpts := make([]string, 0)

	if isSingleMaster {
		alluxioJavaOpts = append(alluxioJavaOpts, fmt.Sprintf("-Dalluxio.master.hostname=%s", instance.GetNameWithSuffix("master-"+mgn+"-0")))
	}

	if journal.Type != "" {
		alluxioJavaOpts = append(alluxioJavaOpts, fmt.Sprintf("-Dalluxio.master.journal.type=%v", journal.Type))
	}
	if journal.Folder != "" {
		alluxioJavaOpts = append(alluxioJavaOpts, fmt.Sprintf("-Dalluxio.master.journal.folder=%v", journal.Folder))
	}

	if isHaEmbedded {
		embeddedJournalAddresses := "-Dalluxio.master.embedded.journal.addresses="
		for i := 0; i < int(MasterCount); i++ {
			embeddedJournalAddresses += fmt.Sprintf("%s-master-%d:19200,", instance.GetNameWithSuffix(mgn), i)
		}
		alluxioJavaOpts = append(alluxioJavaOpts, embeddedJournalAddresses)
	}

	if instance.Spec.ClusterConfig.Properties != nil {
		for key, value := range instance.Spec.ClusterConfig.Properties {
			alluxioJavaOpts = append(alluxioJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if instance.Spec.ClusterConfig.JvmOptions != nil {
		for _, jvmOption := range instance.Spec.ClusterConfig.JvmOptions {
			alluxioJavaOpts = append(alluxioJavaOpts, jvmOption)
		}
	}

	// ALLUXIO_MASTER_JAVA_OPTS
	masterJavaOpts := make([]string, 0)
	masterJavaOpts = append(masterJavaOpts, "-Dalluxio.master.hostname=${ALLUXIO_MASTER_HOSTNAME}")

	if masterRoleGroup != nil && masterRoleGroup.Properties != nil {
		for key, value := range masterRoleGroup.Properties {
			masterJavaOpts = append(masterJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	} else if instance.Spec.Master.RoleConfig.Properties != nil {
		for key, value := range instance.Spec.Master.RoleConfig.Properties {
			masterJavaOpts = append(masterJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if masterRoleGroup != nil && masterRoleGroup.JvmOptions != nil {
		for _, jvmOption := range masterRoleGroup.JvmOptions {
			masterJavaOpts = append(masterJavaOpts, jvmOption)
		}
	} else if instance.Spec.Master.RoleConfig.JvmOptions != nil {
		for _, jvmOption := range instance.Spec.Master.RoleConfig.JvmOptions {
			masterJavaOpts = append(masterJavaOpts, jvmOption)
		}
	}

	// ALLUXIO_JOB_MASTER_JAVA_OPTS
	jobMasterJavaOpts := make([]string, 0)
	jobMasterJavaOpts = append(jobMasterJavaOpts, "-Dalluxio.job.master.hostname=${ALLUXIO_JOB_MASTER_HOSTNAME}")

	if masterRoleGroup != nil && masterRoleGroup.JobMaster.Properties != nil {
		for key, value := range masterRoleGroup.JobMaster.Properties {
			jobMasterJavaOpts = append(jobMasterJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	} else if instance.Spec.Master.RoleConfig.JobMaster.Properties != nil {
		for key, value := range instance.Spec.Master.RoleConfig.JobMaster.Properties {
			jobMasterJavaOpts = append(jobMasterJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if masterRoleGroup != nil && masterRoleGroup.JobMaster.JvmOptions != nil {
		for _, jvmOption := range masterRoleGroup.JobMaster.JvmOptions {
			jobMasterJavaOpts = append(jobMasterJavaOpts, jvmOption)
		}
	} else if instance.Spec.Master.RoleConfig.JobMaster.JvmOptions != nil {
		for _, jvmOption := range instance.Spec.Master.RoleConfig.JobMaster.JvmOptions {
			jobMasterJavaOpts = append(jobMasterJavaOpts, jvmOption)
		}
	}

	// ALLUXIO_WORKER_JAVA_OPTS
	workerJavaOpts := make([]string, 0)
	workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.hostname=${ALLUXIO_WORKER_HOSTNAME}")

	workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.rpc.port=%d", workerRoleGroup.Ports.Rpc))
	workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.web.port=%d", workerRoleGroup.Ports.Web))

	if workerRoleGroup.Ports.Rpc == 0 {
		workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.rpc.port=%d", instance.Spec.Worker.RoleConfig.Ports.Rpc))
	}

	if workerRoleGroup.Ports.Web == 0 {
		workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.web.port=%d", instance.Spec.Worker.RoleConfig.Ports.Web))
	}

	if !shortCircuit.Enabled {
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.user.short.circuit.enabled=false")
	}

	if shortCircuit.Enabled || shortCircuit.Policy == "uuid" {
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.data.server.domain.socket.address=/opt/domain")
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.data.server.domain.socket.as.uuid=true")
	}

	if workerRoleGroup.HostNetwork == nil || *workerRoleGroup.HostNetwork || *instance.Spec.Worker.RoleConfig.HostNetwork {
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.container.hostname=${ALLUXIO_WORKER_CONTAINER_HOSTNAME}")
	}

	if workerRoleGroup.Resources != nil && workerRoleGroup.Resources.Requests != nil && workerRoleGroup.Resources.Requests.Memory() != nil {
		workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.ramdisk.size=%s", workerRoleGroup.Resources.Requests.Memory().String()))
	} else if instance.Spec.Worker.RoleConfig.Resources != nil && instance.Spec.Worker.RoleConfig.Resources.Requests != nil && instance.Spec.Worker.RoleConfig.Resources.Requests.Memory() != nil {
		workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.ramdisk.size=%s", instance.Spec.Worker.RoleConfig.Resources.Requests.Memory().String()))
	}

	if instance.Spec.ClusterConfig.TieredStore != nil {
		workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.tieredstore.levels=%d", len(instance.Spec.ClusterConfig.TieredStore)))

		for _, tier := range instance.Spec.ClusterConfig.TieredStore {
			tierName := fmt.Sprintf("-Dalluxio.worker.tieredstore.level%d", tier.Level)

			if tier.Alias != "" {
				workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("%s.alias=%s", tierName, tier.Alias))
			}

			workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("%s.mediumtype=%s", tierName, tier.MediumType))

			if tier.Path != "" {
				workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("%s.dirs.path=%s", tierName, tier.Path))
			}

			if tier.Quota != "" {
				workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("%s.dirs.quota=%s", tierName, tier.Quota))
			}

			workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("%s.watermark.high.ratio=%.2f", tierName, tier.High))

			workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("%s.watermark.low.ratio=%.2f", tierName, tier.Low))
		}
	}

	if workerRoleGroup.Properties != nil {
		for key, value := range workerRoleGroup.Properties {
			workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	} else if instance.Spec.Worker.RoleConfig.Properties != nil {
		for key, value := range instance.Spec.Worker.RoleConfig.Properties {
			workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if workerRoleGroup.JvmOptions != nil {
		workerJavaOpts = append(workerJavaOpts, workerRoleGroup.JvmOptions...)
	} else if instance.Spec.Worker.RoleConfig.JvmOptions != nil {
		workerJavaOpts = append(workerJavaOpts, instance.Spec.Worker.RoleConfig.JvmOptions...)
	}

	// ALLUXIO_JOB_WORKER_JAVA_OPTS
	jobWorkerJavaOpts := make([]string, 0)

	jobWorkerJavaOpts = append(jobWorkerJavaOpts, "-Dalluxio.worker.hostname=${ALLUXIO_WORKER_HOSTNAME}")

	if workerRoleGroup.JobWorker.Ports != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.rpc.port=%d", workerRoleGroup.JobWorker.Ports.Rpc))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.data.port=%d", workerRoleGroup.JobWorker.Ports.Data))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.web.port=%d", workerRoleGroup.JobWorker.Ports.Web))
	} else if instance.Spec.Worker.RoleConfig.JobWorker.Ports != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.rpc.port=%d", instance.Spec.Worker.RoleConfig.JobWorker.Ports.Rpc))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.data.port=%d", instance.Spec.Worker.RoleConfig.JobWorker.Ports.Data))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.web.port=%d", instance.Spec.Worker.RoleConfig.JobWorker.Ports.Web))
	}

	if workerRoleGroup.JobWorker.Properties != nil {
		for key, value := range workerRoleGroup.JobWorker.Properties {
			jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	} else if instance.Spec.Worker.RoleConfig.JobWorker.Properties != nil {
		for key, value := range instance.Spec.Worker.RoleConfig.JobWorker.Properties {
			jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if workerRoleGroup.JobWorker.JvmOptions != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, workerRoleGroup.JobWorker.JvmOptions...)
	} else if instance.Spec.Worker.RoleConfig.JobWorker.JvmOptions != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, instance.Spec.Worker.RoleConfig.JobWorker.JvmOptions...)
	}

	data := map[string]string{
		"ALLUXIO_JAVA_OPTS":            strings.Join(alluxioJavaOpts, " "),
		"ALLUXIO_MASTER_JAVA_OPTS":     strings.Join(masterJavaOpts, " "),
		"ALLUXIO_JOB_MASTER_JAVA_OPTS": strings.Join(jobMasterJavaOpts, " "),
		"ALLUXIO_WORKER_JAVA_OPTS":     strings.Join(workerJavaOpts, " "),
		"ALLUXIO_JOB_WORKER_JAVA_OPTS": strings.Join(jobWorkerJavaOpts, " "),
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetNameWithSuffix("config" + "-" + mgn),
			Namespace: instance.Namespace,
			Labels:    instance.GetLabels(),
		},
		Data: data,
	}

	if err := ctrl.SetControllerReference(instance, cm, r.Scheme); err != nil {
		return nil, err
	}

	return cm, nil
}

func (r *AlluxioReconciler) makeConfigMap(instance *stackv1alpha1.Alluxio) []*corev1.ConfigMap {
	var configMaps []*corev1.ConfigMap

	if instance.Spec.Master.RoleGroups != nil {
		for masterRoleGroupName, masterRoleGroup := range instance.Spec.Master.RoleGroups {
			if instance.Spec.Worker.RoleConfig != nil {
				for workerRoleGroupName, workerRoleGroup := range instance.Spec.Worker.RoleGroups {
					configMap, err := r.makeConfigMapForRoleGroup(instance, masterRoleGroupName, workerRoleGroupName, masterRoleGroup, workerRoleGroup)
					if err != nil {
						r.Log.Error(err, "Failed to create configMap for role group", "roleGroupName", masterRoleGroupName+"-"+workerRoleGroupName)
						continue
					}
					configMaps = append(configMaps, configMap)
				}
			}

		}
	}
	return configMaps
}

func (r *AlluxioReconciler) reconcileConfigMap(ctx context.Context, instance *stackv1alpha1.Alluxio) error {

	configMap := r.makeConfigMap(instance)
	for _, cm := range configMap {
		if cm == nil {
			continue
		}

		if err := CreateOrUpdate(ctx, r.Client, cm); err != nil {
			r.Log.Error(err, "Failed to create or update configMap", "configMap", cm.Name)
			return err
		}
	}
	return nil
}

func emptyDirVolumeSource(quota string) corev1.VolumeSource {
	sizeLimit := k8sResource.MustParse(quota)
	return corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{
			Medium:    "Memory",
			SizeLimit: &sizeLimit,
		},
	}
}

func hostPathVolumeSource(path string) corev1.VolumeSource {
	hostPathType := corev1.HostPathType("DirectoryOrCreate")
	return corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
			Type: &hostPathType,
		},
	}
}

func pvcVolumeSource(claimName string) corev1.VolumeSource {
	return corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: claimName,
		},
	}
}
