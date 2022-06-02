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

func (lb *LoadBalancer) GetIPByLabel(ctx context.Context, labelKey, labelValue string) (string, error) {
	service, err := lb.getService(ctx, labelKey, labelValue)
	if err != nil {
		return "", err
	}

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

func (lb *LoadBalancer) getService(ctx context.Context, labelKey, labelValue string) (corev1.Service, error) {
	serviceList := &corev1.ServiceList{}
	err := lb.client.List(
		ctx,
		serviceList,
		client.InNamespace(lb.namespace),
		client.MatchingLabels{labelKey: labelValue},
	)
	if err != nil {
		return corev1.Service{}, microerror.Mask(err)
	}

	loadBalancerServices := []corev1.Service{}
	for _, s := range serviceList.Items {
		if s.Spec.Type == corev1.ServiceTypeLoadBalancer {
			loadBalancerServices = append(loadBalancerServices, s)
		}
	}

	if len(loadBalancerServices) != 1 {
		return corev1.Service{}, fmt.Errorf(
			"found %d LoadBalancer services matching label %q: %q, expected 1",
			len(loadBalancerServices),
			labelKey, labelValue,
		)
	}

	return loadBalancerServices[0], nil
}
