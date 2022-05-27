package registrar

import (
	"context"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
	clouddns "google.golang.org/api/dns/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
)

type API struct {
	dnsService *clouddns.Service

	baseDomain       string
	parentGCPProject string
}

func NewAPI(baseDomain string, dnsService *clouddns.Service) *API {
	return &API{
		baseDomain: baseDomain,
		dnsService: dnsService,
	}
}

func (c *API) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	if cluster.Spec.ControlPlaneEndpoint.Host == "" {
		return nil
	}

	apiDomain := fmt.Sprintf("%s.%s.%s.", EndpointAPI, cluster.Name, c.baseDomain)

	record := &clouddns.ResourceRecordSet{
		Name: apiDomain,
		Rrdatas: []string{
			cluster.Spec.ControlPlaneEndpoint.Host,
		},
		Type: RecordA,
	}
	_, err := c.dnsService.ResourceRecordSets.Create(cluster.Spec.Project, cluster.Name, record).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		return nil
	}
	return microerror.Mask(err)
}

func (c *API) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	apiDomain := fmt.Sprintf("%s.%s.%s.", EndpointAPI, cluster.Name, c.baseDomain)
	_, err := c.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, apiDomain, RecordA).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		return nil
	}

	return microerror.Mask(err)
}

func (c *API) getClusterDomain(cluster *capg.GCPCluster) string {
	return fmt.Sprintf("%s.%s.", cluster.Name, c.baseDomain)
}
