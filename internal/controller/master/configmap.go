package master

import (
	"fmt"
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/alluxio-operator/internal/common"
	"github.com/zncdata-labs/alluxio-operator/internal/controller/role"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type ConfigMapReconciler struct {
	common.ConfigurationStyleReconciler[*stackv1alpha1.Alluxio, *stackv1alpha1.MasterRoleGroupSpec]
}

// NewConfigMap new a ConfigMapReconcile
func NewConfigMap(
	scheme *runtime.Scheme,
	instance *stackv1alpha1.Alluxio,
	client client.Client,
	mergedLabels map[string]string,
	mergedCfg *stackv1alpha1.MasterRoleGroupSpec,
) *ConfigMapReconciler {
	return &ConfigMapReconciler{
		ConfigurationStyleReconciler: *common.NewConfigurationStyleReconciler[*stackv1alpha1.Alluxio,
			*stackv1alpha1.MasterRoleGroupSpec](
			scheme,
			instance,
			client,
			mergedLabels,
			mergedCfg,
		),
	}
}

func (s *ConfigMapReconciler) Build(data common.ResourceBuilderData) (client.Object, error) {
	instance := s.Instance
	groupName := data.GroupName
	masterRoleGroup := s.MergedCfg
	workerCacheKey := createMasterGroupCacheKey(instance.GetName(), string(role.Worker), data.GroupName)
	workerRoleGroupObj, ok := common.MergedCache.Get(workerCacheKey)
	if !ok {
		return nil, fmt.Errorf("worker cache not found, key: %s", workerCacheKey)
	}
	workerRoleGroup := workerRoleGroupObj.(*stackv1alpha1.WorkerRoleGroupSpec)
	journal := instance.Spec.ClusterConfig.GetJournal()
	shortCircuit := instance.Spec.ClusterConfig.GetShortCircuit()

	//get master count, isSingleMaster, isHaEmbedded
	//create ALLUXIO_JAVA_OPTS, ALLUXIO_MASTER_JAVA_OPTS, ALLUXIO_JOB_MASTER_JAVA_OPTS, ALLUXIO_WORKER_JAVA_OPTS, ALLUXIO_JOB_WORKER_JAVA_OPTS
	masterCount := s.countMasterAmount(masterRoleGroup)
	isSingleMaster := s.isSingleMaster(masterCount)
	isHaEmbedded := s.isHaEmbedded(&journal, masterCount)
	alluxioJavaOpts := s.createAlluxioJavaOpts(instance, groupName, masterCount, isSingleMaster, isHaEmbedded, &journal)
	masterJavaOpts := s.createMasterJavaOpts(masterRoleGroup)
	jobMasterJavaOpts := s.createJobMasterJavaOpts(masterRoleGroup)
	workerJavaOpts := s.createWorkerJavaOpts(workerRoleGroup, &shortCircuit, instance)
	jobWorkerJavaOpts := s.createJobWorkerJavaOpts(workerRoleGroup)

	//create cmData
	cmData := map[string]string{
		"ALLUXIO_JAVA_OPTS":            alluxioJavaOpts,
		"ALLUXIO_MASTER_JAVA_OPTS":     masterJavaOpts,
		"ALLUXIO_JOB_MASTER_JAVA_OPTS": jobMasterJavaOpts,
		"ALLUXIO_WORKER_JAVA_OPTS":     workerJavaOpts,
		"ALLUXIO_JOB_WORKER_JAVA_OPTS": jobWorkerJavaOpts,
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.CreateMasterConfigMapName(instance.Name, groupName),
			Namespace: instance.Namespace,
			Labels:    instance.GetLabels(),
		},
		Data: cmData,
	}
	return cm, nil
}

func (s *ConfigMapReconciler) ConfigurationOverride(origin client.Object) {
	if cm, ok := origin.(*corev1.ConfigMap); ok {
		cfg := s.MergedCfg
		overrideCfg := cfg.ConfigOverrides
		if overrideCfg == nil {
			return
		}
		data := cm.Data
		// if origin exists key of overrideCfg, then override it
		for key, value := range overrideCfg.OverrideConfig {
			data[key] = value
		}
		cm.Data = data
	} else {
		panic("origin client.object is not ConfigMap")
	}
}

// count master amount
func (s *ConfigMapReconciler) countMasterAmount(masterRoleGroup *stackv1alpha1.MasterRoleGroupSpec) int32 {
	if masterRoleGroup != nil && masterRoleGroup.Replicas != 0 {
		return masterRoleGroup.Replicas
	}
	panic("masterRoleGroup cfg is nil")
}

// is single master
func (s *ConfigMapReconciler) isSingleMaster(masterCount int32) bool {
	return masterCount == 1
}

// is ha embedded
func (s *ConfigMapReconciler) isHaEmbedded(journal *stackv1alpha1.JournalSpec, masterCount int32) bool {
	return journal.Type == "EMBEDDED" && masterCount > 1
}

