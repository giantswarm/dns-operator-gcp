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

const EndpointWildcard = "*"

type Wildcard struct {
	baseDomain string
	dnsService *clouddns.Service
}

func NewWildcard(baseDomain string, dnsService *clouddns.Service) *Wildcard {
	return &Wildcard{
		baseDomain: baseDomain,
		dnsService: dnsService,
	}
}

func (r *Wildcard) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Registering record")
	defer logger.Info("Done registering record")

	apiDomain := fmt.Sprintf("%s.%s.%s.", EndpointWildcard, cluster.Name, r.baseDomain)
	ingressDomain := fmt.Sprintf("%s.%s.%s.", EndpointIngress, cluster.Name, r.baseDomain)

	record := &clouddns.ResourceRecordSet{
		Name: apiDomain,
		Rrdatas: []string{
			ingressDomain,
		},
		Type: RecordCNAME,
	}
	_, err := r.dnsService.ResourceRecordSets.Create(cluster.Spec.Project, cluster.Name, record).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		logger.Info("Skipping. Record already exists")
		return nil
	}

	return microerror.Mask(err)
}

func (r *Wildcard) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Unregistering record")
	defer logger.Info("Done unregistering record")

	wildcardDomain := fmt.Sprintf("%s.%s.%s.", EndpointWildcard, cluster.Name, r.baseDomain)
	_, err := r.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, wildcardDomain, RecordCNAME).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		logger.Info("Skipping. Record already unregistered")
		return nil
	}

	return microerror.Mask(err)
}

func (r *Wildcard) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("wildcard-registrar")
}
