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

var _ = Describe("API Registrar", func() {
	var (
		ctx context.Context

		service      *clouddns.Service
		apiRegistrar *registrar.API

		cluster              *capg.GCPCluster
		clusterName          string
		domain               string
		apiDomain            string
		controlPlaneEndpoint string
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
		apiDomain = fmt.Sprintf("api.%s", domain)

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

		apiRegistrar = registrar.NewAPI(baseDomain, service)
	})

	AfterEach(func() {
		err := service.ManagedZones.Delete(gcpProject, cluster.Name).
			Context(context.Background()).
			Do()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Register", func() {
		var registErr error

		BeforeEach(func() {
			// Any private range IP address will do for the test.
			// Load Balancers in GCP have a external IP address.
			controlPlaneEndpoint = "10.0.0.1"
			cluster.Spec.ControlPlaneEndpoint.Host = controlPlaneEndpoint
		})

		JustBeforeEach(func() {
			registErr = apiRegistrar.Register(ctx, cluster)
		})

		AfterEach(func() {
			_, err := service.ResourceRecordSets.Delete(gcpProject, clusterName, apiDomain, registrar.RecordA).Do()
			Expect(err).To(Or(Not(HaveOccurred()), BeGoogleAPIErrorWithStatus(http.StatusNotFound)))
		})

		It("creates the A record", func() {
			Expect(registErr).NotTo(HaveOccurred())

			record, err := service.ResourceRecordSets.Get(gcpProject, clusterName, apiDomain, registrar.RecordA).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(record.Rrdatas).To(ConsistOf(controlPlaneEndpoint))
		})

		When("the cluster does not have a control plane endpoint yet", func() {
			BeforeEach(func() {
				cluster.Spec.ControlPlaneEndpoint.Host = ""
			})

			It("does not create an A record", func() {
				Expect(registErr).NotTo(HaveOccurred())

				_, err := service.ResourceRecordSets.Get(gcpProject, clusterName, apiDomain, registrar.RecordA).Do()
				Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
			})
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := apiRegistrar.Register(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the record already exists", func() {
			It("returns an error", func() {
				err := apiRegistrar.Register(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Unregister", func() {
		var unregistErr error

		BeforeEach(func() {
			err := apiRegistrar.Register(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			unregistErr = apiRegistrar.Unregister(ctx, cluster)
		})

		It("deletes the A record", func() {
			Expect(unregistErr).NotTo(HaveOccurred())

			_, err := service.ResourceRecordSets.Get(gcpProject, clusterName, apiDomain, registrar.RecordA).Do()
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := apiRegistrar.Unregister(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the record no longer exists", func() {
			It("returns an error", func() {
				err := apiRegistrar.Unregister(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
