package controller

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func volumeSourceFromShortCircuit(shortCircuit stackv1alpha1.ShortCircuitSpec) corev1.VolumeSource {
	switch shortCircuit.VolumeType {

	case "hostPath":
		return hostPathVolumeSource(shortCircuit)
	case "persistentVolumeClaim":
		return pvcVolumeSource(shortCircuit)
	default:
		panic("Unknown volume type")
	}
}

