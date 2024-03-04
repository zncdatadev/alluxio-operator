package controller

import (
	"context"

	"github.com/go-logr/logr"
	alluxiov1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	"github.com/zncdata-labs/alluxio-operator/internal/common"
	"github.com/zncdata-labs/alluxio-operator/internal/controller/master"
	"github.com/zncdata-labs/alluxio-operator/internal/controller/worker"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var roles = make(map[common.Role]common.RoleReconciler)

func RegisterRole(role common.Role, roleReconciler common.RoleReconciler) {
	roles[role] = roleReconciler
}

type ClusterReconciler struct {
	client client.Client
	scheme *runtime.Scheme
	cr     *alluxiov1alpha1.AlluxioCluster
	Log    logr.Logger
}

func NewClusterReconciler(client client.Client, scheme *runtime.Scheme, cr *alluxiov1alpha1.AlluxioCluster) *ClusterReconciler {
	return &ClusterReconciler{
		client: client,
		scheme: scheme,
		cr:     cr,
	}
}

func (c *ClusterReconciler) ReconcileCluster(ctx context.Context) (ctrl.Result, error) {
	// Register roles
	// worker should be registered before master
	RegisterRole(common.Worker, worker.NewRoleWorker(c.scheme, c.cr, c.client, c.Log))
	RegisterRole(common.Master, master.NewRoleMaster(c.scheme, c.cr, c.client, c.Log))

	for _, r := range roles {
		res, err := r.ReconcileRole(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		if res.RequeueAfter > 0 {
			return res, nil
		}
	}
	return ctrl.Result{}, nil
}
