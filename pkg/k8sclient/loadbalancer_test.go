package k8sclient_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/dns-operator-gcp/pkg/k8sclient"
)

var _ = Describe("Service", func() {
	var (
		ctx          context.Context
		service      *corev1.Service
		loadBalancer *k8sclient.LoadBalancer
	)

	BeforeEach(func() {
		ctx = context.Background()
		loadBalancer = k8sclient.NewLoadBalancer(namespace, k8sClient)
	})

	Describe("GetIPByLabel", func() {
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

			otherService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-service",
					Namespace: namespace,
					Labels: map[string]string{
						"some-label": "true",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{Port: 8080},
					},
				},
			}
			Expect(k8sClient.Create(ctx, otherService)).To(Succeed())

			patchedService := service.DeepCopy()
			patchedService.Status = corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP: "10.0.0.1",
						},
					},
				},
			}
			Expect(k8sClient.Status().Patch(ctx, patchedService, client.MergeFrom(service))).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, service)).To(Succeed())
		})

		It("gets the service", func() {
			ip, err := loadBalancer.GetIPByLabel(ctx, "some-label", "true")
			Expect(err).NotTo(HaveOccurred())
			Expect(ip).To(Equal("10.0.0.1"))
		})

		When("the service doesn't have an IP yet", func() {
			BeforeEach(func() {
				nsName := types.NamespacedName{Name: service.Name, Namespace: service.Namespace}
				Expect(k8sClient.Get(ctx, nsName, service)).To(Succeed())

				patchedService := service.DeepCopy()
				patchedService.Status = corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{},
					},
				}
				Expect(k8sClient.Status().Patch(ctx, patchedService, client.MergeFrom(service))).To(Succeed())
			})

			It("returns an error", func() {
				ip, err := loadBalancer.GetIPByLabel(ctx, "some-label", "true")
				Expect(err).To(MatchError(And(
					ContainSubstring(namespace),
					ContainSubstring("ingress-service"),
					ContainSubstring("has 0 LoadBalancer ingresses, expected 1"),
				)))
				Expect(ip).To(Equal(""))
			})
		})

		When("the service has more than one IP", func() {
			BeforeEach(func() {
				nsName := types.NamespacedName{Name: service.Name, Namespace: service.Namespace}
				Expect(k8sClient.Get(ctx, nsName, service)).To(Succeed())

				patchedService := service.DeepCopy()
				patchedService.Status = corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{
							{IP: "10.0.0.1"},
							{IP: "10.0.0.2"},
						},
					},
				}
				Expect(k8sClient.Status().Patch(ctx, patchedService, client.MergeFrom(service))).To(Succeed())
			})

			It("returns an error", func() {
				ip, err := loadBalancer.GetIPByLabel(ctx, "some-label", "true")
				Expect(err).To(MatchError(ContainSubstring("has 2 LoadBalancer ingresses, expected 1")))
				Expect(ip).To(Equal(""))
			})
		})

		When("there is more than one LoadBalancer service matching the label", func() {
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
				ip, err := loadBalancer.GetIPByLabel(ctx, "some-label", "true")
				Expect(err).To(MatchError(ContainSubstring(`found 2 LoadBalancer services matching label "some-label": "true", expected 1`)))
				Expect(ip).To(BeZero())
			})
		})

		When("there is no service matching the label key", func() {
			It("returns an error", func() {
				ip, err := loadBalancer.GetIPByLabel(ctx, "some-other-label", "true")
				Expect(err).To(MatchError(ContainSubstring(`found 0 LoadBalancer services matching label "some-other-label": "true", expected 1`)))
				Expect(ip).To(BeZero())
			})
		})

		When("there is no service matching the label value", func() {
			It("returns an error", func() {
				ip, err := loadBalancer.GetIPByLabel(ctx, "some-label", "false")
				Expect(err).To(HaveOccurred())
				Expect(ip).To(BeZero())
			})
		})

		When("the context has expired", func() {
			It("returns an error", func() {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()

				ip, err := loadBalancer.GetIPByLabel(canceledCtx, "some-label", "false")
				Expect(err).To(HaveOccurred())
				Expect(ip).To(BeZero())
			})
		})
	})
})