// create ALLUXIO_JAVA_OPTS
func (s *ConfigMapReconciler) createAlluxioJavaOpts(
	instance *stackv1alpha1.Alluxio,
	groupName string,
	MasterCount int32,
	isSingleMaster bool,
	isHaEmbedded bool,
	journal *stackv1alpha1.JournalSpec,
) string {
	// ALLUXIO_JAVA_OPTS
	alluxioJavaOpts := make([]string, 0)

	if isSingleMaster {
		name := fmt.Sprintf("-Dalluxio.master.hostname=%s", createMasterStatefulSetName(instance.Name, string(role.Master), groupName)+"-0")
		alluxioJavaOpts = append(alluxioJavaOpts, name)
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
			embeddedJournalAddresses += fmt.Sprintf("%s-master-%d:19200,", instance.GetNameWithSuffix(groupName), i)
		}
		alluxioJavaOpts = append(alluxioJavaOpts, embeddedJournalAddresses)
	}

	if instance.Spec.ClusterConfig.Properties != nil {
		for key, value := range instance.Spec.ClusterConfig.Properties {
			alluxioJavaOpts = append(alluxioJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if instance.Spec.ClusterConfig.JvmOptions != nil {
		alluxioJavaOpts = append(alluxioJavaOpts, instance.Spec.ClusterConfig.JvmOptions...)
	}
	return strings.Join(alluxioJavaOpts, " ")
}

// create ALLUXIO_MASTER_JAVA_OPTS
func (s *ConfigMapReconciler) createMasterJavaOpts(
	masterRoleGroup *stackv1alpha1.MasterRoleGroupSpec,
) string {
	masterJavaOpts := make([]string, 0)
	masterJavaOpts = append(masterJavaOpts, "-Dalluxio.master.hostname=${ALLUXIO_MASTER_HOSTNAME}")

	if masterRoleGroup != nil && masterRoleGroup.Config.Properties != nil {
		for key, value := range masterRoleGroup.Config.Properties {
			masterJavaOpts = append(masterJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if masterRoleGroup != nil && masterRoleGroup.Config.JvmOptions != nil {
		masterJavaOpts = append(masterJavaOpts, masterRoleGroup.Config.JvmOptions...)
	}
	return strings.Join(masterJavaOpts, " ")
}

// create ALLUXIO_JOB_MASTER_JAVA_OPTS
func (s *ConfigMapReconciler) createJobMasterJavaOpts(
	masterRoleGroup *stackv1alpha1.MasterRoleGroupSpec,
) string {
	jobMasterJavaOpts := make([]string, 0)
	jobMasterJavaOpts = append(jobMasterJavaOpts, "-Dalluxio.job.master.hostname=${ALLUXIO_JOB_MASTER_HOSTNAME}")

	if masterRoleGroup != nil && masterRoleGroup.Config.JobMaster.Properties != nil {
		for key, value := range masterRoleGroup.Config.JobMaster.Properties {
			jobMasterJavaOpts = append(jobMasterJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if masterRoleGroup != nil && masterRoleGroup.Config.JobMaster.JvmOptions != nil {
		jobMasterJavaOpts = append(jobMasterJavaOpts, masterRoleGroup.Config.JobMaster.JvmOptions...)
	}
	return strings.Join(jobMasterJavaOpts, " ")
}

// create ALLUXIO_WORKER_JAVA_OPTS
func (s *ConfigMapReconciler) createWorkerJavaOpts(
	workerRoleGroup *stackv1alpha1.WorkerRoleGroupSpec,
	shortCircuit *stackv1alpha1.ShortCircuitSpec,
	instance *stackv1alpha1.Alluxio,
) string {
	workerJavaOpts := make([]string, 0)
	workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.hostname=${ALLUXIO_WORKER_HOSTNAME}")

	workerPort := common.GetWorkerPorts(workerRoleGroup)

	workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.rpc.port=%d", workerPort.Rpc))
	workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.web.port=%d", workerPort.Web))

	if !shortCircuit.Enabled {
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.user.short.circuit.enabled=false")
	}

	if shortCircuit.Enabled || shortCircuit.Policy == "uuid" {
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.data.server.domain.socket.address=/opt/domain")
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.data.server.domain.socket.as.uuid=true")
	}

	if workerRoleGroup.Config.HostNetwork == nil || *workerRoleGroup.Config.HostNetwork {
		workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.container.hostname=${ALLUXIO_WORKER_CONTAINER_HOSTNAME}")
	}

	if workerRoleGroup.Config.Resources != nil && workerRoleGroup.Config.Resources.Memory != nil {
		workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.ramdisk.size=%s", workerRoleGroup.Config.Resources.Memory))
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

	if workerRoleGroup.Config.Properties != nil {
		for key, value := range workerRoleGroup.Config.Properties {
			workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if workerRoleGroup.Config.JvmOptions != nil {
		workerJavaOpts = append(workerJavaOpts, workerRoleGroup.Config.JvmOptions...)
	}
	return strings.Join(workerJavaOpts, " ")
}

// create ALLUXIO_JOB_WORKER_JAVA_OPTS
func (s *ConfigMapReconciler) createJobWorkerJavaOpts(
	workerRoleGroup *stackv1alpha1.WorkerRoleGroupSpec,
) string {
	jobWorkerJavaOpts := make([]string, 0)

	jobWorkerJavaOpts = append(jobWorkerJavaOpts, "-Dalluxio.worker.hostname=${ALLUXIO_WORKER_HOSTNAME}")
	jobWorkerPort := common.GetJobWorkerPorts(workerRoleGroup)
	if jobWorkerPort != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.rpc.port=%d", jobWorkerPort.Rpc))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.data.port=%d", jobWorkerPort.Data))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.web.port=%d", jobWorkerPort.Web))
	}

	if workerRoleGroup.Config.JobWorker.Properties != nil {
		for key, value := range workerRoleGroup.Config.JobWorker.Properties {
			jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if workerRoleGroup.Config.JobWorker.JvmOptions != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, workerRoleGroup.Config.JobWorker.JvmOptions...)
	}
	return strings.Join(jobWorkerJavaOpts, " ")
}
