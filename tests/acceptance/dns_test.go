package acceptance_test

import (
	"context"
	"fmt"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

var _ = Describe("DNS", func() {
	var (
		ctx context.Context

		clusterName string
		host        string
		cluster     *capi.Cluster
		gcpCluster  *capg.GCPCluster
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusterName = generateGUID("test")
		host = fmt.Sprintf("%s.%s.", clusterName, baseDomain)

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
			},
		}
		Expect(k8sClient.Create(ctx, gcpCluster)).To(Succeed())
	})

	It("creates an NS record for the cluster", func() {
		var records []*net.NS
		Eventually(func() error {
			var err error
			records, err = net.LookupNS(host)
			return err
		}, "1m", "500ms").Should(Succeed())

		Expect(records).ToNot(BeEmpty())
	})

	When("the cluster is deleted", func() {
		BeforeEach(func() {
			Expect(k8sClient.Delete(ctx, gcpCluster)).To(Succeed())
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
		})

		It("removes the dns record", func() {
			Eventually(func() error {
				_, err := net.LookupNS(host)
				return err
			}, "1m", "500ms").ShouldNot(Succeed())

			Consistently(func() error {
				_, err := net.LookupNS(host)
				return err
			}, "15s", "500ms").ShouldNot(Succeed())
		})
	})
})
