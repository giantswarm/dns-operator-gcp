package dns_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	clouddns "google.golang.org/api/dns/v1"
)

var _ = Describe("Client", func() {
	var service *clouddns.Service // client *dns.Client

	BeforeEach(func() {
		var err error
		service, err = clouddns.NewService(context.Background())
		Expect(err).NotTo(HaveOccurred())
	})

	It("does something", func() {
		zone := &clouddns.ManagedZone{
			Name:        "mario-test",
			DnsName:     "gmario.gtest.gigantic.io.",
			Description: "DNS zone for WC cluster, managed by GCP DNS operator.",
			Visibility:  "public",
		}
		zone, err := service.ManagedZones.Create("capi-test-phoenix", zone).Do()
		fmt.Printf("%#v\n", zone)
		Expect(err).NotTo(HaveOccurred())
	})
})
