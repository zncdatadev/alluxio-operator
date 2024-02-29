package master

import (
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/alluxio-operator/internal/common"
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

func getMasterPorts(cfg *stackv1alpha1.MasterRoleGroupSpec) *stackv1alpha1.MasterPortsSpec {
	ports := cfg.Config.Ports
	if ports == nil {
		ports = &stackv1alpha1.MasterPortsSpec{
			Web:      stackv1alpha1.MasterWebPort,
			Rpc:      stackv1alpha1.MasterRpcPort,
			Embedded: stackv1alpha1.MasterEmbedded,
		}
	}
	return ports
}

// get job master port
func getJobMasterPorts(cfg *stackv1alpha1.MasterRoleGroupSpec) *stackv1alpha1.JobMasterPortsSpec {
	jobMasterPorts := cfg.Config.JobMaster.Ports
	if jobMasterPorts == nil {
		jobMasterPorts = &stackv1alpha1.JobMasterPortsSpec{
			Web:      stackv1alpha1.JobMasterWebPort,
			Rpc:      stackv1alpha1.JobMasterRpcPort,
			Embedded: stackv1alpha1.JobMasterEmbedded,
		}
	}
	return jobMasterPorts
}
