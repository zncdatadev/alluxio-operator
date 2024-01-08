package controller

import (
	"context"
	stackv1alpha1 "github.com/zncdata-labs/alluxio-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func getPortsFromValue(v reflect.Value, namePrefix *string) []corev1.ServicePort {
	var ports []corev1.ServicePort
	typeOf := v.Type()
	for i := 0; i < v.NumField(); i++ {
		name := typeOf.Field(i).Name
		if namePrefix != nil {
			name = *namePrefix + name
		}
		ports = append(ports, corev1.ServicePort{
			Name: name,
			Port: v.Index(i).Interface().(int32),
		})
	}
	return ports
}

func createService(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme, ports []corev1.ServicePort, serviceName string) *corev1.Service {
	logger := log.FromContext(ctx)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + serviceName,
			Namespace: instance.Namespace,
		},

		Spec: corev1.ServiceSpec{
			Ports: ports,
		},
	}

	if err := controllerutil.SetControllerReference(instance, svc, schema); err != nil {
		logger.Error(err, "Failed to set owner for "+serviceName)
		return nil
	}

	return svc
}

func makeMasterService(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) *corev1.Service {

	masterPortsValue := reflect.ValueOf(instance.Spec.Master.Ports)
	ports := getPortsFromValue(masterPortsValue, nil)
	prefix := "job-"
	jobMasterPortsValue := reflect.ValueOf(instance.Spec.JobMaster.Ports)
	jobPorts := getPortsFromValue(jobMasterPortsValue, &prefix)

	ports = append(ports, jobPorts...)
	svc := createService(ctx, instance, schema, ports, "svc-master")

	return svc
}

func makeWorkerService(ctx context.Context, instance *stackv1alpha1.Alluxio, schema *runtime.Scheme) *corev1.Service {

	workerPortsValue := reflect.ValueOf(instance.Spec.Worker.Ports)
	ports := getPortsFromValue(workerPortsValue, nil)
	prefix := "job-"
	jobWorkerPortsValue := reflect.ValueOf(instance.Spec.JobWorker.Ports)
	jobPorts := getPortsFromValue(jobWorkerPortsValue, &prefix)
	ports = append(ports, jobPorts...)
	svc := createService(ctx, instance, schema, ports, "svc-worker")

	return svc
}
