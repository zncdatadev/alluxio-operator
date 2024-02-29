package worker

import (
	"github.com/zncdata-labs/alluxio-operator/internal/common"
	"github.com/zncdata-labs/alluxio-operator/internal/controller/role"
	ctrl "sigs.k8s.io/controller-runtime"
)

import (
	"context"
	"github.com/go-logr/logr"
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// roleWorker reconciler

type RoleWorker struct {
	role.BaseRoleReconciler[*stackv1alpha1.WorkerSpec]
}

// NewRoleWorker new roleWorker
func NewRoleWorker(
	scheme *runtime.Scheme,
	instance *stackv1alpha1.Alluxio,
	client client.Client,
	log logr.Logger) *RoleWorker {
	r := &RoleWorker{
		BaseRoleReconciler: role.BaseRoleReconciler[*stackv1alpha1.WorkerSpec]{
			Scheme:   scheme,
			Instance: instance,
			Client:   client,
			Log:      log,
			Role:     instance.Spec.Worker,
		},
	}
	r.Labels = r.MergeLabels()
	return r
}

func (r *RoleWorker) RoleName() role.Role {
	return role.Worker
}

func (r *RoleWorker) MergeLabels() map[string]string {
	return r.GetLabels(r.RoleName())
}

func (r *RoleWorker) ReconcileRole(ctx context.Context) (ctrl.Result, error) {
	// role pdb
	if r.Role.Config != nil && r.Role.Config.PodDisruptionBudget != nil {
		pdb := common.NewReconcilePDB(
			r.Client,
			r.Scheme,
			r.Instance,
			r.Labels,
			string(r.RoleName()),
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
		groupReconciler := NewRoleWorkerGroup(r.Scheme, r.Instance, r.Client, name, r.Log)
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

// RoleWorkerGroup worker role group reconcile
type RoleWorkerGroup struct {
	Scheme     *runtime.Scheme
	Instance   *stackv1alpha1.Alluxio
	Client     client.Client
	GroupName  string
	RoleLabels map[string]string
	Log        logr.Logger
}

func NewRoleWorkerGroup(
	scheme *runtime.Scheme,
	instance *stackv1alpha1.Alluxio,
	client client.Client,
	groupName string,
	log logr.Logger) *RoleWorkerGroup {
	r := &RoleWorkerGroup{
		Scheme:    scheme,
		Instance:  instance,
		Client:    client,
		GroupName: groupName,
		Log:       log,
	}
	return r
}

// ReconcileGroup ReconcileRole implements the Role interface
func (m *RoleWorkerGroup) ReconcileGroup(ctx context.Context) (ctrl.Result, error) {
	//reconcile all resources below
	//1. reconcile worker pdb
	//1. reconcile worker pvc
	//1. reconcile worker deployment

	//convert any to *stackv1alpha1.WorkerRoleGroupSpec
	mergedCfgObj := m.MergeGroupConfigSpec()
	mergedGroupCfg := mergedCfgObj.(*stackv1alpha1.WorkerRoleGroupSpec)
	// cache it
	common.MergedCache.Set(createMasterGroupCacheKey(m.Instance.GetName(), string(role.Worker), m.GroupName),
		mergedGroupCfg)
	mergedLabels := m.MergeLabels(mergedGroupCfg)

	if mergedGroupCfg.Config != nil && mergedGroupCfg.Config.PodDisruptionBudget != nil {
		pdb := common.NewReconcilePDB(
			m.Client,
			m.Scheme,
			m.Instance,
			mergedLabels,
			m.GroupName,
			nil)
		res, err := pdb.ReconcileResource(ctx, "", pdb)
		if err != nil {
			m.Log.Error(err, "Reconcile pdb of Worker-role failed", "groupName", m.GroupName)
			return ctrl.Result{}, err
		}
		if res.RequeueAfter > 0 {
			return res, nil
		}
	}

	pvc := NewPvc(m.Scheme, m.Instance, m.Client, mergedLabels, mergedGroupCfg)
	if _, err := pvc.ReconcileResource(ctx, m.GroupName, pvc); err != nil {
		m.Log.Error(err, "Reconcile pvc of Worker-role failed", "groupName", m.GroupName)
		return ctrl.Result{}, err
	}

	deployment := NewDeployment(m.Scheme, m.Instance, m.Client, mergedLabels, mergedGroupCfg, mergedGroupCfg.Replicas)
	if _, err := deployment.ReconcileResource(ctx, m.GroupName, deployment); err != nil {
		m.Log.Error(err, "Reconcile deployment of Worker-role failed", "groupName", m.GroupName)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (m *RoleWorkerGroup) MergeGroupConfigSpec() any {
	originWorkerCfg := m.Instance.Spec.Worker.RoleGroups[m.GroupName]
	instance := m.Instance
	// Merge the role into the role group.
	// if the role group has a config, and role group not has a config, will
	// merge the role's config into the role group's config.
	return mergeConfig(instance.Spec.Worker, originWorkerCfg)
}

func (m *RoleWorkerGroup) MergeLabels(mergedCfg any) map[string]string {
	mergedWorkerCfg := mergedCfg.(*stackv1alpha1.WorkerRoleGroupSpec)
	roleLabels := m.RoleLabels
	mergeLabels := make(common.Map)
	mergeLabels.MapMerge(roleLabels, true)
	mergeLabels.MapMerge(mergedWorkerCfg.Config.MatchLabels, true)
	mergeLabels["app.kubernetes.io/instance"] = strings.ToLower(m.GroupName)
	return mergeLabels
}

// mergeConfig merge the role's config into the role group's config
func mergeConfig(workerRole *stackv1alpha1.WorkerSpec,
	group *stackv1alpha1.WorkerRoleGroupSpec) *stackv1alpha1.WorkerRoleGroupSpec {
	copiedRoleGroup := group.DeepCopy()
	// Merge the role into the role group.
	// if the role group has a config, and role group not has a config, will
	// merge the role's config into the role group's config.
	role.MergeObjects(copiedRoleGroup, workerRole, []string{"RoleGroups"})

	// merge the role's config into the role group's config
	if workerRole.Config != nil && copiedRoleGroup.Config != nil {
		role.MergeObjects(copiedRoleGroup.Config, workerRole.Config, []string{})
	}
	return copiedRoleGroup
}
