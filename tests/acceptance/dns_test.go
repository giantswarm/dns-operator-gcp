package acceptance_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/giantswarm/dns-operator-gcp/pkg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

var _ = Describe("DNS", func() {
	var (
		ctx      context.Context
		resolver *net.Resolver

		clusterName   string
		clusterDomain string
		apiDomain     string
		cluster       *capi.Cluster
		gcpCluster    *capg.GCPCluster
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusterName = generateGUID("test")
		clusterDomain = fmt.Sprintf("%s.%s.", clusterName, baseDomain)
		apiDomain = fmt.Sprintf("%s.%s", dns.EndpointAPI, clusterDomain)

		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(10000),
				}
				return d.DialContext(ctx, "udp", "8.8.8.8:53")
			},
		}

		cluster = &capi.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: capi.ClusterSpec{
				InfrastructureRef: &v1.ObjectReference{
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
	})

	It("creates an NS record for the cluster", func() {
		var records []*net.NS
		Eventually(func() error {
			var err error
			records, err = resolver.LookupNS(ctx, clusterDomain)
			return err
		}, "1m", "500ms").Should(Succeed())

		Expect(records).ToNot(BeEmpty())
	})

	It("creates an A record for the kube api", func() {
		var records []net.IP
		Eventually(func() error {
			var err error
			records, err = resolver.LookupIP(ctx, "ip", apiDomain)
			return err
		}, "1m", "500ms").Should(Succeed())

		Expect(records).To(HaveLen(1))
		Expect(records[0].String()).To(Equal("10.0.0.1"))
	})

	When("the cluster is deleted", func() {
		BeforeEach(func() {
			Eventually(func() error {
				_, err := resolver.LookupNS(ctx, clusterDomain)
				return err
			}, "1m", "500ms").Should(Succeed())

			Expect(k8sClient.Delete(ctx, gcpCluster)).To(Succeed())
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
		})

		It("removes the ns record", func() {
			Eventually(func() error {
				_, err := resolver.LookupNS(ctx, clusterDomain)
				return err
			}, "1m", "500ms").ShouldNot(Succeed())
		})

		It("removes the A record", func() {
			Eventually(func() error {
				var err error
				_, err = resolver.LookupIP(ctx, "ip", apiDomain)
				return err
			}, "1m", "500ms").ShouldNot(Succeed())
		})
	})
})
