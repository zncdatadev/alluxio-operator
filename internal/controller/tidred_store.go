package controller

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func volumeSourceFromTieredStore(tieredStore stackv1alpha1.TieredStore, mediumType string, index int) corev1.Volume {
	volumeName := strings.ToLower(mediumType)
	var volumeSource corev1.VolumeSource

	switch tieredStore.Type {
	case "hostPath":
		volumeSource = hostPathVolumeSource(tieredStore.Path)
	case "persistentVolumeClaim":
		volumeSource = pvcVolumeSource(tieredStore.MediumType)
	case "emptyDir":
		volumeSource = emptyDirVolumeSource(tieredStore.Quota)
	default:
		panic("Unknown volume type")
	}

	return corev1.Volume{
		Name:         volumeName,
		VolumeSource: volumeSource,
	}
}

func makeTieredStoreVolumes(instance *stackv1alpha1.Alluxio) []corev1.Volume {
	var volumes []corev1.Volume

	for _, tieredStore := range instance.Spec.ClusterConfig.TieredStore {
		mediumTypes := strings.Split(tieredStore.MediumType, ",")
		for j, mediumType := range mediumTypes {
			volume := volumeSourceFromTieredStore(*tieredStore, mediumType, j)
			volumes = append(volumes, volume)
		}
	}

	return volumes
}

func makeTieredStoreVolumeMounts(instance *stackv1alpha1.Alluxio) []corev1.VolumeMount {
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
