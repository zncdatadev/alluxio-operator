package controller

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func volumeSourceFromTieredStore(tieredStore stackv1alpha1.TieredStore) corev1.VolumeSource {
	switch tieredStore.Type {
	case "emptyDir":
		return emptyDirVolumeSource(tieredStore)
	case "hostPath":
		return hostPathVolumeSource(tieredStore)
	case "persistentVolumeClaim":
		return pvcVolumeSource(tieredStore)
	default:
		panic("Unknown volume type")
	}
}

func emptyDirVolumeSource(tieredStore stackv1alpha1.TieredStore) corev1.VolumeSource {
	sizeLimit := k8sResource.MustParse(tieredStore.Quota)
	return corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{
			Medium:    "Memory",
			SizeLimit: &sizeLimit,
		},
	}
}

func hostPathVolumeSource(tieredStore stackv1alpha1.TieredStore) corev1.VolumeSource {
	hostPathType := corev1.HostPathType("DirectoryOrCreate")
	return corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: tieredStore.Path,
			Type: &hostPathType,
		},
	}
}

func pvcVolumeSource(tieredStore stackv1alpha1.TieredStore) corev1.VolumeSource {
	return corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: tieredStore.MediumType,
		},
	}
}

func TieredStoreVolumes(instance *stackv1alpha1.Alluxio) []corev1.Volume {
	var pvcs []corev1.Volume
	for _, tieredStore := range instance.Spec.TieredStore {
		volumeSource := volumeSourceFromTieredStore(*tieredStore)

		pvc := corev1.Volume{
			Name:         tieredStore.Name,
			VolumeSource: volumeSource,
		}

		pvcs = append(pvcs, pvc)
	}

	return pvcs
}

func TieredStoreVolumeMounts(instance *stackv1alpha1.Alluxio) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount
	for _, tieredStore := range instance.Spec.TieredStore {
		mount := corev1.VolumeMount{
			Name:      tieredStore.Name,
			MountPath: tieredStore.Path,
		}

		mounts = append(mounts, mount)
	}

	return mounts
}

// make TieredStore pvc
func makeTieredStorePVC(tieredStore stackv1alpha1.TieredStore, namespace string, lables map[string]string, annotations map[string]string) corev1.PersistentVolumeClaim {
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:        tieredStore.MediumType,
			Namespace:   namespace,
			Labels:      lables,
			Annotations: annotations,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName("storage"): k8sResource.MustParse(tieredStore.Quota),
				},
			},
		},
	}
	return pvc
}
