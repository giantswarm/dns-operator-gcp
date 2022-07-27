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

var _ = Describe("Wildcard Registrar", func() {
	var (
		ctx context.Context

		service           *clouddns.Service
		wildcardRegistrar *registrar.Wildcard

		cluster        *capg.GCPCluster
		clusterName    string
		domain         string
		wildcardDomain string
		ingressDomain  string
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
		wildcardDomain = fmt.Sprintf("*.%s", domain)
		ingressDomain = fmt.Sprintf("ingress.%s", domain)

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

		wildcardRegistrar = registrar.NewWildcard(baseDomain, service)
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
			registErr = wildcardRegistrar.Register(ctx, cluster)
		})

		AfterEach(func() {
			_, err := service.ResourceRecordSets.Delete(gcpProject, clusterName, wildcardDomain, registrar.RecordCNAME).Do()
			Expect(err).To(Or(Not(HaveOccurred()), BeGoogleAPIErrorWithStatus(http.StatusNotFound)))
		})

		It("creates the CNAME record", func() {
			Expect(registErr).NotTo(HaveOccurred())

			record, err := service.ResourceRecordSets.Get(gcpProject, clusterName, wildcardDomain, registrar.RecordCNAME).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(record.Rrdatas).To(ConsistOf(ingressDomain))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := wildcardRegistrar.Register(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the record already exists", func() {
			It("returns an error", func() {
				err := wildcardRegistrar.Register(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Unregister", func() {
		var unregistErr error

		BeforeEach(func() {
			err := wildcardRegistrar.Register(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			unregistErr = wildcardRegistrar.Unregister(ctx, cluster)
		})

		It("deletes the CNAME record", func() {
			Expect(unregistErr).NotTo(HaveOccurred())

			_, err := service.ResourceRecordSets.Get(gcpProject, clusterName, wildcardDomain, registrar.RecordCNAME).Do()
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := wildcardRegistrar.Unregister(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the record no longer exists", func() {
			It("returns an error", func() {
				err := wildcardRegistrar.Unregister(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
