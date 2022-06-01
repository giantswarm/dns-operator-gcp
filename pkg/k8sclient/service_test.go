package k8sclient_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/dns-operator-gcp/pkg/k8sclient"
)

var _ = Describe("Service", func() {
	var (
		ctx     context.Context
		service *corev1.Service
		client  *k8sclient.Service
	)

	BeforeEach(func() {
		ctx = context.Background()
		client = k8sclient.NewService(namespace, k8sClient)
	})

	Describe("GetByLabel", func() {
		BeforeEach(func() {
			service = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ingress-service",
					Namespace: namespace,
					Labels: map[string]string{
						"some-label": "true",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{
						{Port: 8080},
					},
				},
			}
			Expect(k8sClient.Create(ctx, service)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, service)).To(Succeed())
		})

		It("gets the service", func() {
			actualService, err := client.GetByLabel(ctx, "some-label", "true")
			Expect(err).NotTo(HaveOccurred())
			Expect(actualService.Name).To(Equal("ingress-service"))
		})

		When("there is more than one service matching the label", func() {
			var otherService *corev1.Service

			BeforeEach(func() {
				otherService = &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "another-ingress-service",
						Namespace: namespace,
						Labels: map[string]string{
							"some-label": "true",
						},
					},
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeLoadBalancer,
						Ports: []corev1.ServicePort{
							{Port: 8080},
						},
					},
				}
				Expect(k8sClient.Create(ctx, otherService)).To(Succeed())
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, otherService)).To(Succeed())
			})

			It("returns an error", func() {
				actualService, err := client.GetByLabel(ctx, "some-label", "true")
				Expect(err).To(MatchError(ContainSubstring(`found 2 services matching label "some-label": "true", expected 1`)))
				Expect(actualService).To(BeZero())
			})
		})

		When("there is no service matching the label key", func() {
			It("returns an error", func() {
				actualService, err := client.GetByLabel(ctx, "some-other-label", "true")
				Expect(err).To(MatchError(ContainSubstring(`found 0 services matching label "some-other-label": "true", expected 1`)))
				Expect(actualService).To(BeZero())
			})
		})

		When("there is no service matching the label value", func() {
			It("returns an error", func() {
				actualService, err := client.GetByLabel(ctx, "some-label", "false")
				Expect(err).To(HaveOccurred())
				Expect(actualService).To(BeZero())
			})
		})

		When("the context has expired", func() {
			It("returns an error", func() {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()

				actualService, err := client.GetByLabel(canceledCtx, "some-label", "false")
				Expect(err).To(HaveOccurred())
				Expect(actualService).To(BeZero())
			})
		})
	})
})
