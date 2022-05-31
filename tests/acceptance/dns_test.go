package acceptance_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/dns-operator-gcp/pkg/registrar"
	"github.com/giantswarm/dns-operator-gcp/tests"
)

var _ = Describe("DNS", func() {
	var (
		ctx      context.Context
		resolver *net.Resolver

		clusterName   string
		clusterDomain string
		apiDomain     string
		ingressDomain string
		cluster       *capi.Cluster
		gcpCluster    *capg.GCPCluster
	)

	BeforeEach(func() {
		SetDefaultEventuallyPollingInterval(time.Second)
		SetDefaultEventuallyTimeout(time.Second * 90)

		ctx = context.Background()
		clusterName = tests.GenerateGUID("test")
		clusterDomain = fmt.Sprintf("%s.%s.", clusterName, baseDomain)
		apiDomain = fmt.Sprintf("%s.%s", registrar.EndpointAPI, clusterDomain)
		ingressDomain = fmt.Sprintf("%s.%s", registrar.EndpointIngress, clusterDomain)

		resolver = &net.Resolver{
			PreferGo:     true,
			StrictErrors: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(10000),
				}
				return d.DialContext(ctx, "udp", "8.8.4.4:53")
			},
		}

		cluster = &capi.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: capi.ClusterSpec{
				InfrastructureRef: &corev1.ObjectReference{
					APIVersion: capg.GroupVersion.String(),
					Kind:       "GCPCluster",
					Name:       clusterName,
					Namespace:  namespace,
				},
			},
		}
		Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

		gcpCluster = &capg.GCPCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: capg.GCPClusterSpec{
				Project: gcpProject,
				ControlPlaneEndpoint: capi.APIEndpoint{
					Host: "10.0.0.1",
				},
			},
		}
		Expect(k8sClient.Create(ctx, gcpCluster)).To(Succeed())

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-service",
				Namespace: registrar.IngressNamespace,
				Labels: map[string]string{
					registrar.LabelIngressKey: registrar.LabelIngressValue,
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 8080},
				},
			},
		}

		patchedService := service.DeepCopy()
		patchedService.Status = corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{
					{
						IP: "10.0.0.2",
					},
				},
			},
		}

		err := k8sClient.Status().Patch(ctx, patchedService, client.MergeFrom(service))
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates an NS record for the cluster", func() {
		var records []*net.NS
		Eventually(func() error {
			var err error
			records, err = resolver.LookupNS(ctx, clusterDomain)
			return err
		}).Should(Succeed())

		Expect(records).ToNot(BeEmpty())
	})

	It("creates an A record for the kube api", func() {
		var records []net.IP
		Eventually(func() error {
			var err error
			records, err = resolver.LookupIP(ctx, "ip", apiDomain)
			return err
		}).Should(Succeed())

		Expect(records).To(HaveLen(1))
		Expect(records[0].String()).To(Equal("10.0.0.1"))
	})

	It("creates an A record for the ingress", func() {
		var records []net.IP
		Eventually(func() error {
			var err error
			records, err = resolver.LookupIP(ctx, "ip", ingressDomain)
			return err
		}).Should(Succeed())

		Expect(records).To(HaveLen(1))
		Expect(records[0].String()).To(Equal("10.0.0.2"))
	})

	It("creates a CNAME record for the wildcard domain", func() {
		wildcardDomain := fmt.Sprintf("%s.%s", uuid.NewString(), clusterDomain)
		var record string
		Eventually(func() error {
			var err error
			record, err = resolver.LookupCNAME(ctx, wildcardDomain)
			return err
		}).Should(Succeed())

		Expect(record).To(Equal(ingressDomain))

		var records []string
		Eventually(func() error {
			var err error
			records, err = resolver.LookupHost(ctx, wildcardDomain)
			return err
		}).Should(Succeed())

		Expect(records).To(ConsistOf("10.0.0.2"))
	})

	When("the cluster is deleted", func() {
		BeforeEach(func() {
			Eventually(func() error {
				_, err := resolver.LookupNS(ctx, clusterDomain)
				return err
			}).Should(Succeed())

			Expect(k8sClient.Delete(ctx, gcpCluster)).To(Succeed())
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
		})

		It("does not prevent the cluster deletion", func() {
			nsName := types.NamespacedName{
				Name:      gcpCluster.Name,
				Namespace: gcpCluster.Namespace,
			}

			Eventually(func() error {
				return k8sClient.Get(ctx, nsName, &capg.GCPCluster{})
			}).ShouldNot(Succeed())
		})

		It("removes the ns record", func() {
			Eventually(func() error {
				_, err := resolver.LookupNS(ctx, clusterDomain)
				return err
			}).ShouldNot(Succeed())
		})

		It("removes the api A record", func() {
			Eventually(func() error {
				var err error
				_, err = resolver.LookupIP(ctx, "ip", apiDomain)
				return err
			}).ShouldNot(Succeed())
		})

		It("removes the ingress A record", func() {
			Eventually(func() error {
				var err error
				_, err = resolver.LookupIP(ctx, "ip", ingressDomain)
				return err
			}).ShouldNot(Succeed())
		})
	})
})
