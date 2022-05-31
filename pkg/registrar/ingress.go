package registrar

import (
	"context"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	clouddns "google.golang.org/api/dns/v1"
	corev1 "k8s.io/api/core/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func (r *Ingress) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := log.FromContext(ctx)
	logger = logger.WithName("ingress-registrar")

	logger.Info("Registering record")
	defer logger.Info("Done registering record")

	ingressDomain := fmt.Sprintf("%s.%s.%s.", EndpointIngress, cluster.Name, r.baseDomain)
	ingressIP, err := r.getLoadBalancerIP(ctx, logger)
	if err != nil {
		return microerror.Mask(err)
	}

	record := &clouddns.ResourceRecordSet{
		Name: ingressDomain,
		Rrdatas: []string{
			ingressIP,
		},
		Type: RecordA,
	}
	_, err = r.dnsService.ResourceRecordSets.Create(cluster.Spec.Project, cluster.Name, record).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		logger.Info("Skipping. Record already exists")
		return nil
	}

	return microerror.Mask(err)
}

func (r *Ingress) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Unregistering record")
	defer logger.Info("Done unregistering record")

	ingressDomain := fmt.Sprintf("%s.%s.%s.", EndpointIngress, cluster.Name, r.baseDomain)
	_, err := r.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, ingressDomain, RecordA).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		logger.Info("Skipping. Record already unregistered")
		return nil
	}

	return microerror.Mask(err)
}

func (r *Ingress) getLoadBalancerIP(ctx context.Context, logger logr.Logger) (string, error) {
	service, err := r.serviceClient.GetByLabel(ctx, LabelIngressKey, LabelIngressValue)
	if err != nil {
		logger.Error(err, "Failed to get LoadBalancer service account")
		return "", err
	}

	if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
		logger.Error(err,
			"Service not LoadBalancer type",
			"service.name", service.Name,
			"service.namespace", service.Namespace,
		)
		return "", fmt.Errorf("found %s Service, expected type LoadBalancer", service.Spec.Type)
	}

	if len(service.Status.LoadBalancer.Ingress) != 1 {
		logger.Error(err,
			"Found more than one Load Balancer ingresses",
			"service.name", service.Name,
			"service.namespace", service.Namespace,
		)
		return "", fmt.Errorf(
			"found %d LoadBalancer ingresses, expected 1",
			len(service.Status.LoadBalancer.Ingress),
		)
	}

	return service.Status.LoadBalancer.Ingress[0].IP, nil
}

func (r *Ingress) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("ingress-registrar")
}
