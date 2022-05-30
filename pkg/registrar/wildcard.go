package registrar

import (
	"context"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
	clouddns "google.golang.org/api/dns/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
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

func (w *Wildcard) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	apiDomain := fmt.Sprintf("%s.%s.%s.", EndpointWildcard, cluster.Name, w.baseDomain)
	ingressDomain := fmt.Sprintf("%s.%s.%s.", EndpointIngress, cluster.Name, w.baseDomain)

	record := &clouddns.ResourceRecordSet{
		Name: apiDomain,
		Rrdatas: []string{
			ingressDomain,
		},
		Type: RecordCNAME,
	}
	_, err := w.dnsService.ResourceRecordSets.Create(cluster.Spec.Project, cluster.Name, record).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		return nil
	}

	return microerror.Mask(err)
}

func (w *Wildcard) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	wildcardDomain := fmt.Sprintf("%s.%s.%s.", EndpointWildcard, cluster.Name, w.baseDomain)
	_, err := w.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, wildcardDomain, RecordCNAME).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		return nil
	}

	return microerror.Mask(err)
}
