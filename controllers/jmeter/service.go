package jmeter

import (
	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewService(cr *perfv1alpha1.JMeter, selector map[string]string) *v1.Service {
	objectMeta := metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}

	service := &v1.Service{
		ObjectMeta: objectMeta,
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{Port: 1099, TargetPort: intstr.FromString("rmi")},
			},
			Selector: selector,
		},
	}

	return service
}
