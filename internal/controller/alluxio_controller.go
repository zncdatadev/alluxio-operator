/*
Copyright 2023 zncdata-labs.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
)

// AlluxioReconciler reconciles a Alluxio object
type AlluxioReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=stack.zncdata.net,resources=alluxios,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=alluxios/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stack.zncdata.net,resources=alluxios/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Alluxio object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AlluxioReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	instance := &stackv1alpha1.Alluxio{}
	err := r.Get(ctx, types.NamespacedName{Name: "alluxio-master", Namespace: req.Namespace}, instance)
	if client.IgnoreNotFound(err) != nil {
		logger.Error(err, "unable to fetch Master StatefulSet")
		return ctrl.Result{}, err
	}

	if err := r.reconcileMasterStatefulSet(ctx, instance, r.Scheme); err != nil {
		logger.Error(err, "unable to reconcile Master StatefulSet")
		return ctrl.Result{}, err
	}

	if err := r.reconcileMasterService(ctx, instance, r.Scheme); err != nil {
		logger.Error(err, "unable to reconcile Master Service")
		return ctrl.Result{}, err
	}

	if err := r.reconcileWorkerDaemonSet(ctx, instance, r.Scheme); err != nil {
		logger.Error(err, "unable to reconcile Worker StatefulSet")
		return ctrl.Result{}, err
	}

	if err := r.reconcileWorkerService(ctx, instance, r.Scheme); err != nil {
		logger.Error(err, "unable to reconcile Worker Service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlluxioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&stackv1alpha1.Alluxio{}).
		Complete(r)
}
