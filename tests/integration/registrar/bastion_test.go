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
	"github.com/giantswarm/dns-operator-gcp/v2/pkg/registrar/registrarfakes"
	"github.com/giantswarm/dns-operator-gcp/v2/tests"
)

var _ = Describe("Bastion Registrar", func() {
	var (
		ctx context.Context

		service          *clouddns.Service
		bastionRegistrar *registrar.Bastion

		bastionsClient *registrarfakes.FakeBastionsClient

		cluster       *capg.GCPCluster
		clusterName   string
		domain        string
		bastionDomain string
	)

	BeforeEach(func() {
		ctx = context.Background()

		var err error
		service, err = clouddns.NewService(context.Background())
		Expect(err).NotTo(HaveOccurred())

		bastionsClient = new(registrarfakes.FakeBastionsClient)

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
		bastionDomain = fmt.Sprintf("bastion1.%s", domain)

		zone := &clouddns.ManagedZone{
			Name:        cluster.Name,
			DnsName:     domain,
			Description: "zone created for integration test",
			Visibility:  "public",
		}
		_, err = service.ManagedZones.Create(gcpProject, zone).
			Context(context.Background()).
			Do()
		Expect(err).NotTo(HaveOccurred())

		bastionsClient.GetBastionIPListReturns([]string{"1.2.3.4"}, nil)

		bastionRegistrar = registrar.NewBastion(baseDomain, bastionsClient, service)
	})

	AfterEach(func() {
		err := service.ManagedZones.Delete(gcpProject, cluster.Name).
			Context(context.Background()).
			Do()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Register", func() {
		var registErr error

		JustBeforeEach(func() {
			registErr = bastionRegistrar.Register(ctx, cluster)
		})

		AfterEach(func() {
			_, err := service.ResourceRecordSets.Delete(gcpProject, clusterName, bastionDomain, registrar.RecordA).Do()
			Expect(err).To(Or(Not(HaveOccurred()), BeGoogleAPIErrorWithStatus(http.StatusNotFound)))
		})

		It("creates the bastion A record", func() {
			Expect(registErr).NotTo(HaveOccurred())

			record, err := service.ResourceRecordSets.Get(gcpProject, clusterName, bastionDomain, registrar.RecordA).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(record.Rrdatas).To(ConsistOf("1.2.3.4"))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := bastionRegistrar.Register(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the record already exists", func() {
			It("returns an error", func() {
				err := bastionRegistrar.Register(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Unregister", func() {
		var unregistErr error

		When("the zone is not registered", func() {
			JustBeforeEach(func() {
				unregistErr = bastionRegistrar.Unregister(ctx, cluster)
			})

			It("does not return an error", func() {
				err := bastionRegistrar.Unregister(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("the zone is registered", func() {
			BeforeEach(func() {
				err := bastionRegistrar.Register(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func() {
				unregistErr = bastionRegistrar.Unregister(ctx, cluster)
			})

			It("deletes the bastion A record", func() {
				Expect(unregistErr).NotTo(HaveOccurred())

				_, err := service.ResourceRecordSets.Get(gcpProject, clusterName, bastionDomain, registrar.RecordA).Do()
				Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
			})

			When("the context has been cancelled", func() {
				It("returns an error", func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()

					err := bastionRegistrar.Unregister(ctx, cluster)
					Expect(err).To(MatchError(ContainSubstring("context canceled")))
				})
			})

			When("the record no longer exists", func() {
				It("does not return an error", func() {
					err := bastionRegistrar.Unregister(ctx, cluster)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
