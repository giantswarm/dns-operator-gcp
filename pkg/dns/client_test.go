package dns_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/giantswarm/dns-operator-gcp/pkg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clouddns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
)

var _ = Describe("Client", func() {
	var (
		ctx context.Context

		service *clouddns.Service
		client  *dns.Client

		cluster     *capg.GCPCluster
		clusterName string
	)

	BeforeEach(func() {
		ctx = context.Background()

		var err error
		service, err = clouddns.NewService(context.Background())
		Expect(err).NotTo(HaveOccurred())

		clusterName = generateGUID("test")
		cluster = &capg.GCPCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterName,
			},
			Spec: capg.GCPClusterSpec{
				Project: gcpProject,
			},
		}
		client = dns.NewClient(baseDomain, service)
	})

	Describe("CreateZone", func() {
		var createErr error

		JustBeforeEach(func() {
			createErr = client.CreateZone(ctx, cluster)
		})

		AfterEach(func() {
			err := service.ManagedZones.Delete(gcpProject, clusterName).Do()
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error", func() {
			Expect(createErr).NotTo(HaveOccurred())
		})

		It("creates a dns zone for the cluster", func() {
			actualZone, err := service.ManagedZones.Get(gcpProject, clusterName).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualZone.Name).To(Equal(cluster.Name))
			Expect(actualZone.DnsName).To(Equal(fmt.Sprintf("%s.%s.", cluster.Name, baseDomain)))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := client.CreateZone(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the zone already exists", func() {
			It("does not return an error", func() {
				err := client.CreateZone(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeleteZone", func() {
		var deleteErr error

		BeforeEach(func() {
			err := client.CreateZone(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			deleteErr = client.DeleteZone(ctx, cluster)
		})

		It("does not return an error", func() {
			Expect(deleteErr).NotTo(HaveOccurred())
		})

		It("deletes the dns zone", func() {
			actualZone, err := service.ManagedZones.Get(gcpProject, clusterName).Do()

			var googleErr *googleapi.Error
			Expect(errors.As(err, &googleErr)).To(BeTrue())
			Expect(googleErr.Code).To(Equal(http.StatusNotFound))
			Expect(actualZone).To(BeNil())
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := client.DeleteZone(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the zone already exists", func() {
			It("does not return an error", func() {
				err := client.DeleteZone(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
