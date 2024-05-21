package worker

import (
	"github.com/zncdatadev/alluxio-operator/internal/common"
)

func createDeploymentName(instanceName string, roleName string, groupName string) string {
	return common.NewResourceNameGenerator(instanceName, roleName, groupName).GenerateResourceName("")
}

func createWorkerGroupCacheKey(instanceName string, roleName string, groupName string) string {
	return common.NewResourceNameGenerator(instanceName, roleName, groupName).GenerateResourceName("cache")
}

func createPvcName(pvcName string, groupName string) string {
	return pvcName + "-" + groupName
}
