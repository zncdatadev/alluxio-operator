package worker

import (
	"strings"

	alluxiov1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func VolumeSourceFromTieredStore(tieredStore alluxiov1alpha1.TieredStore, mediumType string, index int) corev1.Volume {
	volumeName := strings.ToLower(mediumType)
	var volumeSource corev1.VolumeSource

	switch tieredStore.Type {
	case "hostPath":
		volumeSource = HostPathVolumeSource(tieredStore.Path)
	case "persistentVolumeClaim":
		volumeSource = PvcVolumeSource(tieredStore.MediumType)
	case "emptyDir":
		volumeSource = EmptyDirVolumeSource(tieredStore.Quota)
	default:
		panic("Unknown volume type")
	}

	return corev1.Volume{
		Name:         volumeName,
		VolumeSource: volumeSource,
	}
}

func MakeTieredStoreVolumes(instance *alluxiov1alpha1.AlluxioCluster) []corev1.Volume {
	var volumes []corev1.Volume

	for _, tieredStore := range instance.Spec.ClusterConfig.TieredStore {
		mediumTypes := strings.Split(tieredStore.MediumType, ",")
		for j, mediumType := range mediumTypes {
			volume := VolumeSourceFromTieredStore(*tieredStore, mediumType, j)
			volumes = append(volumes, volume)
		}
	}

	return volumes
}

func MakeTieredStoreVolumeMounts(instance *alluxiov1alpha1.AlluxioCluster) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount
	for _, tieredStore := range instance.Spec.ClusterConfig.TieredStore {
		mount := corev1.VolumeMount{
			Name:      strings.ToLower(tieredStore.MediumType),
			MountPath: tieredStore.Path,
		}

		mounts = append(mounts, mount)
	}

	return mounts
}
