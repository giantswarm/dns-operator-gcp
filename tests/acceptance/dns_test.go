package acceptance_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/dns-operator-gcp/pkg/k8sclient"
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
		bastionDomain string
		ingressDomain string
		cluster       *capi.Cluster
		gcpCluster    *capg.GCPCluster
		machine       *capg.GCPMachine
	)

	BeforeEach(func() {
		SetDefaultEventuallyPollingInterval(time.Second)
		SetDefaultEventuallyTimeout(time.Second * 90)

		ctx = context.Background()
		clusterName = tests.GenerateGUID("test")
		clusterDomain = fmt.Sprintf("%s.%s.", clusterName, baseDomain)
		apiDomain = fmt.Sprintf("%s.%s", registrar.EndpointAPI, clusterDomain)
		bastionDomain = fmt.Sprintf("%s.%s", registrar.EndpointBastion(1), clusterDomain)
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

		machine = &capg.GCPMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-bastion-1",
				Namespace: namespace,
				Labels: map[string]string{
					k8sclient.LabelBastionKey: k8sclient.BastionLabel(clusterName),
				},
			},
		}
		Expect(k8sClient.Create(ctx, machine)).To(Succeed())

		patchedMachine := machine.DeepCopy()
		patchedMachine.Status = capg.GCPMachineStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    "ExternalIP",
					Address: "1.2.3.4",
				},
			},
		}
		Expect(k8sClient.Status().Patch(ctx, patchedMachine, client.MergeFrom(machine))).To(Succeed())

	})

	It("creates the cluster DNS records", func() {
		By("creating an NS record for the cluster domain")
		var nsRecords []*net.NS
		Eventually(func() error {
			var err error
			nsRecords, err = resolver.LookupNS(ctx, clusterDomain)
			return err
		}).Should(Succeed())

		Expect(nsRecords).ToNot(BeEmpty())

		By("creating an A record for the kube api")
		var records []net.IP
		Eventually(func() error {
			var err error
			records, err = resolver.LookupIP(ctx, "ip", apiDomain)
			return err
		}).Should(Succeed())

		Expect(records).To(HaveLen(1))
		Expect(records[0].String()).To(Equal("10.0.0.1"))

		By("creating an A record for the bastion1")
		Eventually(func() error {
			var err error
			records, err = resolver.LookupIP(ctx, "ip", bastionDomain)
			return err
		}).Should(Succeed())

		Expect(records).To(HaveLen(1))
		Expect(records[0].String()).To(Equal("1.2.3.4"))

		By("updating an A record for the bastion1")
		// update bastion machine with different IP
		patchedMachine := machine.DeepCopy()
		patchedMachine.Status.Addresses = []corev1.NodeAddress{
			{
				Type:    "ExternalIP",
				Address: "1.2.3.5",
			},
		}
		Expect(k8sClient.Status().Patch(ctx, patchedMachine, client.MergeFrom(machine))).To(Succeed())

		// trigger reconciliation loop by update to speed up tests
		patchedCluster := gcpCluster.DeepCopy()
		patchedCluster.Labels = map[string]string{
			"extra": "label",
		}
		Expect(k8sClient.Status().Patch(ctx, patchedCluster, client.MergeFrom(gcpCluster))).To(Succeed())

		Eventually(func(g Gomega) (string, error) {
			var err error
			records, err = resolver.LookupIP(ctx, "ip", bastionDomain)
			if err != nil {
				return "", err
			}

			g.Expect(records).To(HaveLen(1))
			return records[0].String(), nil
		}).Should(Equal("1.2.3.5"))

		By("creating a CNAME record for the wildcard domain")
		wildcardDomain := fmt.Sprintf("%s.%s", uuid.NewString(), clusterDomain)
		var dnsResponse *dns.Msg
		Eventually(func() error {
			var err error
			dnsMessage := new(dns.Msg)
			dnsMessage.SetQuestion(wildcardDomain, dns.TypeCNAME)
			dnsMessage.RecursionDesired = true

			dnsClient := new(dns.Client)
			dnsResponse, _, err = dnsClient.Exchange(dnsMessage, "8.8.8.8:53")

			return err
		}).Should(Succeed())

		Expect(dnsResponse.Answer[0].(*dns.CNAME).Target).To(Equal(ingressDomain))
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

		It("creates the cluster DNS records", func() {
			By("not preventng the cluster deletion")
			nsName := types.NamespacedName{
				Name:      gcpCluster.Name,
				Namespace: gcpCluster.Namespace,
			}

			Eventually(func() error {
				return k8sClient.Get(ctx, nsName, &capg.GCPCluster{})
			}).ShouldNot(Succeed())

			By("removing the ns record")
			Eventually(func() error {
				_, err := resolver.LookupNS(ctx, clusterDomain)
				return err
			}).ShouldNot(Succeed())

			By("removing the api A record")
			Eventually(func() error {
				var err error
				_, err = resolver.LookupIP(ctx, "ip", apiDomain)
				return err
			}).ShouldNot(Succeed())

			By("removing the ingress A record")
			Eventually(func() error {
				var err error
				_, err = resolver.LookupIP(ctx, "ip", ingressDomain)
				return err
			}).ShouldNot(Succeed())
		})
	})
})
