package dns_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	baseDomain string
	gcpProject string
)

func TestDns(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dns Suite")
}

var _ = BeforeSuite(func() {
	baseDomain = getEnvOrFail("CLOUD_DNS_BASE_DOMAIN")
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
