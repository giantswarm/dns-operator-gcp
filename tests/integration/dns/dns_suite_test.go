package dns_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
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
	baseDomain = getEnvOrFail("CLOUD_DNS_BASE_DOMAIN")
	parentDNSZone = getEnvOrFail("CLOUD_DNS_PARENT_ZONE")
	gcpProject = getEnvOrFail("GCP_PROJECT_ID")
})

func generateGUID(prefix string) string {
	guid := uuid.NewString()

	return fmt.Sprintf("%s-%s", prefix, guid[:13])
}

func getEnvOrFail(env string) string {
	value := os.Getenv(env)
	if value == "" {
		Fail(fmt.Sprintf("%s not exported", env))
	}

	return value
}
