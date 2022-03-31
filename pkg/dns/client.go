package dns

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	clouddns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
)

type Client struct {
	dnsService *clouddns.Service

	baseDomain string
	gcpProject string
}

func NewClient(gcpProject, baseDomain string, dnsService *clouddns.Service) *Client {
	return &Client{
		gcpProject: gcpProject,
		baseDomain: baseDomain,
		dnsService: dnsService,
	}
}

func (c *Client) CreateZone(ctx context.Context, cluster *v1beta1.GCPCluster) error {
	domain := fmt.Sprintf("%s.%s.", cluster.Name, c.baseDomain)
	zone := &clouddns.ManagedZone{
		Name:        cluster.Name,
		DnsName:     domain,
		Description: "DNS zone for WC cluster, managed by GCP DNS operator.",
		Visibility:  "public",
	}
	_, err := c.dnsService.ManagedZones.Create(c.gcpProject, zone).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusConflict) {
		return nil
	}

	return err
}

func (c *Client) DeleteZone(ctx context.Context, cluster *v1beta1.GCPCluster) error {
	err := c.dnsService.ManagedZones.Delete(c.gcpProject, cluster.Name).
		Context(ctx).
		Do()

	if hasHttpCode(err, http.StatusNotFound) {
		return nil
	}
	return err
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
