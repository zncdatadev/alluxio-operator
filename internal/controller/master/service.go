package master

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/alluxio-operator/internal/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceReconciler struct {
	common.GeneralResourceStyleReconciler[*stackv1alpha1.Alluxio, *stackv1alpha1.MasterRoleGroupSpec]
}

// NewService New a ServiceReconciler
func NewService(
	scheme *runtime.Scheme,
	instance *stackv1alpha1.Alluxio,
	client client.Client,
	mergedLabels map[string]string,
	mergedCfg *stackv1alpha1.MasterRoleGroupSpec,
) *ServiceReconciler {
	return &ServiceReconciler{
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

func (s *ServiceReconciler) Build(data common.ResourceBuilderData) (client.Object, error) {
	instance := s.Instance
	mergedGroupCfg := s.MergedCfg
	roleGroupName := data.GroupName
	roleName := s.RoleName
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createMasterSvcName(instance.Name, roleGroupName, roleName, "0"),
			Namespace: instance.Namespace,
			Labels:    s.MergedLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "rpc",
					Port: mergedGroupCfg.Config.Ports.Rpc,
				},
				{
					Name: "web",
					Port: mergedGroupCfg.Config.Ports.Web,
				},
				{
					Name: "embedded",
					Port: mergedGroupCfg.Config.Ports.Embedded,
				},
				{
					Name: "job-rpc",
					Port: mergedGroupCfg.Config.JobMaster.Ports.Rpc,
				},
				{
					Name: "job-web",
					Port: mergedGroupCfg.Config.JobMaster.Ports.Web,
				},
				{
					Name: "job-embedded",
					Port: mergedGroupCfg.Config.JobMaster.Ports.Embedded,
				},
			},
			Selector:  s.MergedLabels,
			ClusterIP: "None",
		},
	}
	return svc, nil
}
