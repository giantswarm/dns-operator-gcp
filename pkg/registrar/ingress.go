package registrar

import (
	"context"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	clouddns "google.golang.org/api/dns/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	LabelIngressKey   = "app.kubernetes.io/name"
	LabelIngressValue = "nginx-ingress-controller"

	IngressNamespace = "kube-system"
	EndpointIngress  = "ingress"
)

//counterfeiter:generate . LoadBalancerClient
type LoadBalancerClient interface {
	GetIPByLabel(context.Context, string, string) (string, error)
}

type Ingress struct {
	baseDomain         string
	loadBalancerClient LoadBalancerClient
	dnsService         *clouddns.Service
}

func NewIngress(baseDomain string, dnsService *clouddns.Service, serviceClient LoadBalancerClient) *Ingress {
	return &Ingress{
		baseDomain:         baseDomain,
		loadBalancerClient: serviceClient,
		dnsService:         dnsService,
	}
}

func (r *Ingress) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Registering record")
	defer logger.Info("Done registering record")

	ingressIP, err := r.loadBalancerClient.GetIPByLabel(ctx, LabelIngressKey, LabelIngressValue)
	if err != nil {
		logger.Error(err, "Failed to get LoadBalancer ip")
		return microerror.Mask(err)
	}

	ingressDomain := fmt.Sprintf("%s.%s.%s.", EndpointIngress, cluster.Name, r.baseDomain)
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

func (r *Ingress) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("ingress-registrar")
}
