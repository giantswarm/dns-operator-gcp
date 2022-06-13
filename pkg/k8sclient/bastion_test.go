package k8sclient_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/dns-operator-gcp/controllers"
	"github.com/giantswarm/dns-operator-gcp/pkg/k8sclient"
)

var _ = Describe("MachineList", func() {
	var (
		ctx      context.Context
		machine  *capg.GCPMachine
		bastions *k8sclient.Bastions
		cluster  *capg.GCPCluster
	)

	BeforeEach(func() {
		ctx = context.Background()
		bastions = k8sclient.NewBastions(k8sClient, controllers.FinalizerDNS)
		cluster = &capg.GCPCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: namespace,
			},
		}
	})

	Describe("GetBastionIPList", func() {
		BeforeEach(func() {
			machine = &capg.GCPMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster-bastion-1",
					Namespace: namespace,
					Labels: map[string]string{
						k8sclient.BastionLabelKey: k8sclient.BastionLabel("test-cluster"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, machine)).To(Succeed())

			otherMachine := &capg.GCPMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-cluster-bastion-1",
					Namespace: namespace,
					Labels: map[string]string{
						k8sclient.BastionLabelKey: k8sclient.BastionLabel("test-cluster-1"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, otherMachine)).To(Succeed())

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

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, machine)).To(Succeed())
		})

		It("gets the bastion machine", func() {
			ipList, err := bastions.GetBastionIPList(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(ipList).To(Equal([]string{"1.2.3.4"}))
		})

		When("the bastion doesn't have an IP yet", func() {
			BeforeEach(func() {
				nsName := types.NamespacedName{Name: machine.Name, Namespace: machine.Namespace}
				Expect(k8sClient.Get(ctx, nsName, machine)).To(Succeed())

				patchedMachine := machine.DeepCopy()
				patchedMachine.Status = capg.GCPMachineStatus{
					Addresses: []corev1.NodeAddress{},
				}
				Expect(k8sClient.Status().Patch(ctx, patchedMachine, client.MergeFrom(machine))).To(Succeed())
			})

			It("returns an error", func() {
				ipList, err := bastions.GetBastionIPList(ctx, cluster)
				Expect(err).To(MatchError(And(
					ContainSubstring("bastion IP is not yet available"),
				)))
				Expect(ipList).To(BeNil())
			})
		})

		When("there is more than one Bastion host", func() {
			var otherMachine *capg.GCPMachine

			BeforeEach(func() {
				otherMachine = &capg.GCPMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-bastion-1",
						Namespace: namespace,
						Labels: map[string]string{
							k8sclient.BastionLabelKey: k8sclient.BastionLabel("test-cluster"),
						},
					},
				}
				Expect(k8sClient.Create(ctx, otherMachine)).To(Succeed())
				patchedMachine := otherMachine.DeepCopy()
				patchedMachine.Status = capg.GCPMachineStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    "ExternalIP",
							Address: "1.2.3.5",
						},
					},
				}
				Expect(k8sClient.Status().Patch(ctx, patchedMachine, client.MergeFrom(otherMachine))).To(Succeed())

			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, otherMachine)).To(Succeed())
			})

			It("return multiple bastions ip", func() {
				ipList, err := bastions.GetBastionIPList(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
				Expect(ipList).To(Equal([]string{"1.2.3.4", "1.2.3.5"}))
			})
		})

		When("there is no bastion matching the label key", func() {
			It("returns an error", func() {

				otherCluster := &capg.GCPCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-other",
						Namespace: namespace,
					},
				}
				ipList, err := bastions.GetBastionIPList(ctx, otherCluster)
				Expect(err).To(MatchError(ContainSubstring(`bastion IP is not yet available`)))
				Expect(ipList).To(BeNil())
			})
		})

		When("the context has expired", func() {
			It("returns an error", func() {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()

				ip, err := bastions.GetBastionIPList(canceledCtx, cluster)
				Expect(err).To(HaveOccurred())
				Expect(ip).To(BeZero())
			})
		})
	})
})
