package master

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	alluxiov1alpha1 "github.com/zncdatadev/alluxio-operator/api/v1alpha1"
	"github.com/zncdatadev/alluxio-operator/internal/common"
	"github.com/zncdatadev/alluxio-operator/internal/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// roleMaster reconciler

type RoleMaster struct {
	common.BaseRoleReconciler[*alluxiov1alpha1.MasterSpec]
}

// NewRoleMaster new roleMaster
func NewRoleMaster(
	scheme *runtime.Scheme,
	instance *alluxiov1alpha1.AlluxioCluster,
	client client.Client,
	log logr.Logger) *RoleMaster {
	r := &RoleMaster{
		BaseRoleReconciler: common.BaseRoleReconciler[*alluxiov1alpha1.MasterSpec]{
			Scheme:   scheme,
			Instance: instance,
			Client:   client,
			Log:      log,
			Role:     instance.Spec.Master,
		},
	}
	r.Labels = r.MergeLabels()
	return r
}

func (r *RoleMaster) RoleName() common.Role {
	return common.Master
}

func (r *RoleMaster) MergeLabels() map[string]string {
	return r.GetLabels(r.RoleName())
}

func (r *RoleMaster) ReconcileRole(ctx context.Context) (ctrl.Result, error) {
	if r.Role.Config != nil && r.Role.Config.PodDisruptionBudget != nil {
		pdb := common.NewReconcilePDB(r.Client, r.Scheme, r.Instance, r.Labels, string(r.RoleName()),
			r.Role.Config.PodDisruptionBudget)
		res, err := pdb.ReconcileResource(ctx, "", pdb)
		if err != nil {
			return ctrl.Result{}, err
		}
		if res.RequeueAfter > 0 {
			return res, nil
		}
	}

	for name := range r.Role.RoleGroups {
		groupReconciler := NewRoleMasterGroup(r.Scheme, r.Instance, r.Client, name, r.Labels, r.Log)
		res, err := groupReconciler.ReconcileGroup(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		if res.RequeueAfter > 0 {
			return res, nil
		}
	}
	return ctrl.Result{}, nil
}

// RoleMasterGroup master role group reconcile
type RoleMasterGroup struct {
	Scheme     *runtime.Scheme
	Instance   *alluxiov1alpha1.AlluxioCluster
	Client     client.Client
	GroupName  string
	RoleLabels map[string]string
	Log        logr.Logger
}

func NewRoleMasterGroup(
	scheme *runtime.Scheme,
	instance *alluxiov1alpha1.AlluxioCluster,
	client client.Client,
	groupName string,
	roleLabels map[string]string,
	log logr.Logger) *RoleMasterGroup {
	r := &RoleMasterGroup{
		Scheme:     scheme,
		Instance:   instance,
		Client:     client,
		GroupName:  groupName,
		RoleLabels: roleLabels,
		Log:        log,
	}
	return r
}

// ReconcileGroup ReconcileRole implements the Role interface
func (m *RoleMasterGroup) ReconcileGroup(ctx context.Context) (ctrl.Result, error) {
	//reconcile all resources below
	//1. reconcile master pdb
	//1. reconcile master statefulset
	//1. reconcile master service

	//convert any to *alluxiov1alpha1.MasterRoleGroupSpec
	mergedCfgObj := m.MergeGroupConfigSpec()
	mergedGroupCfg := mergedCfgObj.(*alluxiov1alpha1.MasterRoleGroupSpec)
	// cache it
	common.MergedCache.Set(createMasterGroupCacheKey(m.Instance.GetName(), string(common.Master), m.GroupName),
		mergedGroupCfg)

	mergedLabels := m.MergeLabels(mergedGroupCfg)
	//pdb
	if mergedGroupCfg.Config != nil && mergedGroupCfg.Config.PodDisruptionBudget != nil {
		pdb := common.NewReconcilePDB(m.Client, m.Scheme, m.Instance, mergedLabels, m.GroupName, nil)
		if resource, err := pdb.ReconcileResource(ctx, m.GroupName, pdb); err != nil {
			m.Log.Error(err, "Reconcile pdb of Master-role failed", "groupName", m.GroupName)
			return ctrl.Result{}, err
		} else {
			if resource.RequeueAfter > 0 {
				return resource, nil
			}
		}
	}
	// configmap
	configmap := NewConfigMap(m.Scheme, m.Instance, m.Client, m.GroupName, mergedLabels, mergedGroupCfg)
	if _, err := configmap.ReconcileResource(ctx, m.GroupName, configmap); err != nil {
		m.Log.Error(err, "Reconcile configmap of Master-role failed", "groupName", m.GroupName)
		return ctrl.Result{}, err
	}

	//loggin  configmap
	masterLogDataBuilder := &LogDataBuilder{cfg: mergedGroupCfg}
	loggin := common.NewLoggingReconciler(
		m.Scheme, m.Instance, m.Client, m.GroupName, mergedLabels, mergedGroupCfg, masterLogDataBuilder, common.Master)
	if _, err := loggin.ReconcileResource(ctx, m.GroupName, loggin); err != nil {
		m.Log.Error(err, "Reconcile loggin of Master-role failed", "groupName", m.GroupName)
		return ctrl.Result{}, err
	}

	// statefulSet
	statefulSet := NewStatefulSet(m.Scheme, m.Instance, m.Client, m.GroupName, mergedLabels, mergedGroupCfg, mergedGroupCfg.Replicas)
	if _, err := statefulSet.ReconcileResource(ctx, m.GroupName, statefulSet); err != nil {
		m.Log.Error(err, "Reconcile statefulSet of Master-role failed", "groupName", m.GroupName)
		return ctrl.Result{}, err
	}
	// service
	svc := NewService(m.Scheme, m.Instance, m.Client, m.GroupName, mergedLabels, mergedGroupCfg)
	if _, err := svc.ReconcileResource(ctx, m.GroupName, svc); err != nil {
		m.Log.Error(err, "Reconcile service of Master-role failed", "groupName", m.GroupName)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (m *RoleMasterGroup) MergeGroupConfigSpec() any {
	originMasterCfg := m.Instance.Spec.Master.RoleGroups[m.GroupName]
	instance := m.Instance
	// Merge the role into the role group.
	// if the role group has a config, and role group not has a config, will
	// merge the role's config into the role group's config.
	return mergeConfig(instance.Spec.Master, originMasterCfg)
}

func (m *RoleMasterGroup) MergeLabels(mergedCfg any) map[string]string {
	mergedMasterCfg := mergedCfg.(*alluxiov1alpha1.MasterRoleGroupSpec)
	roleLabels := m.RoleLabels
	mergeLabels := make(util.Map)
	mergeLabels.MapMerge(roleLabels, true)
	mergeLabels.MapMerge(mergedMasterCfg.Config.MatchLabels, true)
	mergeLabels["app.kubernetes.io/instance"] = strings.ToLower(m.GroupName)
	return mergeLabels
}

// mergeConfig merge the role's config into the role group's config
func mergeConfig(masterRole *alluxiov1alpha1.MasterSpec,
	group *alluxiov1alpha1.MasterRoleGroupSpec) *alluxiov1alpha1.MasterRoleGroupSpec {
	copiedRoleGroup := group.DeepCopy()
	// Merge the role into the role group.
	// if the role group has a config, and role group not has a config, will
	// merge the role's config into the role group's config.
	common.MergeObjects(copiedRoleGroup, masterRole, []string{"RoleGroups"})

	// merge the role's config into the role group's config
	if masterRole.Config != nil && copiedRoleGroup.Config != nil {
		common.MergeObjects(copiedRoleGroup.Config, masterRole.Config, []string{})
	}
	return copiedRoleGroup
}

type LogDataBuilder struct {
	cfg *alluxiov1alpha1.MasterRoleGroupSpec
}

// MakeContainerLog4jData implement RoleLoggingDataBuilder
func (c *LogDataBuilder) MakeContainerLog4jData() map[string]string {
	cfg := c.cfg
	data := make(map[string]string)
	//master logger data
	if cfg.Config.Logging != nil {
		masterLogger := common.PropertiesValue(common.MasterLogger, cfg.Config.Logging.Metastore)
		data[common.CreateLoggerConfigMapKey(common.MasterLogger)] = masterLogger
	}
	//job master logger data
	if cfg.Config.JobMaster.Logging != nil {
		jobMasterLogger := common.PropertiesValue(common.JobMasterLogger, cfg.Config.JobMaster.Logging.Metastore)
		data[common.CreateLoggerConfigMapKey(common.JobMasterLogger)] = jobMasterLogger
	}
	return data
}
