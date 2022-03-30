package dns

import (
	"context"

	clouddns "google.golang.org/api/dns/v2"
	"sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
)

type Client struct {
	dnsService *clouddns.Service

	baseDomain string
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) CreateZone(ctx context.Context, cluster *v1beta1.GCPCluster) error {
	return nil
}
