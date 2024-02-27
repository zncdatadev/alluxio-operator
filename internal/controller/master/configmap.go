package master

import (
	"fmt"
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/alluxio-operator/internal/common"
	"github.com/zncdata-labs/alluxio-operator/internal/role"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type ConfigMapReconciler struct {
	common.GeneralResourceStyleReconciler[*stackv1alpha1.Alluxio, *stackv1alpha1.MasterRoleGroupSpec]
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
		GeneralResourceStyleReconciler: *common.NewGeneraResourceStyleReconciler[*stackv1alpha1.Alluxio,
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

	var MasterCount int32
	var isSingleMaster, isHaEmbedded bool

	if masterRoleGroup != nil && masterRoleGroup.Replicas != 0 {
		MasterCount = masterRoleGroup.Replicas
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
		name := fmt.Sprintf("-Dalluxio.master.hostname=%s", instance.GetNameWithSuffix("master-"+groupName+"-0"))
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

	// ALLUXIO_MASTER_JAVA_OPTS
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

	// ALLUXIO_JOB_MASTER_JAVA_OPTS
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

	// ALLUXIO_WORKER_JAVA_OPTS
	workerJavaOpts := make([]string, 0)
	workerJavaOpts = append(workerJavaOpts, "-Dalluxio.worker.hostname=${ALLUXIO_WORKER_HOSTNAME}")

	workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.rpc.port=%d", workerRoleGroup.Config.Ports.Rpc))
	workerJavaOpts = append(workerJavaOpts, fmt.Sprintf("-Dalluxio.worker.web.port=%d", workerRoleGroup.Config.Ports.Web))

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

	// ALLUXIO_JOB_WORKER_JAVA_OPTS
	jobWorkerJavaOpts := make([]string, 0)

	jobWorkerJavaOpts = append(jobWorkerJavaOpts, "-Dalluxio.worker.hostname=${ALLUXIO_WORKER_HOSTNAME}")

	if workerRoleGroup.Config.JobWorker.Ports != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.rpc.port=%d", workerRoleGroup.Config.JobWorker.Ports.Rpc))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.data.port=%d", workerRoleGroup.Config.JobWorker.Ports.Data))
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-Dalluxio.job.worker.web.port=%d", workerRoleGroup.Config.JobWorker.Ports.Web))
	}

	if workerRoleGroup.Config.JobWorker.Properties != nil {
		for key, value := range workerRoleGroup.Config.JobWorker.Properties {
			jobWorkerJavaOpts = append(jobWorkerJavaOpts, fmt.Sprintf("-D%s=%s", key, value))
		}
	}

	if workerRoleGroup.Config.JobWorker.JvmOptions != nil {
		jobWorkerJavaOpts = append(jobWorkerJavaOpts, workerRoleGroup.Config.JobWorker.JvmOptions...)
	}

	cmData := map[string]string{
		"ALLUXIO_JAVA_OPTS":            strings.Join(alluxioJavaOpts, " "),
		"ALLUXIO_MASTER_JAVA_OPTS":     strings.Join(masterJavaOpts, " "),
		"ALLUXIO_JOB_MASTER_JAVA_OPTS": strings.Join(jobMasterJavaOpts, " "),
		"ALLUXIO_WORKER_JAVA_OPTS":     strings.Join(workerJavaOpts, " "),
		"ALLUXIO_JOB_WORKER_JAVA_OPTS": strings.Join(jobWorkerJavaOpts, " "),
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

func (s *ConfigMapReconciler) ConfigOverride(origin map[string]string) {
	cfg := s.MergedCfg
	overrideCfg := cfg.ConfigOverrides
	// if origin exists key of overrideCfg, then override it
	for key, value := range overrideCfg.OverrideConfig {
		origin[key] = value
	}
}
