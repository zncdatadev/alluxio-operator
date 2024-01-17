package controller

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func MasterScheduler(instance *stackv1alpha1.Alluxio, dep *appsv1.StatefulSet, roleGroup *stackv1alpha1.RoleMasterSpec) {

	if roleGroup.NodeSelector != nil {
		dep.Spec.Template.Spec.NodeSelector = roleGroup.NodeSelector
	} else if instance.Spec.Master.RoleConfig != nil && instance.Spec.Master.RoleConfig.NodeSelector != nil {
		dep.Spec.Template.Spec.NodeSelector = instance.Spec.Master.RoleConfig.NodeSelector
	} else if instance.Spec.ClusterConfig != nil && instance.Spec.ClusterConfig.NodeSelector != nil {
		dep.Spec.Template.Spec.NodeSelector = instance.Spec.ClusterConfig.NodeSelector
	}

	if roleGroup.Tolerations != nil {
		dep.Spec.Template.Spec.Tolerations = []corev1.Toleration{
			*roleGroup.Tolerations.DeepCopy(),
		}
	} else if instance.Spec.Master.RoleConfig != nil && instance.Spec.Master.RoleConfig.Tolerations != nil {
		dep.Spec.Template.Spec.Tolerations = []corev1.Toleration{
			*instance.Spec.Master.RoleConfig.Tolerations.DeepCopy(),
		}
	} else if instance.Spec.ClusterConfig != nil && instance.Spec.ClusterConfig.Tolerations != nil {
		dep.Spec.Template.Spec.Tolerations = []corev1.Toleration{
			*instance.Spec.ClusterConfig.Tolerations.DeepCopy(),
		}
	}

	if roleGroup.Affinity != nil {
		dep.Spec.Template.Spec.Affinity = roleGroup.Affinity.DeepCopy()
	} else if instance.Spec.Master.RoleConfig != nil && instance.Spec.Master.RoleConfig.Affinity != nil {
		dep.Spec.Template.Spec.Affinity = instance.Spec.Master.RoleConfig.Affinity.DeepCopy()
	} else if instance.Spec.ClusterConfig != nil && instance.Spec.ClusterConfig.Affinity != nil {
		dep.Spec.Template.Spec.Affinity = instance.Spec.ClusterConfig.Affinity.DeepCopy()
	}
}
