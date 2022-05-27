package k8sclient

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service struct {
	client client.Client
}

func NewService(client client.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) GetByLabel(ctx context.Context, labelKey, labelValue string) (corev1.Service, error) {
	serviceList := &corev1.ServiceList{}
	err := s.client.List(ctx, serviceList, client.MatchingLabels{labelKey: labelValue})
	if err != nil {
		return corev1.Service{}, microerror.Mask(err)
	}

	if len(serviceList.Items) != 1 {
		return corev1.Service{}, fmt.Errorf("found %d matching services, expected 1", len(serviceList.Items))
	}

	return serviceList.Items[0], nil
}
