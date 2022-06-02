package k8sclient

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LoadBalancer struct {
	namespace string
	client    client.Client
}

func NewLoadBalancer(namespace string, client client.Client) *LoadBalancer {
	return &LoadBalancer{
		client: client,
	}
}

func (s *LoadBalancer) GetIPByLabel(ctx context.Context, labelKey, labelValue string) (string, error) {
	serviceList := &corev1.ServiceList{}
	err := s.client.List(
		ctx,
		serviceList,
		client.InNamespace(s.namespace),
		client.MatchingLabels{labelKey: labelValue},
	)
	if err != nil {
		return "", microerror.Mask(err)
	}

	services := []corev1.Service{}
	for _, s := range serviceList.Items {
		if s.Spec.Type == corev1.ServiceTypeLoadBalancer {
			services = append(services, s)
		}
	}

	if len(services) != 1 {
		return "", fmt.Errorf(
			"found %d LoadBalancer services matching label %q: %q, expected 1",
			len(services),
			labelKey, labelValue,
		)
	}

	service := services[0]

	if len(service.Status.LoadBalancer.Ingress) != 1 {
		return "", fmt.Errorf(
			"service %s/%s has %d LoadBalancer ingresses, expected 1",
			service.Name,
			service.Namespace,
			len(service.Status.LoadBalancer.Ingress),
		)
	}

	return service.Status.LoadBalancer.Ingress[0].IP, nil
}
