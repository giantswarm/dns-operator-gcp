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

const EndpointAPI = "api"

type API struct {
	baseDomain string
	dnsService *clouddns.Service
}

func NewAPI(baseDomain string, dnsService *clouddns.Service) *API {
	return &API{
		baseDomain: baseDomain,
		dnsService: dnsService,
	}
}

func (r *API) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Registering record")
	defer logger.Info("Done registering record")

	if cluster.Spec.ControlPlaneEndpoint.Host == "" {
		logger.Info("Skipping. Cluster does not have controplane endpoint yet")
		return nil
	}

	apiDomain := fmt.Sprintf("%s.%s.%s.", EndpointAPI, cluster.Name, r.baseDomain)

	record := &clouddns.ResourceRecordSet{
		Name: apiDomain,
		Rrdatas: []string{
			cluster.Spec.ControlPlaneEndpoint.Host,
		},
		Type: RecordA,
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

func (r *API) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Unregistering record")
	defer logger.Info("Done unregistering record")

	apiDomain := fmt.Sprintf("%s.%s.%s.", EndpointAPI, cluster.Name, r.baseDomain)
	_, err := r.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, apiDomain, RecordA).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		logger.Info("Skipping. Record already unregistered")
		return nil
	}

	return microerror.Mask(err)
}

func (r *API) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("api-registrar")
}
