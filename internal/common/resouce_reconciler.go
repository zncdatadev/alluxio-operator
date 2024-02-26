package common

import (
	"context"
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	opgostatus "github.com/zncdata-labs/operator-go/pkg/status"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

type IReconciler interface {
	ReconcileResource(ctx context.Context, groupName string, do IDoReconciler) (ctrl.Result, error)
}

type IDoReconciler interface {
	Build(data ResourceBuilderData) (client.Object, error)
	DoReconcile(ctx context.Context, resource client.Object) (ctrl.Result, error)
}

type BaseResourceReconciler[T client.Object, G any] struct {
	Instance T
	Scheme   *runtime.Scheme
	Client   client.Client
	RoleName string

	MergedLabels map[string]string
	MergedCfg    G
}

// NewBaseResourceReconciler new a BaseResourceReconciler
func NewBaseResourceReconciler[T client.Object, G any](
	scheme *runtime.Scheme,
	instance T,
	client client.Client,
	mergedLabels map[string]string,
	mergedCfg G) *BaseResourceReconciler[T, G] {
	return &BaseResourceReconciler[T, G]{
		Instance:     instance,
		Scheme:       scheme,
		Client:       client,
		MergedLabels: mergedLabels,
		MergedCfg:    mergedCfg,
	}
}

type ResourceBuilderData struct {
	Labels    map[string]string
	GroupName string
	//group config spec after merge
	MergedGroupCfg any
}

// NewResourceBuilderData returns a new ResourceBuilderData
func NewResourceBuilderData(labels map[string]string, groupName string,
	mergedGroupCfg any) ResourceBuilderData {
	return ResourceBuilderData{
		Labels:         labels,
		GroupName:      groupName,
		MergedGroupCfg: mergedGroupCfg,
	}
}

func (b *BaseResourceReconciler[T, G]) ReconcileResource(ctx context.Context, groupName string, do IDoReconciler) (ctrl.Result, error) {
	// 1. mergelables
	// 2. build resource
	// 3. setControllerReference
	data := NewResourceBuilderData(b.MergedLabels, groupName, b.MergedCfg)
	resource, err := do.Build(data)
	if err != nil {
		return ctrl.Result{}, err
	} else {
		if err := b.setControllerReference(resource); err != nil {
			return ctrl.Result{}, err
		}
	}
	//do reconcile
	//return b.DoReconcile(ctx, resource)
	if i, ok := do.(IDoReconciler); ok {
		return i.DoReconcile(ctx, resource)
	} else {
		return ctrl.Result{}, nil
	}
}

func (b *BaseResourceReconciler[T, G]) setControllerReference(resource client.Object) error {
	err := controllerutil.SetControllerReference(b.Instance, resource, b.Scheme)
	if err != nil {
		return err
	}
	return nil
}

func (b *BaseResourceReconciler[T, G]) apply(ctx context.Context, dep client.Object) (ctrl.Result, error) {
	if err := ctrl.SetControllerReference(b.Instance, dep, b.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	mutant, err := CreateOrUpdate(ctx, b.Client, dep)
	if err != nil {
		return ctrl.Result{}, err
	}

	if mutant {
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	return ctrl.Result{}, nil
}

// GeneralResourceStyleReconciler general style resource reconcile
// this reconciler is used to reconcile the general style resources
// such as configMap, secret, svc, etc.
type GeneralResourceStyleReconciler[T client.Object, G any] struct {
	BaseResourceReconciler[T, G]
}

func NewGeneraResourceStyleReconciler[T client.Object, G any](
	scheme *runtime.Scheme,
	instance T,
	client client.Client,
	mergedLabels map[string]string,
	mergedCfg G,
) *GeneralResourceStyleReconciler[T, G] {
	return &GeneralResourceStyleReconciler[T, G]{
		BaseResourceReconciler: *NewBaseResourceReconciler[T, G](
			scheme,
			instance,
			client,
			mergedLabels,
			mergedCfg),
	}
}

// override doReconcile method
func (s *GeneralResourceStyleReconciler[T, G]) DoReconcile(
	ctx context.Context,
	resource client.Object) (ctrl.Result, error) {
	return s.apply(ctx, resource)
}

// DeploymentStyleReconciler deployment style reconciler
// this reconciler is used to reconcile the deployment style resources
// such as deployment, statefulSet, etc.
// it will do the following things:
// 1. apply the resource
// 2. check if the resource is satisfied
// 3. if not, return requeue
// 4. if satisfied, return nil

type DeploymentStyleReconciler[T client.Object, G any] struct {
	BaseResourceReconciler[T, G]
	replicas int32
}

func NewDeploymentStyleReconciler[T client.Object, G any](
	scheme *runtime.Scheme,
	instance T,
	client client.Client,
	mergedLabels map[string]string,
	mergedCfg G,
	replicas int32,
) *DeploymentStyleReconciler[T, G] {
	return &DeploymentStyleReconciler[T, G]{
		BaseResourceReconciler: *NewBaseResourceReconciler[T, G](
			scheme,
			instance,
			client,
			mergedLabels,
			mergedCfg),
		replicas: replicas,
	}
}

func (s *DeploymentStyleReconciler[T, G]) DoReconcile(ctx context.Context, resource client.Object) (ctrl.Result, error) {
	// apply resource
	// check if the resource is satisfied
	// if not, return requeue
	// if satisfied, return nil

	s.CommandOverride(resource)
	s.EnvOverride(resource)

	if res, err := s.apply(ctx, resource); err != nil {
		return ctrl.Result{}, err
	} else if res.RequeueAfter > 0 {
		return res, nil
	}

	// Check if the pods are satisfied
	satisfied, err := s.CheckPodsSatisfied(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	if satisfied {
		err = s.updateStatus(
			metav1.ConditionTrue,
			"DeploymentSatisfied",
			"Deployment is satisfied",
		)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	err = s.updateStatus(
		metav1.ConditionFalse,
		"DeploymentNotSatisfied",
		"Deployment is not satisfied",
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

type ConditonsGetter interface {
	GetConditions() *[]metav1.Condition
}

type DeploymentStyleOverride interface {
	CommandOverride(resource client.Object)
	EnvOverride(resource client.Object)
}

func (s *DeploymentStyleReconciler[T, G]) CheckPodsSatisfied(ctx context.Context) (bool, error) {
	pods := corev1.PodList{}
	podListOptions := []client.ListOption{
		client.InNamespace(s.Instance.GetNamespace()),
		client.MatchingLabels(s.MergedLabels),
	}
	err := s.Client.List(ctx, &pods, podListOptions...)
	if err != nil {
		return false, err
	}

	return len(pods.Items) == int(s.replicas), nil
}

func (s *DeploymentStyleReconciler[T, G]) updateStatus(
	status metav1.ConditionStatus,
	reason string,
	message string) error {
	apimeta.SetStatusCondition(s.GetConditions(), metav1.Condition{
		Type:               opgostatus.ConditionTypeAvailable,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: s.Instance.GetGeneration(),
	})
	return s.Client.Status().Update(context.Background(), s.Instance)
}

func ConvertToResourceRequirements(resources *stackv1alpha1.ResourcesSpec) *corev1.ResourceRequirements {
	if resources == nil {
		return nil
	}
	return &corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{corev1.ResourceCPU: *resources.CPU.Max, corev1.ResourceMemory: *resources.Memory.Limit},
		Requests: corev1.ResourceList{corev1.ResourceCPU: *resources.CPU.Min, corev1.ResourceMemory: *resources.Memory.Limit},
	}
}
