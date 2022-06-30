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

type Zone struct {
	dnsService *clouddns.Service

	baseDomain       string
	parentDNSZone    string
	parentGCPProject string
}

func NewZone(baseDomain, parentDNSZone, parentGCPProject string, dnsService *clouddns.Service) *Zone {
	return &Zone{
		baseDomain:       baseDomain,
		parentDNSZone:    parentDNSZone,
		parentGCPProject: parentGCPProject,
		dnsService:       dnsService,
	}
}

func (r *Zone) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Registering record")
	defer logger.Info("Done registering record")

	domain := r.getClusterDomain(cluster)
	zone, err := r.createManagedZone(ctx, logger, domain, cluster)
	if err != nil {
		logger.Error(err, "Failed to register managed zone")
		return microerror.Mask(err)
	}

	return r.registerNSInParentZone(ctx, logger, domain, zone)
}

func (r *Zone) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	logger := r.getLogger(ctx)

	logger.Info("Registering record")
	defer logger.Info("Done registering record")

	domain := r.getClusterDomain(cluster)

	_, err := r.dnsService.ResourceRecordSets.Delete(r.parentGCPProject, r.parentDNSZone, domain, RecordNS).
		Context(ctx).
		Do()

	if err != nil && !hasHttpCode(err, http.StatusNotFound) {
		logger.Info("Skipping. Record already unregistered")
		return microerror.Mask(err)
	}

	err = r.dnsService.ManagedZones.Delete(cluster.Spec.Project, cluster.Name).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		logger.Info("Zone already deleted")
		return nil
	}
	return microerror.Mask(err)
}

func (r *Zone) registerNSInParentZone(ctx context.Context, logger logr.Logger, domain string, zone *clouddns.ManagedZone) error {
	nsRecord := &clouddns.ResourceRecordSet{
		Name:    domain,
		Rrdatas: zone.NameServers,
		Type:    RecordNS,
	}
	_, err := r.dnsService.ResourceRecordSets.Create(r.parentGCPProject, r.parentDNSZone, nsRecord).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		logger.Info("Skipping. Record already exists")
		return nil
	}

	return microerror.Mask(err)
}

func (r *Zone) createManagedZone(ctx context.Context, logger logr.Logger, domain string, cluster *capg.GCPCluster) (*clouddns.ManagedZone, error) {
	zone := &clouddns.ManagedZone{
		Name:        cluster.Name,
		DnsName:     domain,
		Description: "DNS zone for cluster, managed by GCP DNS operator.",
		Visibility:  "public",
	}
	zone, err := r.dnsService.ManagedZones.Create(cluster.Spec.Project, zone).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		logger.Info("Getting existing zone")
		return r.getManagedZone(ctx, cluster)
	}

	if err != nil {
		return nil, err
	}

	return zone, err
}

func (r *Zone) getManagedZone(ctx context.Context, cluster *capg.GCPCluster) (*clouddns.ManagedZone, error) {
	return r.dnsService.ManagedZones.Get(cluster.Spec.Project, cluster.Name).
		Context(ctx).
		Do()
}

func (r *Zone) getClusterDomain(cluster *capg.GCPCluster) string {
	return fmt.Sprintf("%s.%s.", cluster.Name, r.baseDomain)
}

func (r *Zone) getLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)
	return logger.WithName("zone-registrar")
}
