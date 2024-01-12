package controller

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func volumeSourceFromShortCircuit(shortCircuit stackv1alpha1.ShortCircuitSpec) corev1.VolumeSource {
	switch shortCircuit.VolumeType {

	case "hostPath":
		return hostPathVolumeSource(shortCircuit.HosePath)
	case "persistentVolumeClaim":
		return pvcVolumeSource(shortCircuit.PvcName)
	default:
		panic("Unknown volume type")
	}
}

func makeShortCircuitVolumes(instance *stackv1alpha1.Alluxio) []corev1.Volume {
	shortCircuit := instance.Spec.ClusterConfig.GetShortCircuit()
	var volumes []corev1.Volume
	volume := corev1.Volume{
		Name:         "alluxio-domain",
		VolumeSource: volumeSourceFromShortCircuit(shortCircuit),
	}
	volumes = append(volumes, volume)

	return volumes
}
