package registrar

import (
	"context"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
	clouddns "google.golang.org/api/dns/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
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
	domain := r.getClusterDomain(cluster)
	zone, err := r.createManagedZone(ctx, domain, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	return r.registerNSInParentZone(ctx, domain, zone)
}

func (r *Zone) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	domain := r.getClusterDomain(cluster)

	_, err := r.dnsService.ResourceRecordSets.Delete(r.parentGCPProject, r.parentDNSZone, domain, RecordNS).
		Context(ctx).
		Do()

	if err != nil && !hasHttpCode(err, http.StatusNotFound) {
		return microerror.Mask(err)
	}

	err = r.dnsService.ManagedZones.Delete(cluster.Spec.Project, cluster.Name).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		return nil
	}
	return microerror.Mask(err)
}

func (r *Zone) registerNSInParentZone(ctx context.Context, domain string, zone *clouddns.ManagedZone) error {
	nsRecord := &clouddns.ResourceRecordSet{
		Name:    domain,
		Rrdatas: zone.NameServers,
		Type:    RecordNS,
	}
	_, err := r.dnsService.ResourceRecordSets.Create(r.parentGCPProject, r.parentDNSZone, nsRecord).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		return nil
	}

	return microerror.Mask(err)
}

func (r *Zone) createManagedZone(ctx context.Context, domain string, cluster *capg.GCPCluster) (*clouddns.ManagedZone, error) {
	zone := &clouddns.ManagedZone{
		Name:        cluster.Name,
		DnsName:     domain,
		Description: "DNS zone for WC cluster, managed by GCP DNS operator.",
		Visibility:  "public",
	}
	zone, err := r.dnsService.ManagedZones.Create(cluster.Spec.Project, zone).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
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
