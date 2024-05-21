package worker

import (
	alluxiov1alpha1 "github.com/zncdatadev/alluxio-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
)

func VolumeSourceFromShortCircuit(shortCircuit alluxiov1alpha1.ShortCircuitSpec, roleGroupName string) corev1.VolumeSource {
	switch shortCircuit.VolumeType {

	case "hostPath":
		return HostPathVolumeSource(shortCircuit.HosePath)
	case "persistentVolumeClaim":
		return PvcVolumeSource(createPvcName(shortCircuit.PvcName, roleGroupName))
	default:
		panic("Unknown volume type")
	}
}

func MakeShortCircuitVolumes(instance *alluxiov1alpha1.AlluxioCluster, roleGroupName string) []corev1.Volume {
	shortCircuit := instance.Spec.ClusterConfig.GetShortCircuit()
	var volumes []corev1.Volume
	volume := corev1.Volume{
		Name:         "alluxio-domain",
		VolumeSource: VolumeSourceFromShortCircuit(shortCircuit, roleGroupName),
	}
	volumes = append(volumes, volume)

	return volumes
}

func MakeShortCircuitVolumeMounts() []corev1.VolumeMount {
	var mounts []corev1.VolumeMount
	mount := corev1.VolumeMount{
		Name:      "alluxio-domain",
		MountPath: "/opt/domain",
	}

	mounts = append(mounts, mount)

	return mounts
}

func EmptyDirVolumeSource(quota string) corev1.VolumeSource {
	sizeLimit := k8sResource.MustParse(quota)
	return corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{
			Medium:    "Memory",
			SizeLimit: &sizeLimit,
		},
	}
}

func HostPathVolumeSource(path string) corev1.VolumeSource {
	hostPathType := corev1.HostPathType("DirectoryOrCreate")
	return corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
			Type: &hostPathType,
		},
	}
}

func PvcVolumeSource(claimName string) corev1.VolumeSource {
	return corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: claimName,
		},
	}
}
