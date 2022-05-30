package registrar

import (
	"context"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
	clouddns "google.golang.org/api/dns/v1"
	corev1 "k8s.io/api/core/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
)

const (
	LabelIngressKey   = "app.kubernetes.io/name"
	LabelIngressValue = "nginx-ingress-controller"

	IngressNamespace = "kube-system"
	EndpointIngress  = "ingress"
)

//counterfeiter:generate . ServiceClient
type ServiceClient interface {
	GetByLabel(context.Context, string, string) (corev1.Service, error)
}

type Ingress struct {
	baseDomain    string
	serviceClient ServiceClient
	dnsService    *clouddns.Service
}

func NewIngress(baseDomain string, dnsService *clouddns.Service, serviceClient ServiceClient) *Ingress {
	return &Ingress{
		baseDomain:    baseDomain,
		serviceClient: serviceClient,
		dnsService:    dnsService,
	}
}

func (i *Ingress) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	service, err := i.serviceClient.GetByLabel(ctx, LabelIngressKey, LabelIngressValue)
	if err != nil {
		return microerror.Mask(err)
	}

	if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return fmt.Errorf("found %s Service, expected type LoadBalancer", service.Spec.Type)
	}

	if len(service.Status.LoadBalancer.Ingress) != 1 {
		return fmt.Errorf(
			"found %d LoadBalancer ingresses, expected 1",
			len(service.Status.LoadBalancer.Ingress),
		)
	}

	ingressIP := service.Status.LoadBalancer.Ingress[0].IP
	ingressDomain := fmt.Sprintf("%s.%s.%s.", EndpointIngress, cluster.Name, i.baseDomain)

	record := &clouddns.ResourceRecordSet{
		Name: ingressDomain,
		Rrdatas: []string{
			ingressIP,
		},
		Type: RecordA,
	}
	_, err = i.dnsService.ResourceRecordSets.Create(cluster.Spec.Project, cluster.Name, record).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		return nil
	}

	return nil
}

func (i *Ingress) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	ingressDomain := fmt.Sprintf("%s.%s.%s.", EndpointIngress, cluster.Name, i.baseDomain)
	_, err := i.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, ingressDomain, RecordA).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		return nil
	}

	return microerror.Mask(err)
}
