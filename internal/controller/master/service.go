package master

import (
	alluxiov1alpha1 "github.com/zncdatadev/alluxio-operator/api/v1alpha1"
	"github.com/zncdatadev/alluxio-operator/internal/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceReconciler struct {
	common.GeneralResourceStyleReconciler[*alluxiov1alpha1.AlluxioCluster, *alluxiov1alpha1.MasterRoleGroupSpec]
}

// NewService New a ServiceReconciler
func NewService(
	scheme *runtime.Scheme,
	instance *alluxiov1alpha1.AlluxioCluster,
	client client.Client,
	groupName string,
	mergedLabels map[string]string,
	mergedCfg *alluxiov1alpha1.MasterRoleGroupSpec,
) *ServiceReconciler {
	return &ServiceReconciler{
		GeneralResourceStyleReconciler: *common.NewGeneraResourceStyleReconciler[*alluxiov1alpha1.AlluxioCluster,
			*alluxiov1alpha1.MasterRoleGroupSpec](
			scheme,
			instance,
			client,
			groupName,
			mergedLabels,
			mergedCfg,
		),
	}
}

func (s *ServiceReconciler) Build() (client.Object, error) {
	instance := s.Instance
	mergedGroupCfg := s.MergedCfg
	roleGroupName := s.GroupName
	roleName := common.Master
	masterPorts := getMasterPorts(mergedGroupCfg)
	jobMasterPorts := getJobMasterPorts(mergedGroupCfg)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createMasterSvcName(instance.Name, string(roleName), roleGroupName, "0"),
			Namespace: instance.Namespace,
			Labels:    s.MergedLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "rpc",
					Port: masterPorts.Rpc,
				},
				{
					Name: "web",
					Port: masterPorts.Web,
				},
				{
					Name: "embedded",
					Port: masterPorts.Embedded,
				},
				{
					Name: "job-rpc",
					Port: jobMasterPorts.Rpc,
				},
				{
					Name: "job-web",
					Port: jobMasterPorts.Web,
				},
				{
					Name: "job-embedded",
					Port: jobMasterPorts.Embedded,
				},
			},
			Selector:  s.MergedLabels,
			ClusterIP: "None",
		},
	}
	return svc, nil
}
