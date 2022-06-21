package registrar

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	clouddns "google.golang.org/api/dns/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//counterfeiter:generate . BastionsClient
type BastionsClient interface {
	GetBastionIPList(ctx context.Context, cluster *capg.GCPCluster) ([]string, error)
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

	bastionIPList, err := r.bastionsClient.GetBastionIPList(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	for i, bastionIP := range bastionIPList {
		bastionDomain := fmt.Sprintf("%s.%s.%s.", EndpointBastion(i+1), cluster.Name, r.baseDomain)
		logger := logger.WithValues("record", bastionDomain)
		logger.Info("Registering record")

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
			err = r.updateBastionRecordIfNotUptoDate(ctx, cluster, record, logger)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}
		logger.Info("Done Registering record", "ip", bastionIP)
	}

	return nil
}

func (r *Bastion) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	recordList, err := r.dnsService.ResourceRecordSets.List(cluster.Spec.Project, cluster.Name).
		Context(ctx).
		Do()

	if err != nil {
		return microerror.Mask(err)
	}

	for _, record := range recordList.Rrsets {
		// remove all bastion dns records
		if strings.HasPrefix(record.Name, "bastion") {
			logger := logger.WithValues("record", record.Name)
			logger.Info("Unregistering record")

			_, err = r.dnsService.ResourceRecordSets.Delete(cluster.Spec.Project, cluster.Name, record.Name, RecordA).
				Context(ctx).
				Do()

			if hasHttpCode(err, http.StatusNotFound) {
				logger.Info("Skipping. Record already unregistered")
				continue
			}
			if err != nil {
				return microerror.Mask(err)
			}
			logger.Info("Done unregistering record")
		}
	}
	return nil
}

func (r *Bastion) updateBastionRecordIfNotUptoDate(ctx context.Context, cluster *capg.GCPCluster, bastionRecord *clouddns.ResourceRecordSet, logger logr.Logger) error {
	bastionIP := bastionRecord.Rrdatas[0]
	// record exists, check if the IP matches
	rr, err := r.dnsService.ResourceRecordSets.Get(cluster.Spec.Project, cluster.Name, bastionRecord.Name, RecordA). //nolint:govet
																Context(ctx).
																Do()
	if err != nil {
		return microerror.Mask(err)
	}

	if len(rr.Rrdatas) > 0 || rr.Rrdatas[0] != bastionIP {
		logger.Info("Bastion record exists but its not up to date. Updating record")

		_, err = r.dnsService.ResourceRecordSets.Patch(cluster.Spec.Project, cluster.Name, bastionRecord.Name, RecordA, bastionRecord).
			Context(ctx).
			Do()
		if err != nil {
			return microerror.Mask(err)
		}
		logger.Info("Updated Bastion record", "ip", bastionIP)
	} else {
		logger.Info("Skipping. Record already exists and is up to date.")
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
