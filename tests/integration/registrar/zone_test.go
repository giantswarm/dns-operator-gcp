package registrar_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clouddns "google.golang.org/api/dns/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/giantswarm/dns-operator-gcp/v2/pkg/registrar"
	"github.com/giantswarm/dns-operator-gcp/v2/tests"
)

var _ = Describe("Zone Registrar", func() {
	var (
		ctx context.Context

		service       *clouddns.Service
		zoneRegistrar *registrar.Zone

		cluster     *capg.GCPCluster
		clusterName string
		domain      string
	)

	BeforeEach(func() {
		ctx = context.Background()

		var err error
		service, err = clouddns.NewService(context.Background())
		Expect(err).NotTo(HaveOccurred())

		clusterName = tests.GenerateGUID("test")
		cluster = &capg.GCPCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterName,
			},
			Spec: capg.GCPClusterSpec{
				Project: gcpProject,
			},
		}
		domain = fmt.Sprintf("%s.%s.", cluster.Name, baseDomain)

		zoneRegistrar = registrar.NewZone(baseDomain, parentDNSZone, gcpProject, service)
	})

	Describe("Register", func() {
		var registErr error

		JustBeforeEach(func() {
			registErr = zoneRegistrar.Register(ctx, cluster)
		})

		AfterEach(func() {
			_, err := service.ResourceRecordSets.Delete(gcpProject, parentDNSZone, domain, registrar.RecordNS).Do()
			Expect(err).NotTo(HaveOccurred())

			err = service.ManagedZones.Delete(gcpProject, clusterName).Do()
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error", func() {
			Expect(registErr).NotTo(HaveOccurred())
		})

		It("creates a dns zone for the cluster and an NS record in the parent zone", func() {
			actualZone, err := service.ManagedZones.Get(gcpProject, clusterName).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualZone.Name).To(Equal(cluster.Name))
			Expect(actualZone.DnsName).To(Equal(domain))

			record, err := service.ResourceRecordSets.Get(gcpProject, parentDNSZone, domain, registrar.RecordNS).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(record.Rrdatas).To(ConsistOf(actualZone.NameServers))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := zoneRegistrar.Register(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the zone already exists", func() {
			It("does not return an error", func() {
				err := zoneRegistrar.Register(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Unregister", func() {
		var deleteErr error

		BeforeEach(func() {
			err := zoneRegistrar.Register(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			deleteErr = zoneRegistrar.Unregister(ctx, cluster)
		})

		It("does not return an error", func() {
			Expect(deleteErr).NotTo(HaveOccurred())
		})

		It("deletes the dns zone and NS record", func() {
			actualZone, err := service.ManagedZones.Get(gcpProject, clusterName).Do()
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
			Expect(actualZone).To(BeNil())

			record, err := service.ResourceRecordSets.Get(gcpProject, parentDNSZone, domain, registrar.RecordNS).Do()
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
			Expect(record).To(BeNil())
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := zoneRegistrar.Unregister(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the zone does not exists", func() {
			It("return an error", func() {
				err := zoneRegistrar.Unregister(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
