package dns_test

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	clouddns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"

	"github.com/giantswarm/dns-operator-gcp/pkg/dns"
	"github.com/giantswarm/dns-operator-gcp/tests"
)

var _ = Describe("Client", func() {
	var (
		ctx context.Context

		service *clouddns.Service
		client  *dns.Client

		cluster     *capg.GCPCluster
		clusterName string
		domain      string
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

		client = dns.NewClient(baseDomain, parentDNSZone, gcpProject, service)
	})

	Describe("CreateZone", func() {
		var createErr error

		JustBeforeEach(func() {
			createErr = client.CreateZone(ctx, cluster)
		})

		AfterEach(func() {
			_, err := service.ResourceRecordSets.Delete(gcpProject, parentDNSZone, domain, dns.RecordNS).Do()
			Expect(err).NotTo(HaveOccurred())

			err = service.ManagedZones.Delete(gcpProject, clusterName).Do()
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error", func() {
			Expect(createErr).NotTo(HaveOccurred())
		})

		It("creates a dns zone for the cluster and an NS record in the parent zone", func() {
			actualZone, err := service.ManagedZones.Get(gcpProject, clusterName).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualZone.Name).To(Equal(cluster.Name))
			Expect(actualZone.DnsName).To(Equal(domain))

			record, err := service.ResourceRecordSets.Get(gcpProject, parentDNSZone, domain, dns.RecordNS).Do()
			Expect(err).NotTo(HaveOccurred())
			Expect(record.Rrdatas).To(ConsistOf(actualZone.NameServers))
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := client.CreateZone(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the zone already exists", func() {
			It("does not return an error", func() {
				err := client.CreateZone(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("A Records", func() {
		var (
			apiDomain            string
			controlPlaneEndpoint string

			createErr error
		)

		BeforeEach(func() {
			// Any private range IP address will do for the test.
			// Load Balancers in GCP have a external IP address.
			controlPlaneEndpoint = "10.0.0.1"
			cluster.Spec.ControlPlaneEndpoint.Host = controlPlaneEndpoint
			apiDomain = fmt.Sprintf("api.%s.%s.", clusterName, baseDomain)

			err := client.CreateZone(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, err := service.ResourceRecordSets.Delete(gcpProject, parentDNSZone, domain, dns.RecordNS).Do()
			Expect(err).NotTo(HaveOccurred())

			err = service.ManagedZones.Delete(gcpProject, clusterName).Do()
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("CreateARecords", func() {
			JustBeforeEach(func() {
				createErr = client.CreateARecords(ctx, cluster)
			})

			AfterEach(func() {
				_, err := service.ResourceRecordSets.Delete(gcpProject, clusterName, apiDomain, dns.RecordA).Do()
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates the A record", func() {
				Expect(createErr).NotTo(HaveOccurred())

				record, err := service.ResourceRecordSets.Get(gcpProject, clusterName, apiDomain, dns.RecordA).Do()
				Expect(err).NotTo(HaveOccurred())
				Expect(record.Rrdatas).To(ConsistOf(controlPlaneEndpoint))
			})

			When("the context has been cancelled", func() {
				It("returns an error", func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()

					err := client.CreateARecords(ctx, cluster)
					Expect(err).To(MatchError(ContainSubstring("context canceled")))
				})
			})

			When("the record already exists", func() {
				It("returns an error", func() {
					err := client.CreateARecords(ctx, cluster)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Describe("DeleteARecords", func() {
			BeforeEach(func() {
				err := client.CreateARecords(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func() {
				createErr = client.DeleteARecords(ctx, cluster)
			})

			It("deletes the A record", func() {
				Expect(createErr).NotTo(HaveOccurred())

				_, err := service.ResourceRecordSets.Get(gcpProject, clusterName, apiDomain, dns.RecordA).Do()
				Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
			})

			When("the context has been cancelled", func() {
				It("returns an error", func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()

					err := client.DeleteARecords(ctx, cluster)
					Expect(err).To(MatchError(ContainSubstring("context canceled")))
				})
			})

			When("the record no longer exists", func() {
				It("returns an error", func() {
					err := client.DeleteARecords(ctx, cluster)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("DeleteZone", func() {
		var deleteErr error

		BeforeEach(func() {
			err := client.CreateZone(ctx, cluster)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			deleteErr = client.DeleteZone(ctx, cluster)
		})

		It("does not return an error", func() {
			Expect(deleteErr).NotTo(HaveOccurred())
		})

		It("deletes the dns zone and NS record", func() {
			actualZone, err := service.ManagedZones.Get(gcpProject, clusterName).Do()
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
			Expect(actualZone).To(BeNil())

			record, err := service.ResourceRecordSets.Get(gcpProject, parentDNSZone, domain, dns.RecordNS).Do()
			Expect(err).To(BeGoogleAPIErrorWithStatus(http.StatusNotFound))
			Expect(record).To(BeNil())
		})

		When("the context has been cancelled", func() {
			It("returns an error", func() {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()

				err := client.DeleteZone(ctx, cluster)
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		When("the zone does not exists", func() {
			It("return an error", func() {
				err := client.DeleteZone(ctx, cluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

type beGoogleAPIErrorWithStatusMatcher struct {
	expected int
}

func BeGoogleAPIErrorWithStatus(expected int) types.GomegaMatcher {
	return &beGoogleAPIErrorWithStatusMatcher{expected: expected}
}

func (m *beGoogleAPIErrorWithStatusMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return false, nil
	}

	actualError, isError := actual.(error)
	if !isError {
		return false, fmt.Errorf("%#v is not an error", actual)
	}

	matches, err := BeAssignableToTypeOf(actualError).Match(&googleapi.Error{})
	if err != nil || !matches {
		return false, err
	}

	googleAPIError, isGoogleAPIError := actual.(*googleapi.Error)
	if !isGoogleAPIError {
		return false, fmt.Errorf("%#v is not a google api error", actual)
	}
	return Equal(googleAPIError.Code).Match(m.expected)
}

func (m *beGoogleAPIErrorWithStatusMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf("to be a google api error with status code: %s", m.getExpectedStatusText()),
	)
}

func (m *beGoogleAPIErrorWithStatusMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf("to not be a google api error with status: %s", m.getExpectedStatusText()),
	)
}

func (m *beGoogleAPIErrorWithStatusMatcher) getExpectedStatusText() string {
	return fmt.Sprintf("%d %s", m.expected, http.StatusText(m.expected))
}
