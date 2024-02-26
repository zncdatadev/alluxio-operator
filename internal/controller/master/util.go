package master

import (
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
