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

//counterfeiter:generate . BastionsClient
type BastionsClient interface {
	GetBastionIPList(ctx context.Context, cluster *capg.GCPCluster) ([]string, error)
	AddFinalizerToBastions(ctx context.Context, cluster *capg.GCPCluster) error
	RemoveFinalizerFromBastions(ctx context.Context, cluster *capg.GCPCluster) error
}

type Bastion struct {
	baseDomain     string
	bastionsClient BastionsClient
	dnsService     *clouddns.Service
}

func NewBastion(baseDomain string, bastionsClient BastionsClient, dnsService *clouddns.Service) *Bastion {
	return &Bastion{
		baseDomain:     baseDomain,
		bastionsClient: bastionsClient,
		dnsService:     dnsService,
	}
}

func (r *Bastion) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)
	err := r.bastionsClient.AddFinalizerToBastions(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	bastionIPList, err := r.bastionsClient.GetBastionIPList(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	for i, bastionIP := range bastionIPList {
		bastionDomain := fmt.Sprintf("%s.%s.%s.", EndpointBastion(i+1), cluster.Name, r.baseDomain)
		logger.Info(fmt.Sprintf("Registering record for %s", EndpointBastion(i+1)))

		record := &clouddns.ResourceRecordSet{
			Name: bastionDomain,
			Rrdatas: []string{
				bastionIP,
			},
			Type: RecordA,
		}
		_, err = r.dnsService.ResourceRecordSets.Create(cluster.Spec.Project, cluster.Name, record).
			Context(ctx).
			Do()

		if hasHttpCode(err, http.StatusConflict) {
			// record exists, check if the IP matches
			rr, err := r.dnsService.ResourceRecordSets.Get(cluster.Spec.Project, cluster.Name, bastionDomain, RecordA). //nolint:govet
																	Context(ctx).
																	Do()
			if err != nil {
				return microerror.Mask(err)
			}

			if len(rr.Rrdatas) != 1 || rr.Rrdatas[0] != bastionIP {
				logger.Info(fmt.Sprintf("Bastion record %s exists but its not up to date. Udating record", EndpointBastion(i+1)))

				_, err = r.dnsService.ResourceRecordSets.Patch(cluster.Spec.Project, cluster.Name, bastionDomain, RecordA, record).
					Context(ctx).
					Do()
				if err != nil {
					return microerror.Mask(err)
				}
				logger.Info(fmt.Sprintf("Updated Bastion record %s.", EndpointBastion(i+1)))
			} else {
				logger.Info(fmt.Sprintf("Skipping. Record for %s already exists", EndpointBastion(i+1)))
				continue
			}
		}
		if err != nil {
			return microerror.Mask(err)
		}
		logger.Info(fmt.Sprintf("Done Registering record for %s", EndpointBastion(i+1)))
	}

	return nil
}

func (r *Bastion) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	bastionIPList, err := r.bastionsClient.GetBastionIPList(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	for i := range bastionIPList {
		bastionDomain := fmt.Sprintf("%s.%s.%s.", EndpointBastion(i+1), cluster.Name, r.baseDomain)
		logger.Info(fmt.Sprintf("Unregistering record for %s", EndpointBastion(i+1)))

		_, err = r.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, bastionDomain, RecordA).
			Context(ctx).
			Do()

		if hasHttpCode(err, http.StatusNotFound) {
			logger.Info(fmt.Sprintf("Skipping. Record for %s already unregistered", EndpointBastion(i+1)))
			continue
		}
		if err != nil {
			return microerror.Mask(err)
		}
		logger.Info(fmt.Sprintf("Done unregistering record for %s", EndpointBastion(i+1)))
	}

	err = r.bastionsClient.RemoveFinalizerFromBastions(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Bastion) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("bastion-registrar")
}

func EndpointBastion(index int) string {
	return fmt.Sprintf("bastion%d", index)
}
