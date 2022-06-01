package registrar_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clouddns "google.golang.org/api/dns/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/giantswarm/dns-operator-gcp/pkg/registrar"
	"github.com/giantswarm/dns-operator-gcp/pkg/registrar/registrarfakes"
	"github.com/giantswarm/dns-operator-gcp/tests"
)

var _ = Describe("API Registrar", func() {
	var (
		ctx context.Context

		clusterName   string
		domain        string
		ingressDomain string
		ingressIP     string

		service corev1.Service
		cluster *capg.GCPCluster

		serviceClient    *registrarfakes.FakeServiceClient
		dnsClient        *clouddns.Service
		ingressRegistrar *registrar.Ingress
	)

	BeforeEach(func() {
		ctx = context.Background()

		serviceClient = new(registrarfakes.FakeServiceClient)

		var err error
		dnsClient, err = clouddns.NewService(context.Background())
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
		ingressDomain = fmt.Sprintf("ingress.%s", domain)

		zone := &clouddns.ManagedZone{
			Name:        cluster.Name,
			DnsName:     domain,
			Description: "zone created for integration test",
			Visibility:  "public",
		}
		_, err = dnsClient.ManagedZones.Create(gcpProject, zone).
			Context(context.Background()).
			Do()
		Expect(err).NotTo(HaveOccurred())

		ingressIP = "10.0.0.1"
		ingressName := fmt.Sprintf("%s-ingress", clusterName)

		service = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressName,
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP: ingressIP,
						},
					},
				},
			},
		}
		serviceClient.GetByLabelReturns(service, nil)

		ingressRegistrar = registrar.NewIngress(baseDomain, dnsClient, serviceClient)
	})

	AfterEach(func() {
		err := dnsClient.ManagedZones.Delete(gcpProject, cluster.Name).
			Context(context.Background()).
			Do()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Register", func() {
		var registErr error

		JustBeforeEach(func() {
			registErr = ingressRegistrar.Register(ctx, cluster)
		})

		AfterEach(func() {
			_, err := dnsClient.ResourceRecordSets.Delete(gcpProject, clusterName, ingressDomain, registrar.RecordA).Do()
			Expect(err).To(Or(Not(HaveOccurred()), BeGoogleAPIErrorWithStatus(http.StatusNotFound)))
		})

		It("creates the A record", func() {
			Expect(registErr).NotTo(HaveOccurred())

			record, err := dnsClient.ResourceRecordSets.Get(gcpProject, clusterName, ingressDomain, registrar.RecordA).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(record.Rrdatas).To(ConsistOf(ingressIP))
		})

		When("the record already exists", func() {
			It("does not return an error", func() {
				err := ingressRegistrar.Register(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("the service is not LoadBalancer type", func() {
			BeforeEach(func() {
				service.Spec.Type = corev1.ServiceTypeClusterIP
				serviceClient.GetByLabelReturns(service, nil)
			})

			It("returns an error", func() {
				Expect(registErr).To(MatchError(ContainSubstring("found ClusterIP Service, expected type LoadBalancer")))
			})
		})

		When("the service does not have an ingress IP yet", func() {
			BeforeEach(func() {
				service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{}
				serviceClient.GetByLabelReturns(service, nil)
			})

			It("returns an error", func() {
				Expect(registErr).To(MatchError(ContainSubstring("found 0 LoadBalancer ingresses, expected 1")))
			})
		})

		When("the service has more than one ingress IP", func() {
			BeforeEach(func() {
				service.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
					{IP: ingressIP},
					{IP: "10.1.0.1"},
				}
				serviceClient.GetByLabelReturns(service, nil)
			})

			It("returns an error", func() {
				Expect(registErr).To(MatchError(ContainSubstring("found 2 LoadBalancer ingresses, expected 1")))
			})
		})

		When("getting the service returns an error", func() {
			BeforeEach(func() {
				serviceClient.GetByLabelReturns(corev1.Service{}, errors.New("boom"))
			})

			It("returns an error", func() {
				Expect(registErr).To(MatchError(ContainSubstring("boom")))
			})
		})
	})

	Describe("Unregister", func() {
		var unregistErr error

		BeforeEach(func() {
			err := ingressRegistrar.Register(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			unregistErr = ingressRegistrar.Unregister(ctx, cluster)
		})

		It("deletes the A record", func() {
			Expect(unregistErr).NotTo(HaveOccurred())

			_, err := dnsClient.ResourceRecordSets.Get(gcpProject, clusterName, ingressDomain, registrar.RecordA).Do()
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := ingressRegistrar.Unregister(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the record no longer exists", func() {
			It("returns an error", func() {
				err := ingressRegistrar.Unregister(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
