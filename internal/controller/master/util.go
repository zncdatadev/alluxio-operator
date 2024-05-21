package master

import (
	alluxiov1alpha1 "github.com/zncdatadev/alluxio-operator/api/v1alpha1"
	"github.com/zncdatadev/alluxio-operator/internal/common"
)

func createMasterSvcName(instanceName string, roleName string, groupName string, extraSuffix string) string {
	return common.NewResourceNameGenerator(instanceName, roleName, groupName).GenerateResourceName(extraSuffix)
}

func createMasterStatefulSetName(instanceName string, roleName string, groupName string) string {
	return common.NewResourceNameGenerator(instanceName, roleName, groupName).GenerateResourceName("")
}

func createMasterGroupCacheKey(instanceName string, roleName string, groupName string) string {
	return common.NewResourceNameGenerator(instanceName, roleName, groupName).GenerateResourceName("cache")
}

func getMasterPorts(cfg *alluxiov1alpha1.MasterRoleGroupSpec) *alluxiov1alpha1.MasterPortsSpec {
	ports := cfg.Config.Ports
	if ports == nil {
		ports = &alluxiov1alpha1.MasterPortsSpec{
			Web:      alluxiov1alpha1.MasterWebPort,
			Rpc:      alluxiov1alpha1.MasterRpcPort,
			Embedded: alluxiov1alpha1.MasterEmbedded,
		}
	}
	return ports
}

// get job master port
func getJobMasterPorts(cfg *alluxiov1alpha1.MasterRoleGroupSpec) *alluxiov1alpha1.JobMasterPortsSpec {
	jobMasterPorts := cfg.Config.JobMaster.Ports
	if jobMasterPorts == nil {
		jobMasterPorts = &alluxiov1alpha1.JobMasterPortsSpec{
			Web:      alluxiov1alpha1.JobMasterWebPort,
			Rpc:      alluxiov1alpha1.JobMasterRpcPort,
			Embedded: alluxiov1alpha1.JobMasterEmbedded,
		}
	}
	return jobMasterPorts
}
