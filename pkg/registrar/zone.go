package registrar

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/giantswarm/microerror"
	clouddns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
)

const (
	RecordNS = "NS"
	RecordA  = "A"

	EndpointAPI = "api"
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

func (c *Zone) Register(ctx context.Context, cluster *capg.GCPCluster) error {
	domain := c.getClusterDomain(cluster)
	zone, err := c.createManagedZone(ctx, domain, cluster)
	if hasHttpCode(err, http.StatusConflict) {
		return nil
	}
	if err != nil {
		return microerror.Mask(err)
	}

	return c.registerNSInParentZone(ctx, domain, zone)
}

func (c *Zone) Unregister(ctx context.Context, cluster *capg.GCPCluster) error {
	domain := c.getClusterDomain(cluster)

	_, err := c.dnsService.ResourceRecordSets.Delete(c.parentGCPProject, c.parentDNSZone, domain, RecordNS).
		Context(ctx).
		Do()

	if err != nil && !hasHttpCode(err, http.StatusNotFound) {
		return microerror.Mask(err)
	}

	err = c.dnsService.ManagedZones.Delete(cluster.Spec.Project, cluster.Name).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		return nil
	}
	return microerror.Mask(err)
}

func (c *Zone) registerNSInParentZone(ctx context.Context, domain string, zone *clouddns.ManagedZone) error {
	nsRecord := &clouddns.ResourceRecordSet{
		Name:    domain,
		Rrdatas: zone.NameServers,
		Type:    RecordNS,
	}
	_, err := c.dnsService.ResourceRecordSets.Create(c.parentGCPProject, c.parentDNSZone, nsRecord).
		Context(ctx).
		Do()

	return microerror.Mask(err)
}

func (c *Zone) createManagedZone(ctx context.Context, domain string, cluster *capg.GCPCluster) (*clouddns.ManagedZone, error) {
	zone := &clouddns.ManagedZone{
		Name:        cluster.Name,
		DnsName:     domain,
		Description: "DNS zone for WC cluster, managed by GCP DNS operator.",
		Visibility:  "public",
	}
	zone, err := c.dnsService.ManagedZones.Create(cluster.Spec.Project, zone).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	return zone, err
}

func (c *Zone) getClusterDomain(cluster *capg.GCPCluster) string {
	return fmt.Sprintf("%s.%s.", cluster.Name, c.baseDomain)
}

func hasHttpCode(err error, statusCode int) bool {
	var googleErr *googleapi.Error
	if errors.As(err, &googleErr) {
		if googleErr.Code == statusCode {
			return true
		}
	}

	return false
}