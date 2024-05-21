package common

import (
	alluxiov1alpha1 "github.com/zncdatadev/alluxio-operator/api/v1alpha1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PDBReconciler struct {
	BaseResourceReconciler[*alluxiov1alpha1.AlluxioCluster, any]
	name   string
	labels map[string]string
	pdb    *alluxiov1alpha1.PodDisruptionBudgetSpec
}

func NewReconcilePDB(
	client client.Client,
	schema *runtime.Scheme,
	cr *alluxiov1alpha1.AlluxioCluster,
	labels map[string]string,
	name string,
	pdb *alluxiov1alpha1.PodDisruptionBudgetSpec,
) *PDBReconciler {
	return &PDBReconciler{
		BaseResourceReconciler: *NewBaseResourceReconciler[*alluxiov1alpha1.AlluxioCluster, any](
			schema,
			cr,
			client,
			"",
			labels,
			nil,
		),
		name:   name,
		labels: labels,
		pdb:    pdb,
	}
}

func (r *PDBReconciler) Build() (client.Object, error) {
	obj := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.name,
			Namespace: r.Instance.Namespace,
			Labels:    r.labels,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: r.labels,
			},
		},
	}

	if r.pdb.MinAvailable > 0 {
		obj.Spec.MinAvailable = &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: r.pdb.MinAvailable,
		}
	}

	if r.pdb.MaxUnavailable > 0 {
		obj.Spec.MaxUnavailable = &intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: r.pdb.MaxUnavailable,
		}
	}
	return obj, nil
}
