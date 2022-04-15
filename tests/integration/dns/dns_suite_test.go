package dns_test

import (
	"testing"
	"time"

	"github.com/giantswarm/dns-operator-gcp/tests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const slowTestThreshold = 30 * time.Second

var (
	baseDomain    string
	parentDNSZone string
	gcpProject    string
)

func TestDns(t *testing.T) {
	suiteConfig, reporterConfig := GinkgoConfiguration()
	reporterConfig.SlowSpecThreshold = slowTestThreshold

	RegisterFailHandler(Fail)
	RunSpecs(t, "Dns Suite", suiteConfig, reporterConfig)
}

var _ = BeforeSuite(func() {
	baseDomain = tests.GetEnvOrSkip("CLOUD_DNS_BASE_DOMAIN")
	parentDNSZone = tests.GetEnvOrSkip("CLOUD_DNS_PARENT_ZONE")
	gcpProject = tests.GetEnvOrSkip("GCP_PROJECT_ID")
})
